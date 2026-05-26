package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type OutboxManager interface {
	SendMessageToOutbox(
		ctx context.Context,
		tx pgx.Tx,
		routingKey string,
		message []byte,
	) error
}

type PgxIface interface {
	Begin(context.Context) (pgx.Tx, error)
	Close()
}

type AccountsService struct {
	Db      PgxIface
	Queries repository.DbQueries
	OutboxManager
}

type PaymentRequest struct {
	FromAccountId   int64 `json:"fromAccountId" binding:"required"`
	ToAccountId     int64 `json:"toAccountId" binding:"required"`
	AmountInPennies int64 `json:"amountInPennies" binding:"required"`
}

const PAYMENT_STARTED_ROUTING_KEY = "account.payment.started"

func (service *AccountsService) StartPayment(ctx context.Context, idempotencyKey pgtype.UUID, senderAccountId int64, receiverAccountId int64, amountInPennies int64) error {
	// TODO: add overdraft logic
	// TODO: add custom error structs for logging
	// TODO: cant send money to self and amount needs to be positive

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

	// idempotency check
	_, err = queries.DuplicatePaymentCheck(ctx, repository.DuplicatePaymentCheckParams{
		AccountID:      senderAccountId,
		IdempotencyKey: idempotencyKey,
	})

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Printf("Unable to run duplicate payment check: %v", err)
		return err
	}

	if err == nil {
		log.Printf("Payment for this idempotency key exists: %d, %s, %v", senderAccountId, idempotencyKey, err)
		return errors.New("Payment for this idempotency key exists")
	}

	err = updateAccountBalance(ctx, queries, senderAccountId, amountInPennies, true, true)

	if err != nil {
		return err
	}

	description := fmt.Sprintf("Transfer to %s", receiverAccountRow.AccountHolderName)
	transactionRow, err := createAccountTransaction(ctx, queries, senderAccountId, description)

	if err != nil {
		return err
	}

	_, err = createTransactionLedgerEntry(
		ctx,
		queries,
		idempotencyKey,
		false,
		true,
		transactionRow.ID,
		senderAccountId,
		receiverAccountId,
		amountInPennies,
	)

	if err != nil {
		return err
	}

	messagePayload, err := getBalanceUpdateMessage(
		idempotencyKey,
		transactionRow.ID,
		senderAccountId,
		receiverAccountId,
		amountInPennies,
	)

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

func updateAccountBalance(
	ctx context.Context,
	queries repository.Querier,
	accountId int64,
	amountInPennies int64,
	isDebit bool,
	checkForSufficientFunds bool,
) error {
	account, err := queries.GetAccountForUpdate(ctx, accountId)

	if err != nil {
		return fmt.Errorf("Failed to account balance (ID %d): %v", accountId, err)
	}

	if checkForSufficientFunds && amountInPennies > account.BalanceInPennies {
		return fmt.Errorf("Account doesn't have enough to make payment: %v", err)
	}

	newBalance := account.BalanceInPennies

	if isDebit {
		newBalance = account.BalanceInPennies - amountInPennies
	} else {
		newBalance = account.BalanceInPennies + amountInPennies
	}

	log.Printf("Updating account balance (ID %d): %d -> %d", accountId, account.BalanceInPennies, newBalance)

	return queries.UpdateBalance(ctx, repository.UpdateBalanceParams{
		ID:               accountId,
		BalanceInPennies: newBalance,
	})
}

func createAccountTransaction(
	ctx context.Context,
	queries repository.Querier,
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
	queries repository.Querier,
	indempotencyKey pgtype.UUID,
	isCompensatingTxn bool,
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
		IdempotencyKey:      indempotencyKey,
		IsCompensatingTxn:   isCompensatingTxn,
		TransactionID:       transactionId,
		AccountID:           accountId,
		OtherPartyAccountID: otherParyAccountId,
		AmountInPennies:     ledgerAmount,
	}
	return queries.CreateTransactionLedgerEntry(ctx, ledgerEntryData)
}

