package repository

import (
	"context"

	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/create-go-app/fiber-go-template/app/interfaces/repositories"
	"github.com/create-go-app/fiber-go-template/platform/database"
	"gorm.io/gorm"
)

type WalletRepositoryImpl struct {
	db *gorm.DB
}

// ListAll implements [repositories.WalletRepository].
func (r *WalletRepositoryImpl) ListAll(
	ctx context.Context,
) ([]models.Wallet, error) {

	var wallets []models.Wallet

	err := r.getDB(ctx).
		Preload("BlockchainAddresses").
		Find(&wallets).
		Error

	return wallets, err
}

// GetById implements [repositories.WalletRepository].
func (r *WalletRepositoryImpl) GetById(
	ctx context.Context,
	walletId string,
) (*models.Wallet, error) {

	var wallet models.Wallet

	err := r.db.
		WithContext(ctx).
		Table("Wallets").
		Preload("BlockchainAddresses").
		Where("WalletId = ?", walletId).
		First(&wallet).
		Error

	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

func NewWalletRepository(db *gorm.DB) repositories.WalletRepository {
	return &WalletRepositoryImpl{db}
}
func (r *WalletRepositoryImpl) getDB(ctx context.Context) *gorm.DB {
	if tx := database.GetTx(ctx); tx != nil {
		return tx
	}
	return r.db.WithContext(ctx)
}
func (r *WalletRepositoryImpl) Create(
	ctx context.Context,
	w *models.Wallet,
) error {
	return r.getDB(ctx).Create(w).Error
}
