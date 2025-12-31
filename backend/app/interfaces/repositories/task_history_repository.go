package repositories

import (
	"context"

	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/google/uuid"
)

type TaskHistoryRepository interface {
	GetTaskHistories(ctx context.Context) ([]models.TaskHistory, error)
	GetTaskHistoriesByTaskID(ctx context.Context, taskID uuid.UUID) ([]models.TaskHistory, error)
	GetTaskHistory(ctx context.Context, id uuid.UUID) (models.TaskHistory, error)
	CreateTaskHistory(ctx context.Context, h *models.TaskHistory) error
}
