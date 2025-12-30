package services

import (
	"context"

	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/create-go-app/fiber-go-template/pkg/core"
)

type AuthService interface {
	SignUp(ctx context.Context, input *models.SignUp) (*core.ApiResponse, error)
	SignIn(ctx context.Context, input *models.SignIn) (*core.ApiResponse, error)
	SignOut(ctx context.Context, c any) (*core.ApiResponse, error)
}
