package rabbitmq

import (
	"context"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	msgs       <-chan amqp.Delivery
}

const ACCOUNTS_EXCHANGE_NAME = "accounts-service.events"
const FRAUD_EXCHANGE_NAME = "fraud-service.events"
const FRAUD_CHECKS_QUEUE_NAME = "fraud_checks"
const PAYMENT_STARTED_ROUTING_KEY = "account.payment.started"

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
		ACCOUNTS_EXCHANGE_NAME,
		"topic",
		true,  // durability
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // args
	)

	if err != nil {
		log.Fatalf("Unable to redeclare accounts exchange: %v\n", err)
	}

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
		log.Fatalf("Unable to setup fraud exchange: %v\n", err)
	}

	fraudQueue, err := ch.QueueDeclare(
		FRAUD_CHECKS_QUEUE_NAME,
		true,  // durability
		false, // delete once used
		false, // exclusive
		false, // no-wait
		nil,
	)

	if err != nil {
		log.Fatalf("Unable to create fraud checks queue: %v\n", err)
	}

	// TODO: dig into this...
	err = ch.Qos(1, 0, false)

	if err != nil {
		log.Fatalf("Unable to setup Qos for fraud checks queue: %v\n", err)
	}

	err = ch.QueueBind(
		fraudQueue.Name,
		PAYMENT_STARTED_ROUTING_KEY,
		ACCOUNTS_EXCHANGE_NAME,
		false, // no-wait
		nil,   // args
	)

	if err != nil {
		log.Fatalf("Unable to bind fraud checks queue to account exchange: %v\n", err)
	}

	msgs, err := ch.Consume(
		fraudQueue.Name,
		"fraud_consumer", // consumer
		false,            // auto-ack - require manual confirmation
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)

	if err != nil {
		log.Fatalf("Unable to setup fraud consumer: %v\n", err)
	}

	return &RabbitMQClient{
		connection: conn,
		channel:    ch,
		msgs:       msgs,
	}
}

func (client *RabbitMQClient) PublishFraudMessage(ctx context.Context, routingKey string, body []byte) error {
	return client.channel.PublishWithContext(ctx,
		FRAUD_EXCHANGE_NAME,
		routingKey, // e.g. "fraud.payment.passed"
		false,      // Mandatory
		false,      // Immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // write messages to disk
			Body:         body,
		},
	)
}

type BalanceUpdateMessage struct {
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
