package database

import (
	"context"

	"github.com/create-go-app/fiber-go-template/app/interfaces/repositories"
	"gorm.io/gorm"
)

type GormTransactionManager struct {
	db *gorm.DB
}

func NewGormTransactionManager(db *gorm.DB) repositories.TransactionManager {
	return &GormTransactionManager{db: db}
}

func (tm *GormTransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctxWithTx := WithTx(ctx, tx)
		return fn(ctxWithTx)
	})
}
