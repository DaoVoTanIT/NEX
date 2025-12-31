package services

import (
	"context"

	"github.com/create-go-app/fiber-go-template/pkg/core"
)

type TokenService interface {
	Renew(ctx context.Context, c any, refreshToken string) (*core.ApiResponse, error)
}
