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

		var messageBody FraudResultMessage
		err := json.Unmarshal(message.Body, &messageBody)

		if err != nil {
			// TODO: setup deadletter queue - dropping message here for simplicity for now
			log.Printf("Payment settle worker failed to handle message: %v", err)
			message.Nack(false, false)
			continue
		}

		log.Printf("PERFORM PAYMENT SETTLE HERE....")

		err = accountsService.SettlePayment(
			ctx,
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
