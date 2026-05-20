package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/jobs"
	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/repository"
	"github.com/jackc/pgx/v5"
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

	// early exit - check receiver existence before txn start
	receiverAccountRow, err := service.Queries.GetAccount(ctx, receiverAccountId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("Reciever account (%d) does not exist: %v", receiverAccountId, err)
		}
		return fmt.Errorf("Failed to determine if reciever exists: %v", err)
	}

	tx, err := service.Db.Begin(ctx)

	if err != nil {
		return fmt.Errorf("Failed to start UpdateBalance tx: %w", err)
	}
	defer tx.Rollback(ctx)

	queries := service.Queries.WithTx(tx)

	err = updateSendersBalance(ctx, queries, senderAccountId, amountInPennies)

	if err != nil {
		return err
	}

	description := fmt.Sprintf("Transfer to %s", receiverAccountRow.AccountHolderName)
	transactionRow, err := createAccountTransaction(ctx, queries, senderAccountId, description)

	if err != nil {
		return err
	}

	_, err = createTransactionLedgerEntry(ctx, queries, true, transactionRow.ID, senderAccountId, receiverAccountId, amountInPennies)

	if err != nil {
		return err
	}

	// send message to rabbitmq exchange using outbox pattern
	err = service.sendPaymentStartedMessage(ctx, tx, transactionRow.ID, senderAccountId, receiverAccountId, amountInPennies)

	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return fmt.Errorf("Transaction commit error - failed to update balance: %v", err)
	}

	return nil
}

func updateSendersBalance(ctx context.Context, queries *repository.Queries, senderAccountId int64, amountInPennies int64) error {
	account, err := queries.GetAccountForUpdate(ctx, senderAccountId)

	if err != nil {
		return fmt.Errorf("Failed to receive senders balance (ID %d): %v", senderAccountId, err)
	}

	if amountInPennies > account.BalanceInPennies {
		return fmt.Errorf("Sender doesn't have enough to make payment: %v", err)
	}

	newBalance := account.BalanceInPennies - amountInPennies
	log.Printf("Updating account balance (ID %d): %d -> %d", senderAccountId, account.BalanceInPennies, newBalance)

	return queries.UpdateBalance(ctx, repository.UpdateBalanceParams{
		ID:               senderAccountId,
		BalanceInPennies: newBalance,
	})
}

func createAccountTransaction(
	ctx context.Context,
	queries *repository.Queries,
	senderAccountId int64,
	description string,
) (repository.CreateTransactionRow, error) {
	transactionData := repository.CreateTransactionParams{
		Description: description,
		AccountID:   senderAccountId,
	}
	return queries.CreateTransaction(ctx, transactionData)
}

func createTransactionLedgerEntry(
	ctx context.Context,
	queries *repository.Queries,
	isDebit bool,
	transactionId int64,
	senderAccountId int64,
	receiverAccountId int64,
	amountInPennies int64,
) (int64, error) {
	ledgerAmount := amountInPennies

	if isDebit {
		ledgerAmount = -ledgerAmount
	}

	ledgerEntryData := repository.CreateTransactionLedgerEntryParams{
		TransactionID:       transactionId,
		AccountID:           senderAccountId,
		OtherPartyAccountID: receiverAccountId,
		AmountInPennies:     ledgerAmount,
	}
	return queries.CreateTransactionLedgerEntry(ctx, ledgerEntryData)
}

type BalanceUpdateMessage struct {
	AccountTransactionId int64 `json:"account_transaction_id"`
	SenderAccountId      int64 `json:"sender_account_id"`
	ReceiverAccountId    int64 `json:"receiver_account_id"`
	AmountInPennies      int64 `json:"amount_in_pennies"`
}

func getBalanceUpdateMessage(accountTransactionId, senderAccountId int64, receiverAccountId int64, amountInPennies int64) ([]byte, error) {
	payload, err := json.Marshal(BalanceUpdateMessage{
		AccountTransactionId: accountTransactionId,
		SenderAccountId:      senderAccountId,
		ReceiverAccountId:    receiverAccountId,
		AmountInPennies:      amountInPennies,
	})

	if err != nil {
		return []byte{}, err
	}

	return payload, nil
}

func (service *AccountsService) sendPaymentStartedMessage(
	ctx context.Context,
	tx pgx.Tx,
	accountTransactionId int64,
	senderAccountId int64,
	receiverAccountId int64,
	amountInPennies int64,
) error {
	messagePayload, err := getBalanceUpdateMessage(accountTransactionId, senderAccountId, receiverAccountId, amountInPennies)

	if err != nil {
		return fmt.Errorf("Failed to prepare balance update message for outbox table: %v", err)
	}

	_, err = service.RiverClient.InsertTx(ctx, tx, jobs.RabbitMQPublishArgs{
		RoutingKey: "account.payment.started",
		Payload:    messagePayload,
	}, nil)

	if err != nil {
		log.Printf("Failed to update outbox table with balance update message: %v", err)
		return err
	}

	return nil
}
