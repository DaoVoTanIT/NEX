package models

import (
	"time"

	"github.com/google/uuid"
)

type TaskHistory struct {
	ID        uuid.UUID `db:"id" json:"id" validate:"required,uuid"`
	TaskID    uuid.UUID `json:"task_id" db:"task_id,validate:"required,uuid"`
	Action    string    `json:"action" db:"action"`
	OldValue  string    `json:"old_value" db:"old_value"`
	NewValue  string    `json:"new_value" db:"new_value"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	CreatedBy uuid.UUID `json:"created_by" db:"created_by" validate:"required,uuid"`
}
