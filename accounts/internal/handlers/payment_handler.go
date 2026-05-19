package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PaymentRequest struct {
	FromAccountId   int `json:"fromAccountId" binding:"required"`
	ToAccountId     int `json:"toAccountId" binding:"required"`
	AmountInPennies int `json:"amountInPennies" binding:"required"`
}

func (handler *AccountHandler) PaymentHandler(ctx *gin.Context) {
	var paymentRequest PaymentRequest
	err := ctx.BindJSON(&paymentRequest)

	if err != nil {
		log.Printf("Invalid payment request body: %v", err)
		ctx.Status(400)
		return
	}

	log.Printf("paymentRequest %v", paymentRequest)

	// Note: no auth/ownership checks for simplicity - obviously wouldn't do this in real project

	// Atomically, check balance for funds, if ok update and create outbox...

	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
