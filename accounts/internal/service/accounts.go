package service

import (
	"context"
	"log"

	"github.com/Zaffri/micro-transaction-ledger/accounts/internal/repository"
)

type AccountsService struct {
	Repo repository.Querier
}

func (service *AccountsService) UpdateBalance(ctx context.Context, args repository.UpdateBalanceParams) error {
	// TODO: business checks

	err := service.Repo.UpdateBalance(ctx, args)

	if err != nil {
		log.Printf("Failed to update balance: %v", err)
		return err
	}

	return nil
}
