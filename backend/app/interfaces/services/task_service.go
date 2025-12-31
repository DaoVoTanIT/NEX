package services

import (
	"context"

	"github.com/create-go-app/fiber-go-template/app/dto"
	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/create-go-app/fiber-go-template/pkg/core"
	"github.com/google/uuid"
)

type TaskService interface {
	GetTasks(ctx context.Context) (*core.ApiResponse, error)
	GetTask(ctx context.Context, id string) (*core.ApiResponse, error)
	Create(ctx context.Context, userID uuid.UUID, req *dto.CreateTaskReq) (*core.ApiResponse, error)
	Update(ctx context.Context, userID uuid.UUID, task *models.Task) (*core.ApiResponse, error)
	Delete(ctx context.Context, userID uuid.UUID, id string) (*core.ApiResponse, error)
}
