package models

import "time"

// Transaction đại diện bảng "Transactions"
type Transaction struct {
	TransactionId   string    `gorm:"column:TransactionId;primaryKey;type:varchar(128);not null"`
	WalletId        string    `gorm:"column:WalletId;type:varchar(128);not null"`
	FromAddress     string    `gorm:"column:FromAddress;type:varchar(128);not null"`
	ToAddress       string    `gorm:"column:ToAddress;type:varchar(128);not null"`
	Amount          float64   `gorm:"column:Amount;type:decimal(18,8);not null"`
	TransactionDate time.Time `gorm:"column:TransactionDate;type:timestamptz"`
	Status          string    `gorm:"column:Status;type:varchar(50);not null"`

	// 🔗 Relation
	Wallet Wallet `gorm:"foreignKey:WalletId;references:WalletId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Transaction) TableName() string {
	return "Transactions"
}
