package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/router"
)

func main() {
	router := router.GetRoutes()

	// TODO: set timeouts?
	err := http.ListenAndServe(getServiceAddress(), router)

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
