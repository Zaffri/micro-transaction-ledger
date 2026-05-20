package jobs

import (
	"context"

	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/rabbitmq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

func GetRiverClient(ctx context.Context, db *pgxpool.Pool, client *rabbitmq.RabbitMQClient) (*river.Client[pgx.Tx], error) {
	workers := river.NewWorkers()
	err := river.AddWorkerSafely(workers, &RabbitRelayWorker{RabbitClient: client})

	riverClient, err := river.NewClient(riverpgxv5.New(db), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 10},
		},
		Workers: workers,
	})

	err = riverClient.Start(ctx)

	return riverClient, err
}
