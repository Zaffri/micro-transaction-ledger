package jobs

import (
	"context"
	"log"

	"github.com/Zaffri/micro-transaction-ledger/fraud/internal/rabbitmq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

type RiverManager struct {
	RiverClient *river.Client[pgx.Tx]
}

func NewRiverClient(ctx context.Context, db *pgxpool.Pool, client *rabbitmq.RabbitMQClient) (RiverManager, error) {
	workers := river.NewWorkers()
	err := river.AddWorkerSafely(workers, &RabbitRelayWorker{RabbitClient: client})

	riverClient, err := river.NewClient(riverpgxv5.New(db), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 10},
		},
		Workers: workers,
	})

	err = riverClient.Start(ctx)

	if err != nil {
		return RiverManager{}, err
	}

	return RiverManager{RiverClient: riverClient}, nil
}

func (riverManager *RiverManager) SendMessageToOutbox(ctx context.Context, tx pgx.Tx, routingKey string, message []byte) error {
	_, err := riverManager.RiverClient.InsertTx(ctx, tx, RabbitMQPublishArgs{
		RoutingKey: routingKey,
		Payload:    message,
	}, nil)

	if err != nil {
		log.Printf("Failed to update outbox table with %s message: %v", routingKey, err)
		return err
	}

	return nil
}
