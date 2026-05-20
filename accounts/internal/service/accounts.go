package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/jobs"
	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

type AccountsService struct {
	Db          *pgxpool.Pool
	Queries     *repository.Queries // sqlc
	RiverClient *river.Client[pgx.Tx]
}

type PaymentRequest struct {
	FromAccountId   int64 `json:"fromAccountId" binding:"required"`
	ToAccountId     int64 `json:"toAccountId" binding:"required"`
	AmountInPennies int64 `json:"amountInPennies" binding:"required"`
}

func (service *AccountsService) UpdateBalance(ctx context.Context, senderAccountId int64, receiverAccountId int64, amountInPennies int64) error {
	// TODO: add overdraft logic
	// TODO: add custom error structs for logging
	tx, err := service.Db.Begin(ctx)

	if err != nil {
		return fmt.Errorf("Failed to start UpdateBalance tx: %w", err)
	}
	defer tx.Rollback(ctx)

	queries := service.Queries.WithTx(tx)

	// TODO: check if reciever account exists - could move before txn starts for potential quick early exit and add compensating transaction downstream as fallback?
	_, err = queries.GetAccount(ctx, receiverAccountId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("Reciever account (%d) does not exist: %v", receiverAccountId, err)
		}
		return fmt.Errorf("Failed to determine if reciever exists: %v", err)
	}

	// check account balance
	account, err := queries.GetAccountForUpdate(ctx, senderAccountId)

	if err != nil {
		return fmt.Errorf("Failed to receive senders balance (ID %d): %v", senderAccountId, err)
	}

	if amountInPennies > account.BalanceInPennies {
		return fmt.Errorf("Sender doesn't have enough to make payment: %v", err)
	}

	newBalance := account.BalanceInPennies - amountInPennies
	log.Printf("Updating account balance (ID %d): %d -> %d", senderAccountId, account.BalanceInPennies, newBalance)

	err = queries.UpdateBalance(ctx, repository.UpdateBalanceParams{
		ID:               senderAccountId,
		BalanceInPennies: newBalance,
		// TODO: should manage this via db trigger instead
		UpdatedAt: pgtype.Timestamp{
			Time:  time.Now(),
			Valid: true,
		},
	})

	if err != nil {
		return fmt.Errorf("Failed to update balance: %v", err)
	}

	messagePayload, err := getBalanceUpdateMessage(senderAccountId, receiverAccountId, amountInPennies)

	if err != nil {
		return fmt.Errorf("Failed to prepare balance update message for outbox table: %v", err)
	}

	// Add entry to outbox table
	_, err = service.RiverClient.InsertTx(ctx, tx, jobs.RabbitMQPublishArgs{
		RoutingKey: "account.balance.updated",
		Payload:    messagePayload,
	}, nil)

	if err != nil {
		log.Printf("Failed to update outbox table with balance update message: %v", err)
		return err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return fmt.Errorf("Transaction commit error - failed to update balance: %v", err)
	}

	return nil
}

type BalanceUpdateMessage struct {
	SenderAccountId   int64 `json:"sender_account_id"`
	ReceiverAccountId int64 `json:"receiver_account_id"`
	AmountInPennies   int64 `json:"amount_in_pennies"`
}

func getBalanceUpdateMessage(senderAccountId int64, receiverAccountId int64, amountInPennies int64) ([]byte, error) {
	payload, err := json.Marshal(BalanceUpdateMessage{
		SenderAccountId:   senderAccountId,
		ReceiverAccountId: receiverAccountId,
		AmountInPennies:   amountInPennies,
	})

	if err != nil {
		return []byte{}, err
	}

	return payload, nil
}
