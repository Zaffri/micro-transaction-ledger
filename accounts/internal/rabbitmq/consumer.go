package rabbitmq

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/service"
)

func SetupPaymentSettleWorker(ctx context.Context, rabbitClient *RabbitMQClient, accountsService service.AccountsService) {
	for message := range rabbitClient.settleMessages {
		log.Printf("Message recieved: %s", message.RoutingKey)
		messageBody, err := extractFraudResultMessage(message.Body)

		if err != nil {
			// TODO: setup deadletter queue - dropping message here for simplicity for now
			log.Printf("Payment settle worker failed to handle message: %v", err)
			message.Nack(false, false)
			continue
		}

		if !messageBody.IdempotencyKey.Valid {
			log.Printf("Payment settle - idempotency key is not valid, cant continue: %v", messageBody.IdempotencyKey)
			message.Nack(false, false)
			continue
		}

		err = accountsService.SettlePayment(
			ctx,
			messageBody.IdempotencyKey,
			messageBody.AccountTransactionId,
			messageBody.SenderAccountId,
			messageBody.ReceiverAccountId,
			messageBody.AmountInPennies,
		)

		if err != nil {
			// TODO: setup deadletter queue - dropping message here for simplicity for now
			log.Printf("Settling payment failed: %v", err)
			message.Nack(false, false)
			continue
		}

		message.Ack(false)
	}
}

func SetupPaymentFraudWorker(ctx context.Context, rabbitClient *RabbitMQClient, accountsService service.AccountsService) {
	for message := range rabbitClient.fraudMessages {
		log.Printf("Message recieved: %s", message.RoutingKey)
		messageBody, err := extractFraudResultMessage(message.Body)

		if err != nil {
			// TODO: setup deadletter queue - dropping message here for simplicity for now
			log.Printf("Payment reject fraud failed to handle message: %v", err)
			message.Nack(false, false)
			continue
		}

		if !messageBody.IdempotencyKey.Valid {
			log.Printf("Payment settle - idempotency key is not valid, cant continue: %v", messageBody.IdempotencyKey)
			message.Nack(false, false)
			continue
		}

		err = accountsService.RejectFraudPayment(
			ctx,
			messageBody.IdempotencyKey,
			messageBody.AccountTransactionId,
			messageBody.SenderAccountId,
			messageBody.ReceiverAccountId,
			messageBody.AmountInPennies,
		)

		if err != nil {
			// TODO: setup deadletter queue - dropping message here for simplicity for now
			log.Printf("Rejecting fraud payment failed: %v", err)
			message.Nack(false, false)
			continue
		}

		message.Ack(false)
	}
}

func extractFraudResultMessage(payload []byte) (FraudResultMessage, error) {
	var messageBody FraudResultMessage
	err := json.Unmarshal(payload, &messageBody)
	return messageBody, err
}
