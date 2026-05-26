package service

import (
	"context"
	"strings"
	"testing"

	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type MockDbQueries struct {
	repository.Querier
	repository.DBTX     // required for WithTx
	mockGetAccountQuery func(ctx context.Context, receiverAccountId int64) (repository.Account, error)
}

func (m *MockDbQueries) WithTx(tx pgx.Tx) *repository.Queries {
	return repository.New(m.DBTX)
}

func (m *MockDbQueries) GetAccount(ctx context.Context, receiverAccountId int64) (repository.Account, error) {
	return m.mockGetAccountQuery(ctx, receiverAccountId)
}

func (m *MockDbQueries) DuplicatePaymentCheck(ctx context.Context, arg repository.DuplicatePaymentCheckParams) (int64, error) {
	return 1, nil
}

func TestStartPaymentNoReceiver(t *testing.T) {
	mockDbQueries := MockDbQueries{}

	mockDbQueries.mockGetAccountQuery = func(ctx context.Context, receiverAccountId int64) (repository.Account, error) {
		return repository.Account{}, pgx.ErrNoRows
	}

	accountService := AccountsService{
		Queries: &mockDbQueries,
	}

	var idempotencyKey pgtype.UUID
	idempotencyKey.Scan("3b886308-9312-4178-a7fe-b77486732b77")

	ctx := context.Background()
	senderAccountId := int64(1)
	receiverAccountId := int64(2)
	amountInPennies := int64(100)

	err := accountService.StartPayment(ctx, idempotencyKey, senderAccountId, receiverAccountId, amountInPennies)
	expectedError := "Reciever account (2) does not exist"

	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("StartPayment() = %s, expected to contain %s", err.Error(), expectedError)
	}
}

func TestStartPaymentUnableToCheckForDuplicates(t *testing.T) {}

func TestStartPaymentDuplicate(t *testing.T) {}

func TestStartPaymentSuccess(t *testing.T) {}
