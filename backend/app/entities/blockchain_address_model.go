package models

import (
	"time"
)

type BlockchainAddress struct {
	AddressId  string    `gorm:"column:AddressId;primaryKey;type:varchar(128);not null"`
	WalletId   string    `gorm:"column:WalletId;type:varchar(128);not null"`
	Address    string    `gorm:"column:Address;type:varchar(128);not null"`
	CreateDate time.Time `gorm:"column:CreateDate;type:timestamptz"`
	UpdateDate time.Time `gorm:"column:UpdateDate;type:timestamptz"`
}

func (BlockchainAddress) TableName() string {
	return "BlockchainAddresses"
}
