package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxManager interface {
	SendMessageToOutbox(
		ctx context.Context,
		tx pgx.Tx,
		routingKey string,
		message []byte,
	) error
}

type AccountsService struct {
	Db      *pgxpool.Pool
	Queries *repository.Queries // sqlc
	OutboxManager
}

type PaymentRequest struct {
	FromAccountId   int64 `json:"fromAccountId" binding:"required"`
	ToAccountId     int64 `json:"toAccountId" binding:"required"`
	AmountInPennies int64 `json:"amountInPennies" binding:"required"`
}

const PAYMENT_STARTED_ROUTING_KEY = "account.payment.started"

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

	messagePayload, err := getBalanceUpdateMessage(transactionRow.ID, senderAccountId, receiverAccountId, amountInPennies)

	if err != nil {
		return fmt.Errorf("Failed to prepare balance update message for outbox table: %v", err)
	}

	err = service.OutboxManager.SendMessageToOutbox(ctx, tx, PAYMENT_STARTED_ROUTING_KEY, messagePayload)

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
	accountId int64,
	otherParyAccountId int64,
	amountInPennies int64,
) (int64, error) {
	ledgerAmount := amountInPennies

	if isDebit {
		ledgerAmount = -ledgerAmount
	}

	ledgerEntryData := repository.CreateTransactionLedgerEntryParams{
		TransactionID:       transactionId,
		AccountID:           accountId,
		OtherPartyAccountID: otherParyAccountId,
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

func updateReiversBalance(ctx context.Context, queries *repository.Queries, accountId int64, amountInPennies int64) error {
	account, err := queries.GetAccountForUpdate(ctx, accountId)

	if err != nil {
		return fmt.Errorf("Failed to receive senders balance (ID %d): %v", accountId, err)
	}

	newBalance := account.BalanceInPennies + amountInPennies
	log.Printf("Updating account balance (ID %d): %d -> %d", accountId, account.BalanceInPennies, newBalance)

	return queries.UpdateBalance(ctx, repository.UpdateBalanceParams{
		ID:               accountId,
		BalanceInPennies: newBalance,
	})
}

func (service *AccountsService) SettlePayment(
	ctx context.Context,
	accountTransactionId int64,
	senderAccountId int64,
	receiverAccountId int64,
	amountInPennies int64,
) error {
	// 1. start txn
	tx, err := service.Db.Begin(ctx)

	if err != nil {
		return fmt.Errorf("Failed to start ApplyFraudPass tx: %w", err)
	}

	defer tx.Rollback(ctx)

	// 2. wrap queries with txn
	queries := service.Queries.WithTx(tx)

	// 3. write transaction ledger entry for receiver
	_, err = createTransactionLedgerEntry(
		ctx,
		queries,
		false,
		accountTransactionId,
		receiverAccountId,
		senderAccountId,
		amountInPennies,
	)

	if err != nil {
		return fmt.Errorf("Failed to create ledger entry for receiver: %w", err)
	}

	// 4. update transaction to settled status
	err = queries.UpdateTransaction(ctx, repository.UpdateTransactionParams{
		ID:     accountTransactionId,
		Status: "settled",
	})

	if err != nil {
		return fmt.Errorf("Failed to update transaction status to settled: %w", err)
	}

	// 5. update receivers account balance
	err = updateReiversBalance(ctx, queries, receiverAccountId, amountInPennies)

	if err != nil {
		return fmt.Errorf("Failed to update recievers balance: %w", err)
	}

	// 6. commit
	tx.Commit(ctx)
	return nil
}
