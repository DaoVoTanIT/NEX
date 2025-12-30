package models

import (
	"time"

	"github.com/google/uuid"
)

type TaskHistory struct {
	ID        uuid.UUID `db:"id" json:"id" validate:"required,uuid"`
	TaskID    uuid.UUID `json:"task_id" db:"task_id,validate:"required,uuid"`
	Action    string    `json:"action" db:"action"`
	OldStatus string    `json:"old_status" db:"old_status"`
	NewStatus string    `json:"new_status" db:"new_status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	CreatedBy uuid.UUID `json:"created_by" db:"created_by" validate:"required,uuid"`
}
