package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/jobs"
	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/rabbitmq"
	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/router"
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

	riverClient, err := jobs.GetRiverClient(ctx, db, rabbitClient)

	if err != nil {
		log.Fatalf("Failed to start river (for outbox relay): %v\n", err)
	}

	defer riverClient.Stop(ctx)
	log.Println("Successfully started River (outbox relay)")

	router := router.GetRoutes(db, riverClient)

	// TODO: set timeouts?
	err = http.ListenAndServe(getServiceAddress(), router)

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