type BalanceUpdateMessage struct {
	IdempotencyKey       pgtype.UUID `json:"idempotency_key"`
	AccountTransactionId int64       `json:"account_transaction_id"`
	SenderAccountId      int64       `json:"sender_account_id"`
	ReceiverAccountId    int64       `json:"receiver_account_id"`
	AmountInPennies      int64       `json:"amount_in_pennies"`
}

func getBalanceUpdateMessage(idempotencyKey pgtype.UUID, accountTransactionId, senderAccountId int64, receiverAccountId int64, amountInPennies int64) ([]byte, error) {
	payload, err := json.Marshal(BalanceUpdateMessage{
		AccountTransactionId: accountTransactionId,
		IdempotencyKey:       idempotencyKey,
		SenderAccountId:      senderAccountId,
		ReceiverAccountId:    receiverAccountId,
		AmountInPennies:      amountInPennies,
	})

	if err != nil {
		return []byte{}, err
	}

	return payload, nil
}

func (service *AccountsService) SettlePayment(
	ctx context.Context,
	idempotencyKey pgtype.UUID,
	accountTransactionId int64,
	senderAccountId int64,
	receiverAccountId int64,
	amountInPennies int64,
) error {
	tx, err := service.Db.Begin(ctx)

	if err != nil {
		return fmt.Errorf("Failed to start SettlePayment tx: %w", err)
	}

	defer tx.Rollback(ctx)

	queries := service.Queries.WithTx(tx)

	_, err = createTransactionLedgerEntry(
		ctx,
		queries,
		idempotencyKey,
		false,
		false,
		accountTransactionId,
		receiverAccountId,
		senderAccountId,
		amountInPennies,
	)

	if err != nil {
		constrainErr := isConstraintError("transactions_ledger_idempotency_key_check", err)

		if constrainErr {
			log.Printf("Payment has already been settled - dropping message")
			return nil
		}

		return fmt.Errorf("Failed to create ledger entry for receiver: %w", err)
	}

	err = queries.UpdateTransaction(ctx, repository.UpdateTransactionParams{
		ID:     accountTransactionId,
		Status: "settled",
	})

	if err != nil {
		return fmt.Errorf("Failed to update transaction status to settled: %w", err)
	}

	err = updateAccountBalance(ctx, queries, receiverAccountId, amountInPennies, false, false)

	if err != nil {
		return fmt.Errorf("Failed to update recievers balance: %w", err)
	}

	tx.Commit(ctx)
	return nil
}

func (service *AccountsService) RejectFraudPayment(
	ctx context.Context,
	indempotencyKey pgtype.UUID,
	accountTransactionId int64,
	senderAccountId int64,
	receiverAccountId int64,
	amountInPennies int64,
) error {
	tx, err := service.Db.Begin(ctx)

	if err != nil {
		return fmt.Errorf("Failed to start RejectFraudPayment tx: %w", err)
	}

	defer tx.Rollback(ctx)

	queries := service.Queries.WithTx(tx)

	_, err = createTransactionLedgerEntry(
		ctx,
		queries,
		indempotencyKey,
		true,
		false,
		accountTransactionId,
		senderAccountId,
		receiverAccountId,
		amountInPennies,
	)

	if err != nil {
		constrainErr := isConstraintError("transactions_ledger_idempotency_key_check", err)

		if constrainErr {
			log.Printf("Payment has already been rejected - dropping message")
			return nil
		}

		return fmt.Errorf("Failed to create ledger entry to return amount back to sender: %w", err)
	}

	err = queries.UpdateTransaction(ctx, repository.UpdateTransactionParams{
		ID:     accountTransactionId,
		Status: "rejected_fraud",
	})

	if err != nil {
		return fmt.Errorf("Failed to update transaction status to settled: %w", err)
	}

	err = updateAccountBalance(ctx, queries, senderAccountId, amountInPennies, false, false)

	if err != nil {
		return fmt.Errorf("Failed to update recievers balance: %w", err)
	}

	tx.Commit(ctx)
	return nil
}

func isConstraintError(constraintName string, err error) bool {
	var pgError *pgconn.PgError

	if errors.As(err, &pgError) {
		if pgError.Code == pgerrcode.UniqueViolation && pgError.ConstraintName == constraintName {
			return true
		}
	}
	return false
}
