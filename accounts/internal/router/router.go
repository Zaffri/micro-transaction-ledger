package router

import (
	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/handlers"
	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/jobs"
	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetRoutes(db *pgxpool.Pool, riverManager jobs.RiverManager, accountsService service.AccountsService) *gin.Engine {
	router := gin.Default()

	accountsHandler := handlers.AccountHandler{AccountsService: accountsService}

	router.GET("/accounts/health", handlers.HealthHandler)
	router.GET("/accounts/:id", accountsHandler.GetAccount)
	router.GET("/accounts/:id/statement", accountsHandler.GetAccountStatement)
	router.POST("/accounts/payment", accountsHandler.PaymentHandler)

	return router
}
