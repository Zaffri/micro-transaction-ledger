package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Zaffri/micro-transaction-ledger/fraud/internal/jobs"
	"github.com/Zaffri/micro-transaction-ledger/fraud/internal/rabbitmq"
	"github.com/Zaffri/micro-transaction-ledger/fraud/internal/repository"
	"github.com/Zaffri/micro-transaction-ledger/fraud/internal/router"
	"github.com/Zaffri/micro-transaction-ledger/fraud/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_CONNECTION"))

	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	defer db.Close()
	log.Println("Successfully connected to postgres database.")

	rabbitClient := rabbitmq.GetClient(os.Getenv("RABBITMQ_CONNECTION"))
	defer rabbitClient.Close()
	log.Println("Successfully connected to RabbitMQ")

	riverManager, err := jobs.NewRiverClient(ctx, db, rabbitClient)

	if err != nil {
		log.Fatalf("Failed to start river (for outbox relay): %v\n", err)
	}

	defer riverManager.RiverClient.Stop(ctx)
	log.Println("Successfully started River (outbox relay)")

	// TODO: could setup pool of workers - single for now
	go rabbitmq.SetupFraudWorker(ctx, rabbitClient, &service.FraudService{
		Db:            db,
		Queries:       repository.New(db),
		OutboxManager: &riverManager,
	})

	router := router.GetRoutes(db, nil)

	srv := &http.Server{
		Addr:              getServiceAddress(),
		Handler:           router,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	err = srv.ListenAndServe()

	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func getServiceAddress() string {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8081"
	}

	return fmt.Sprintf(":%s", port)
}
