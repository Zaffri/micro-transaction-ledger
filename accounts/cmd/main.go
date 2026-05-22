package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/jobs"
	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/rabbitmq"
	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/repository"
	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/router"
	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/service"
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

	accountsService := service.AccountsService{
		Db:            db,
		Queries:       repository.New(db),
		OutboxManager: &riverManager,
	}

	// TODO: could setup pool of workers - single for now
	go rabbitmq.SetupPaymentSettleWorker(ctx, rabbitClient, accountsService)
	go rabbitmq.SetupPaymentFraudWorker(ctx, rabbitClient, accountsService)

	router := router.GetRoutes(db, riverManager, accountsService)

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
