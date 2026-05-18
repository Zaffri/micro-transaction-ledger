package router

import (
	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/handlers"
	"github.com/gin-gonic/gin"
)

func GetRoutes() *gin.Engine {
	router := gin.Default()

	router.GET("/accounts/health", handlers.HealthHandler)
	router.POST("/accounts/payment", handlers.PaymentHandler)

	return router
}
