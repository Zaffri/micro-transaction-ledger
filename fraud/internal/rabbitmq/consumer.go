package rabbitmq

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Zaffri/micro-transaction-ledger/fraud/internal/service"
)

func SetupFraudWorker(ctx context.Context, rabbitClient *RabbitMQClient, fraudService *service.FraudService) {
	for message := range rabbitClient.msgs {
		log.Printf("Message recieved: %s", message.RoutingKey)

		var messageBody BalanceUpdateMessage
		err := json.Unmarshal(message.Body, &messageBody)

		if err != nil {
			// TODO: setup deadletter queue - dropping message here for simplicity for now
			log.Printf("Fraud worker failed to handle message: %v", err)
			message.Nack(false, false)
			continue
		}

		fraudCheck := service.FraudCheck{
			AccountTransactionId: messageBody.AccountTransactionId,
			SenderAccountId:      messageBody.SenderAccountId,
			ReceiverAccountId:    messageBody.ReceiverAccountId,
			AmountInPennies:      messageBody.AmountInPennies,
		}

		err = fraudCheck.RunChecks()

		if err != nil {
			// TODO: setup deadletter queue - dropping message here for simplicity for now
			log.Printf("Running fraud checks failed: %v", err)
			message.Nack(false, false)
			continue
		}

		err = fraudService.SaveFraudResult(ctx, fraudCheck)

		if err != nil {
			// TODO: setup deadletter queue - dropping message here for simplicity for now
			log.Printf("Saving fraud checks failed: %v", err)
			message.Nack(false, false)
			continue
		}

		message.Ack(false)
	}
}
