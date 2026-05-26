package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/service"
	"github.com/gin-gonic/gin"
)

func TestPaymentInvalidBody(t *testing.T) {
	mockQueries := MockDbQueries{}

	mockAccountsHandler := AccountHandler{
		AccountsService: service.AccountsService{
			Queries: &mockQueries,
		},
	}

	writer := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(writer)

	jsonPayload := `{invalid...}`

	ginCtx.Request, _ = http.NewRequest("POST", "/accounts/payment", strings.NewReader(jsonPayload))
	ginCtx.Request.Header.Set("Content-Type", "application/json")

	mockAccountsHandler.PaymentHandler(ginCtx)
	statusResult := ginCtx.Writer.Status()

	if statusResult != 400 {
		t.Errorf("PaymentHandler() = %d, expected %d", statusResult, 400)
	}
}

func TestPaymentMissingIdempotencyKey(t *testing.T) {
	mockQueries := MockDbQueries{}

	mockAccountsHandler := AccountHandler{
		AccountsService: service.AccountsService{
			Queries: &mockQueries,
		},
	}

	writer := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(writer)

	jsonPayload := `{"senderAccountId": 1, "receiverAccountId": 2, "amountInPennies": 1000}`

	ginCtx.Request, _ = http.NewRequest("POST", "/accounts/payment", strings.NewReader(jsonPayload))
	ginCtx.Request.Header.Set("Content-Type", "application/json")

	mockAccountsHandler.PaymentHandler(ginCtx)
	statusResult := ginCtx.Writer.Status()
	responseBody := writer.Body.String()
	expectedError := "Idempotency-Key must be provided"

	if statusResult != 400 || !strings.Contains(responseBody, expectedError) {
		t.Errorf("PaymentHandler() = %d and %s, expected %d and %s", statusResult, responseBody, 400, expectedError)
	}
}

func TestPaymentInvalidIdempotencyKey(t *testing.T) {
	mockQueries := MockDbQueries{}

	mockAccountsHandler := AccountHandler{
		AccountsService: service.AccountsService{
			Queries: &mockQueries,
		},
	}

	writer := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(writer)

	jsonPayload := `{"senderAccountId": 1, "receiverAccountId": 2, "amountInPennies": 1000}`

	ginCtx.Request, _ = http.NewRequest("POST", "/accounts/payment", strings.NewReader(jsonPayload))
	ginCtx.Request.Header.Set("Content-Type", "application/json")
	ginCtx.Request.Header.Set("Idempotency-Key", "not-a-uuid")

	mockAccountsHandler.PaymentHandler(ginCtx)
	statusResult := ginCtx.Writer.Status()
	responseBody := writer.Body.String()
	expectedError := "Idempotency-Key is invalid UUID"

	if statusResult != 400 || !strings.Contains(responseBody, expectedError) {
		t.Errorf("PaymentHandler() = %d and %s, expected %d and %s", statusResult, responseBody, 400, expectedError)
	}
}
