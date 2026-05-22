package rabbitmq

import (
	"context"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	connection     *amqp.Connection
	channel        *amqp.Channel
	settleMessages <-chan amqp.Delivery
}

const ACCOUNTS_EXCHANGE_NAME = "accounts-service.events"
const FRAUD_EXCHANGE_NAME = "fraud-service.events"
const PAYMENT_SETTLE_QUEUE_NAME = "payment_settle"
const FRAUD_PASSED_ROUTING_KEY = "fraud.payment.passed"
const FRAUD_FAILED_ROUTING_KEY = "fraud.payment.failed"

// TODO: return error and handle in main
func GetClient(connectionString string) *RabbitMQClient {
	// TODO: add try mechanism?
	conn, err := amqp.Dial(connectionString)

	if err != nil {
		log.Fatalf("Unable to connect to RabbitMQ: %v\n", err)
	}

	ch, err := conn.Channel()

	if err != nil {
		log.Fatalf("Unable to create channel: %v\n", err)
	}

	// Redeclare exchange on this side to ensure its setup before binding - prevent startup order problem
	err = ch.ExchangeDeclare(
		FRAUD_EXCHANGE_NAME,
		"topic",
		true,  // durability
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // args
	)

	if err != nil {
		log.Fatalf("Unable to redclare fraud exchange: %v\n", err)
	}

	err = ch.ExchangeDeclare(
		ACCOUNTS_EXCHANGE_NAME,
		"topic",
		true,  // durability
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // args
	)

	if err != nil {
		log.Fatalf("Unable to setup accounts exchange: %v\n", err)
	}

	settleQueue, err := ch.QueueDeclare(
		PAYMENT_SETTLE_QUEUE_NAME,
		true,  // durability
		false, // delete once used
		false, // exclusive
		false, // no-wait
		nil,
	)

	if err != nil {
		log.Fatalf("Unable to create payment settle queue: %v\n", err)
	}

	err = ch.Qos(1, 0, false)

	if err != nil {
		log.Fatalf("Unable to setup Qos for payment settle queue: %v\n", err)
	}

	err = ch.QueueBind(
		settleQueue.Name,
		FRAUD_PASSED_ROUTING_KEY,
		FRAUD_EXCHANGE_NAME,
		false, // no-wait
		nil,   // args
	)

	if err != nil {
		log.Fatalf("Unable to bind payment settle queue to account exchange: %v\n", err)
	}

	settleMsgs, err := ch.Consume(
		settleQueue.Name,
		"payment_settle_consumer", // consumer
		false,                     // auto-ack - require manual confirmation
		false,                     // exclusive
		false,                     // no-local
		false,                     // no-wait
		nil,                       // args
	)

	return &RabbitMQClient{
		connection:     conn,
		channel:        ch,
		settleMessages: settleMsgs,
	}
}

func (client *RabbitMQClient) PublishPayment(ctx context.Context, routingKey string, body []byte) error {
	return client.channel.PublishWithContext(ctx,
		ACCOUNTS_EXCHANGE_NAME,
		routingKey, // e.g. "accounts.payment.started"
		false,      // Mandatory
		false,      // Immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // write messages to disk
			Body:         body,
		},
	)
}

type FraudResultMessage struct {
	FraudPass            bool  `json:"fraud_pass"`
	AccountTransactionId int64 `json:"account_transaction_id"`
	SenderAccountId      int64 `json:"sender_account_id"`
	ReceiverAccountId    int64 `json:"receiver_account_id"`
	AmountInPennies      int64 `json:"amount_in_pennies"`
}

func (client *RabbitMQClient) Close() {
	if client.channel != nil {
		client.channel.Close()
	}
	if client.connection != nil {
		client.connection.Close()
	}
}
