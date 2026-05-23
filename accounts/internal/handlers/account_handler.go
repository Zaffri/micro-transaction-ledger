package handlers

import (
	"errors"
	"log"
	"strconv"

	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type AccountHandler struct {
	AccountsService service.AccountsService
}

func (handler *AccountHandler) GetAccount(ctx *gin.Context) {
	id := ctx.Param("id")

	accountId, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		log.Printf("Invalid param for getAccount ID: %v", err)
		ctx.Status(400)
		return
	}

	account, err := handler.AccountsService.Queries.GetAccount(ctx, accountId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("Account not found for ID %d", accountId)
			ctx.Status(404)
			return
		}

		log.Printf("Failed to retrieve account: %v", err)
		ctx.Status(500)
		return
	}

	ctx.JSON(200, account)
}

func (handler *AccountHandler) GetAccountStatement(ctx *gin.Context) {
	id := ctx.Param("id")

	accountId, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		log.Printf("Invalid param for GetAccountStatement ID: %v", err)
		ctx.Status(400)
		return
	}

	// TODO: add limit/offset for pagination
	ledgerEntries, err := handler.AccountsService.Queries.GetTransactionLedgerEntries(ctx, accountId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("Account not found for ID %d", accountId)
			ctx.Status(404)
			return
		}

		log.Printf("Failed to retrieve account: %v", err)
		ctx.Status(500)
		return
	}

	if ledgerEntries == nil {
		ctx.JSON(200, []any{})
		return
	}

	ctx.JSON(200, ledgerEntries)
}
