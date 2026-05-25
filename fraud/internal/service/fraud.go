package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Zaffri/micro-transaction-ledger/fraud/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const FRAUD_PASSED_ROUTING_KEY = "fraud.payment.passed"
const FRAUD_FAILED_ROUTING_KEY = "fraud.payment.failed"

type OutboxManager interface {
	SendMessageToOutbox(
		ctx context.Context,
		tx pgx.Tx,
		routingKey string,
		message []byte,
	) error
}

type FraudService struct {
	Db      *pgxpool.Pool
	Queries *repository.Queries
	OutboxManager
}

type FraudResult struct {
	Pass      bool
	RiskScore int16
}

type FraudCheck struct {
	IdempotencyKey       pgtype.UUID
	AccountTransactionId int64
	SenderAccountId      int64
	ReceiverAccountId    int64
	AmountInPennies      int64
	Result               FraudResult
}

func (check *FraudCheck) RunChecks() error {
	// Mock fraud value
	fraud := check.AmountInPennies == 60
	riskScore := 0
	pass := false

	if !fraud {
		pass = true
	} else {
		riskScore = 87
	}

	check.Result = FraudResult{
		Pass:      pass,
		RiskScore: int16(riskScore),
	}

	return nil
}

func (service *FraudService) SaveFraudResult(
	ctx context.Context,
	fraudCheck FraudCheck,
) error {
	tx, err := service.Db.Begin(ctx)

	if err != nil {
		return fmt.Errorf("Failed to start ApplyFraudPass tx: %w", err)
	}

	defer tx.Rollback(ctx)

	queries := service.Queries.WithTx(tx)

	_, err = queries.CreateFraudCheck(ctx, repository.CreateFraudCheckParams{
		TransactionID: fraudCheck.AccountTransactionId,
		RiskScore:     fraudCheck.Result.RiskScore,
	})

	if err != nil {
		return fmt.Errorf("Failed to create fraud check: %w", err)
	}

	messagePayload, err := getFraudMessage(fraudCheck)
	routingKey := ""

	if fraudCheck.Result.Pass {
		routingKey = FRAUD_PASSED_ROUTING_KEY
	} else {
		routingKey = FRAUD_FAILED_ROUTING_KEY
	}

	err = service.OutboxManager.SendMessageToOutbox(ctx, tx, routingKey, messagePayload)

	if err != nil {
		return fmt.Errorf("Failed to create fraud result message: %w", err)
	}

	tx.Commit(ctx)
	return nil
}

type FraudResultMessage struct {
	IdempotencyKey       pgtype.UUID `json:"idempotency_key"`
	FraudPass            bool        `json:"fraud_pass"`
	AccountTransactionId int64       `json:"account_transaction_id"`
	SenderAccountId      int64       `json:"sender_account_id"`
	ReceiverAccountId    int64       `json:"receiver_account_id"`
	AmountInPennies      int64       `json:"amount_in_pennies"`
}

func getFraudMessage(fraudCheck FraudCheck) ([]byte, error) {
	payload, err := json.Marshal(FraudResultMessage{
		IdempotencyKey:       fraudCheck.IdempotencyKey,
		FraudPass:            fraudCheck.Result.Pass,
		AccountTransactionId: fraudCheck.AccountTransactionId,
		SenderAccountId:      fraudCheck.SenderAccountId,
		ReceiverAccountId:    fraudCheck.ReceiverAccountId,
		AmountInPennies:      fraudCheck.AmountInPennies,
	})

	if err != nil {
		return []byte{}, err
	}

	return payload, nil
}
