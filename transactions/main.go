package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/transactions/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"service": "Transactions", "status": "ok"})
	})

	port := os.Getenv("PORT")

	if port == "" {
		port = "8081"
	}

	serverAddress := fmt.Sprintf(":%s", port)

	err := http.ListenAndServe(serverAddress, router)

	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
