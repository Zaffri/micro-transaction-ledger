package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Zaffri/micro-transaction-ledger/fraud/internal/rabbitmq"
	"github.com/riverqueue/river"
)

type RabbitMQPublishArgs struct {
	RoutingKey string          `json:"routing_key"`
	Payload    json.RawMessage `json:"payload"`
}

func (RabbitMQPublishArgs) Kind() string { return "rabbitmq.event.publish" }

type RabbitRelayWorker struct {
	river.WorkerDefaults[RabbitMQPublishArgs]
	RabbitClient *rabbitmq.RabbitMQClient
}

// TODO: check if i should use Publisher Confirms?
func (worker *RabbitRelayWorker) Work(ctx context.Context, job *river.Job[RabbitMQPublishArgs]) error {
	// publish message to RabbitMQ (outbox)
	err := worker.RabbitClient.PublishFraudMessage(ctx, job.Args.RoutingKey, job.Args.Payload)

	if err != nil {
		return fmt.Errorf("Unable to relay event to rabbitmq: %w", err)
	}
	log.Printf("Successfully PublishFraudMessage message for: %s", job.Args.RoutingKey)
	return nil
}
