package router

import (
	"github.com/Zaffri/micro-transaction-ledger/fraud/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

func GetRoutes(db *pgxpool.Pool, riverClient *river.Client[pgx.Tx]) *gin.Engine {
	router := gin.Default()

	router.GET("/fraud/health", handlers.HealthHandler)

	return router
}
