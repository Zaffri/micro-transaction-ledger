package router

import (
	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/handlers"
	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/repository"
	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

func GetRoutes(db *pgxpool.Pool, riverClient *river.Client[pgx.Tx]) *gin.Engine {
	router := gin.Default()

	queries := repository.New(db)

	accountsService := service.AccountsService{
		Db:          db,
		Queries:     queries,
		RiverClient: riverClient,
	}
	accountsHandler := handlers.AccountHandler{AccountsService: accountsService}

	router.GET("/accounts/health", handlers.HealthHandler)
	router.GET("/accounts/:id", accountsHandler.GetAccount)
	router.POST("/accounts/payment", accountsHandler.PaymentHandler)

	return router
}
