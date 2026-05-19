package rabbitmq

import (
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

func GetClient(connectionString string) *RabbitMQClient {
	// TODO: add try mechanism?
	conn, err := amqp.Dial(connectionString)

	if err != nil {
		log.Printf("Unable to connect to RabbitMQ: %v\n", err)
		os.Exit(1)
	}

	defer conn.Close()

	ch, err := conn.Channel()

	if err != nil {
		log.Printf("Unable to create channel: %v\n", err)
		os.Exit(1)
	}

	defer ch.Close()

	err = ch.ExchangeDeclare(
		"acounts-service.events",
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

func (client *RabbitMQClient) PublishPayment() {
	// TODO: publish message to payment queue
}

func (client *RabbitMQClient) Close() {
	if client.channel != nil {
		client.channel.Close()
	}
	if client.connection != nil {
		client.connection.Close()
	}
}
