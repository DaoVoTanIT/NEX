package repositories

import (
	"context"

	models "github.com/create-go-app/fiber-go-template/app/entities"
)

type WalletRepository interface {
	Create(ctx context.Context, wallet *models.Wallet) error
	GetById(ctx context.Context, walletId string) (*models.Wallet, error)
	ListAll(ctx context.Context) ([]models.Wallet, error)
}
