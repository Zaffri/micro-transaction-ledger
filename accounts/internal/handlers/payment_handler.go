package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PaymentRequest struct {
	SenderAccountId   int64 `json:"senderAccountId" binding:"required"`
	ReceiverAccountId int64 `json:"receiverAccountId" binding:"required"`
	AmountInPennies   int64 `json:"amountInPennies" binding:"required"`
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

	err = handler.AccountsService.StartPayment(
		ctx,
		paymentRequest.SenderAccountId,
		paymentRequest.ReceiverAccountId,
		paymentRequest.AmountInPennies,
	)

	if err != nil {
		// TODO: handle different scenarios w/ status codes
		log.Printf("Payment failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
