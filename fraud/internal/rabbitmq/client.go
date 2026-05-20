package rabbitmq

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	msgs       <-chan amqp.Delivery
}

const ACCOUNTS_EXCHANGE_NAME = "acounts-service.events"
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
		log.Fatalf("Unable to setup pub/sub exchange: %v\n", err)
	}

	queue, err := ch.QueueDeclare(
		FRAUD_CHECKS_QUEUE_NAME,
		true,  // durability
		false, // delete once used
		false, // exclusive
		false, // no-wait
		nil,
	)

	if err != nil {
		log.Fatalf("Unable to fraud checks queue: %v\n", err)
	}

	// TODO: dig into this...
	err = ch.Qos(1, 0, false)

	if err != nil {
		log.Fatalf("Unable to setup Qos for fraud checks queue: %v\n", err)
	}

	err = ch.QueueBind(
		queue.Name,
		PAYMENT_STARTED_ROUTING_KEY,
		ACCOUNTS_EXCHANGE_NAME,
		false, // no-wait
		nil,   // args
	)

	if err != nil {
		log.Fatalf("Unable to bind fraud checks queue to account exchange: %v\n", err)
	}

	msgs, err := ch.Consume(
		queue.Name,
		"fraud_consumer", // consumer
		false,            // auto-ack - require manual confirmation
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)

	return &RabbitMQClient{
		connection: conn,
		channel:    ch,
		msgs:       msgs,
	}
}

// TODO: outbox pattern for Fraud messages
// func (client *RabbitMQClient) PublishFraudPass(ctx context.Context, routingKey string, body []byte) error {}
// func (client *RabbitMQClient) PublishFraudFail(ctx context.Context, routingKey string, body []byte) error {}

type BalanceUpdateMessage struct {
	AccountTransactionId int64 `json:"account_transaction_id"`
	SenderAccountId      int64 `json:"sender_account_id"`
	ReceiverAccountId    int64 `json:"receiver_account_id"`
	AmountInPennies      int64 `json:"amount_in_pennies"`
}

func SetupFraudWorker(rabbitClient *RabbitMQClient) {
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

		log.Printf("Message processed succesfully: %v", messageBody)
		message.Ack(false)
	}
}

func (client *RabbitMQClient) Close() {
	if client.channel != nil {
		client.channel.Close()
	}
	if client.connection != nil {
		client.connection.Close()
	}
}
