package services

import (
	"context"
	"time"

	"github.com/create-go-app/fiber-go-template/app/interfaces/repositories"
	"github.com/create-go-app/fiber-go-template/app/interfaces/services"
	"github.com/create-go-app/fiber-go-template/pkg/core"
	"github.com/create-go-app/fiber-go-template/pkg/utils"
	"github.com/create-go-app/fiber-go-template/platform/cache"
	"github.com/gofiber/fiber/v2"
)

type TokenServiceImpl struct {
	userRepo     repositories.UserRepository
	cacheService *cache.CacheService
}

func NewTokenService(userRepo repositories.UserRepository, cacheService *cache.CacheService) services.TokenService {
	return &TokenServiceImpl{
		userRepo:     userRepo,
		cacheService: cacheService,
	}
}

func (s *TokenServiceImpl) Renew(ctx context.Context, c any, refreshToken string) (*core.ApiResponse, error) {
	now := time.Now().Unix()

	expiresRefreshToken, err := utils.ParseRefreshToken(refreshToken)
	if err != nil {
		return core.Error(400, "invalid refresh token", err.Error(), nil), nil
	}

	if now >= expiresRefreshToken {
		return core.Error(401, "unauthorized, your session was ended earlier", nil, nil), nil
	}

	cc := c.(*fiber.Ctx)
	claims, err := utils.ExtractTokenMetadata(cc)
	if err != nil {
		return core.Error(500, "token parse error", err.Error(), nil), nil
	}

	userID := claims.UserID

	user, err := s.userRepo.GetUserByID(ctx, userID.String())
	if err != nil {
		return core.Error(404, "user not found", err.Error(), nil), nil
	}

	creds, err := utils.GetCredentialsByRole(user.UserRole)
	if err != nil {
		return core.Error(400, "credentials error", err.Error(), nil), nil
	}

	tokens, err := utils.GenerateNewTokens(userID.String(), creds)
	if err != nil {
		return core.Error(500, "token generation error", err.Error(), nil), nil
	}

	if err := s.cacheService.Set(userID.String(), tokens.Refresh, 0); err != nil {
		return core.Error(500, "cache token failed", err.Error(), nil), nil
	}

	return core.Success(200, "ok", fiber.Map{
		"access":  tokens.Access,
		"refresh": tokens.Refresh,
	}, nil), nil
}
