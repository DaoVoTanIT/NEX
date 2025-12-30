package services

import (
	"context"

	"github.com/create-go-app/fiber-go-template/app/dto"
	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/create-go-app/fiber-go-template/pkg/core"
)

type AuthService interface {
	SignUp(ctx context.Context, input *models.SignUp) (*core.ApiResponse, error)
	SignIn(ctx context.Context, input *models.SignIn) (*core.ApiResponse, error)
	SignOut(ctx context.Context, c any) (*core.ApiResponse, error)
}

type TokenService interface {
	Renew(ctx context.Context, c any, refreshToken string) (*core.ApiResponse, error)
}

type TaskService interface {
	GetTasks(ctx context.Context) (*core.ApiResponse, error)
	GetTask(ctx context.Context, id string) (*core.ApiResponse, error)
	Create(ctx context.Context, c any, req *dto.CreateTaskReq) (*core.ApiResponse, error)
	Update(ctx context.Context, c any, task *models.Task) (*core.ApiResponse, error)
	Delete(ctx context.Context, c any, id string) (*core.ApiResponse, error)
}
