package repository

import (
	"context"

	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/create-go-app/fiber-go-template/app/interfaces/repositories"
	"github.com/create-go-app/fiber-go-template/platform/database"
	"gorm.io/gorm"
)

type BlockchainAddressRepositoryImpl struct {
	db *gorm.DB
}

func (r *BlockchainAddressRepositoryImpl) getDB(ctx context.Context) *gorm.DB {
	if tx := database.GetTx(ctx); tx != nil {
		return tx
	}
	return r.db.WithContext(ctx)
}
func NewBlockchainAddressRepository(db *gorm.DB) repositories.BlockchainAddressRepository {
	return &BlockchainAddressRepositoryImpl{db: db}
}

func (r *BlockchainAddressRepositoryImpl) Create(
	ctx context.Context,
	addr *models.BlockchainAddress,
) error {

	return r.getDB(ctx).Create(addr).Error
}
