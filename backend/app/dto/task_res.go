package dto

import (
	"time"

	"github.com/google/uuid"
)

type TaskRes struct {
	ID          uuid.UUID `gorm:"column:id" json:"id"`
	Title       string    `gorm:"column:title" json:"title"`
	Description string    `gorm:"column:description" json:"description"`
	Status      string    `gorm:"column:status" json:"status"`

	CreatedBy    uuid.UUID `gorm:"column:created_by" json:"created_by"`
	CreateByName string    `gorm:"column:create_by_name" json:"create_by_name"`

	AssignedTo     uuid.UUID `gorm:"column:assigned_to" json:"assigned_to"`
	AssignedToName string    `gorm:"column:assigned_to_name" json:"assigned_to_name"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}
