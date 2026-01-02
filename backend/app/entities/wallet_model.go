package models

import "time"

// Wallet đại diện bảng "Wallets"
type Wallet struct {
	WalletId         string    `gorm:"column:WalletId;primaryKey;type:varchar(128);not null"`
	WalletName       string    `gorm:"column:WalletName;type:varchar(256);not null"`
	SecretPhraseHash string    `gorm:"column:SecretPhraseHash;type:text;not null"`
	PassphraseHash   string    `gorm:"column:PassphraseHash;type:text"`
	CreateDate       time.Time `gorm:"column:CreateDate;type:timestamptz"`
	UpdateDate       time.Time `gorm:"column:UpdateDate;type:timestamptz"`

	// 🔗 Relations
	BlockchainAddresses []BlockchainAddress `gorm:"foreignKey:WalletId;references:WalletId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Transactions        []Transaction       `gorm:"foreignKey:WalletId;references:WalletId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Wallet) TableName() string {
	return "Wallets"
}
