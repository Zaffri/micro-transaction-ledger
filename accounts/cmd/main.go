package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/rabbitmq"
	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/router"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	db, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_CONNECTION"))

	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	defer db.Close()
	log.Println("Successfully connected to postgres database.")

	rabbitClient := rabbitmq.GetClient(os.Getenv("RABBITMQ_CONNECTION"))
	defer rabbitClient.Close()

	log.Println("Successfully connected to RabbitMQ")

	router := router.GetRoutes(db)

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
