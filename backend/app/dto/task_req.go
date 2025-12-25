package dto

import "github.com/google/uuid"

type CreateTaskReq struct {
	Title       string    `json:"title" validate:"required,lte=255"`
	Description string    `json:"description"`
	AssignedTo  uuid.UUID `json:"assigned_to" validate:"uuid4"`
}
