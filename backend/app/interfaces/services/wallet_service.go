package services

import (
	"context"

	"github.com/create-go-app/fiber-go-template/app/dto"
	"github.com/create-go-app/fiber-go-template/pkg/core"
)

type WalletService interface {
	CreateWallet(ctx context.Context, req *dto.CreateWalletReq) (*dto.CreateWalletRes, error)
	RestoreWallet(ctx context.Context, req *dto.RestoreWalletReq) (*core.ApiResponse, error)
}
