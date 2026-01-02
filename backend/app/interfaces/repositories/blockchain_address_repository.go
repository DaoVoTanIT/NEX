package repositories

import (
	"context"

	models "github.com/create-go-app/fiber-go-template/app/entities"
)

type BlockchainAddressRepository interface {
	Create(ctx context.Context, addr *models.BlockchainAddress) error
}
