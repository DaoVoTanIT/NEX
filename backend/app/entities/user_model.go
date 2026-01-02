package models

import "time"

// Users đại diện bảng "Users"
type Users struct {
	UserId        string    `gorm:"column:UserId;primaryKey;type:varchar(128);not null"`
	Name          string    `gorm:"column:Name;type:varchar(128);not null"`
	Email         string    `gorm:"column:Email;type:varchar(128);not null"`
	PasswordHash  string    `gorm:"column:PasswordHash;type:text"`
	CreateDate    time.Time `gorm:"column:CreateDate;type:timestamptz"`
	UpdateDate    time.Time `gorm:"column:UpdateDate;type:timestamptz"`
	CreatedUserId string    `gorm:"column:CreatedUserId;type:varchar(128)"`
	UpdatedUserId string    `gorm:"column:UpdatedUserId;type:varchar(128)"`
	UserStatus    int       `gorm:"column:UserStatus;type:int"`
	UserRole      string    `gorm:"column:UserRole;type:varchar(128);not null"`

	Wallets []Wallet `gorm:"foreignKey:UserId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Users) TableName() string {
	return "Users"
}
