package rabbitmq

import (
	"context"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

const EXCHANGE_NAME = "accounts-service.events"

// TODO: return error and handle in main
func GetClient(connectionString string) *RabbitMQClient {
	// TODO: add try mechanism?
	conn, err := amqp.Dial(connectionString)

	if err != nil {
		log.Printf("Unable to connect to RabbitMQ: %v\n", err)
		os.Exit(1)
	}

	ch, err := conn.Channel()

	if err != nil {
		log.Printf("Unable to create channel: %v\n", err)
		os.Exit(1)
	}

	err = ch.ExchangeDeclare(
		EXCHANGE_NAME,
		"topic",
		true,  // durability
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // args
	)

	if err != nil {
		log.Printf("Unable to setup pub/sub exchange: %v\n", err)
		os.Exit(1)
	}

	return &RabbitMQClient{
		connection: conn,
		channel:    ch,
	}
}

func (client *RabbitMQClient) PublishPayment(ctx context.Context, routingKey string, body []byte) error {
	return client.channel.PublishWithContext(ctx,
		EXCHANGE_NAME,
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

func (client *RabbitMQClient) Close() {
	if client.channel != nil {
		client.channel.Close()
	}
	if client.connection != nil {
		client.connection.Close()
	}
}
