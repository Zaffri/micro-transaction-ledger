package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/repository"
	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type MockDbQueries struct {
	repository.Querier
	mockGetAccountQuery func(ctx context.Context, receiverAccountId int64) (repository.Account, error)
}

func (m *MockDbQueries) WithTx(tx pgx.Tx) *repository.Queries {
	return &repository.Queries{}
}

func (m *MockDbQueries) GetAccount(ctx context.Context, receiverAccountId int64) (repository.Account, error) {
	return m.mockGetAccountQuery(ctx, receiverAccountId)
}

func TestGetAccountInvalidParam(t *testing.T) {
	mockQueries := MockDbQueries{}

	mockQueries.mockGetAccountQuery = func(ctx context.Context, receiverAccountId int64) (repository.Account, error) {
		return repository.Account{}, nil
	}

	mockAccountsHandler := AccountHandler{
		AccountsService: service.AccountsService{
			Queries: &mockQueries,
		},
	}

	writer := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(writer)

	ginCtx.Request, _ = http.NewRequest("GET", "/accounts/1", nil)
	ginCtx.Request.Header.Set("Content-Type", "application/json")
	ginCtx.Params = gin.Params{
		{Key: "id", Value: "notanumber"},
	}

	mockAccountsHandler.GetAccount(ginCtx)
	statusResult := ginCtx.Writer.Status()

	if statusResult != 400 {
		t.Errorf("GetAccount() = %d, expected %d", statusResult, 400)
	}
}

func TestGetAccountNotFound(t *testing.T) {
	mockQueries := MockDbQueries{}

	mockQueries.mockGetAccountQuery = func(ctx context.Context, receiverAccountId int64) (repository.Account, error) {
		return repository.Account{}, pgx.ErrNoRows
	}

	mockAccountsHandler := AccountHandler{
		AccountsService: service.AccountsService{
			Queries: &mockQueries,
		},
	}

	writer := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(writer)

	ginCtx.Request, _ = http.NewRequest("GET", "/accounts/1", nil)
	ginCtx.Request.Header.Set("Content-Type", "application/json")
	ginCtx.Params = gin.Params{
		{Key: "id", Value: "1"},
	}

	mockAccountsHandler.GetAccount(ginCtx)
	statusResult := ginCtx.Writer.Status()

	if statusResult != 404 {
		t.Errorf("GetAccount() = %d, expected %d", statusResult, 404)
	}
}

func TestGetAccountServerError(t *testing.T) {
	mockQueries := MockDbQueries{}

	mockQueries.mockGetAccountQuery = func(ctx context.Context, receiverAccountId int64) (repository.Account, error) {
		return repository.Account{}, errors.New("Error")
	}

	mockAccountsHandler := AccountHandler{
		AccountsService: service.AccountsService{
			Queries: &mockQueries,
		},
	}

	writer := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(writer)

	ginCtx.Request, _ = http.NewRequest("GET", "/accounts/1", nil)
	ginCtx.Request.Header.Set("Content-Type", "application/json")
	ginCtx.Params = gin.Params{
		{Key: "id", Value: "1"},
	}

	mockAccountsHandler.GetAccount(ginCtx)
	statusResult := ginCtx.Writer.Status()

	if statusResult != 500 {
		t.Errorf("GetAccount() = %d, expected %d", statusResult, 500)
	}
}

func TestGetAccountOk(t *testing.T) {
	mockQueries := MockDbQueries{}

	mockQueries.mockGetAccountQuery = func(ctx context.Context, receiverAccountId int64) (repository.Account, error) {
		return repository.Account{}, nil
	}

	mockAccountsHandler := AccountHandler{
		AccountsService: service.AccountsService{
			Queries: &mockQueries,
		},
	}

	writer := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(writer)

	ginCtx.Request, _ = http.NewRequest("GET", "/accounts/1", nil)
	ginCtx.Request.Header.Set("Content-Type", "application/json")
	ginCtx.Params = gin.Params{
		{Key: "id", Value: "1"},
	}

	mockAccountsHandler.GetAccount(ginCtx)
	statusResult := ginCtx.Writer.Status()

	if statusResult != 200 {
		t.Errorf("GetAccount() = %d, expected %d", statusResult, 200)
	}
}
