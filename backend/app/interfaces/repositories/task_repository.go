package repositories

import (
	"context"

	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/google/uuid"
)

type TaskRepository interface {
	GetTasks(ctx context.Context) ([]models.Task, error)
	GetTask(ctx context.Context, id uuid.UUID) (models.Task, error)
	GetTasksByStatus(ctx context.Context, status string) ([]models.Task, error)
	CreateTask(ctx context.Context, t *models.Task) error
	UpdateTask(ctx context.Context, id uuid.UUID, t *models.Task) error
	DeleteTask(ctx context.Context, id uuid.UUID) error
}
