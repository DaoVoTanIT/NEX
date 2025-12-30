package services

import (
	"context"
	"time"

	"github.com/create-go-app/fiber-go-template/pkg/core"
	"github.com/create-go-app/fiber-go-template/pkg/utils"
	"github.com/create-go-app/fiber-go-template/platform/cache"
	"github.com/create-go-app/fiber-go-template/platform/database"
	"github.com/gofiber/fiber/v2"
)

type DefaultTokenService struct{}

func (s *DefaultTokenService) Renew(ctx context.Context, c any, refreshToken string) (*core.ApiResponse, error) {
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

	db, err := database.OpenDBConnection()
	if err != nil {
		return core.Error(500, "database error", err.Error(), nil), nil
	}

	user, err := db.GetUserByID(userID)
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

	connRedis, err := cache.RedisConnection()
	if err != nil {
		return core.Error(500, "redis error", err.Error(), nil), nil
	}

	if err := connRedis.Set(context.Background(), userID.String(), tokens.Refresh, 0).Err(); err != nil {
		return core.Error(500, "cache token failed", err.Error(), nil), nil
	}

	return core.Success(200, "ok", fiber.Map{
		"access":  tokens.Access,
		"refresh": tokens.Refresh,
	}, nil), nil
}
