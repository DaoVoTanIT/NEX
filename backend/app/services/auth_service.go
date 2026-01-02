package services

import (
	"context"
	"time"

	"github.com/create-go-app/fiber-go-template/pkg/core"
	"github.com/gofiber/fiber/v2"

	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/create-go-app/fiber-go-template/app/interfaces/repositories"
	"github.com/create-go-app/fiber-go-template/app/interfaces/services"
	"github.com/create-go-app/fiber-go-template/pkg/utils"
	"github.com/create-go-app/fiber-go-template/platform/cache"
	"github.com/google/uuid"
)

type AuthServiceImpl struct {
	userRepo     repositories.UserRepository
	cacheService *cache.CacheService
}

func NewAuthService(userRepo repositories.UserRepository, cacheService *cache.CacheService) services.AuthService {
	return &AuthServiceImpl{
		userRepo:     userRepo,
		cacheService: cacheService,
	}
}

func (s *AuthServiceImpl) SignUp(ctx context.Context, input *models.SignUp) (*core.ApiResponse, error) {
	validate := utils.NewValidator()
	if err := validate.Struct(input); err != nil {
		return core.Error(400, "validation error", utils.ValidatorErrors(err), nil), nil
	}

	role, err := utils.VerifyRole(input.UserRole)
	if err != nil {
		return core.Error(400, "invalid role", err.Error(), nil), nil
	}

	user := &models.Users{
		UserId:       uuid.New().String(),
		CreateDate:   time.Now(),
		Email:        input.Email,
		PasswordHash: utils.GeneratePassword(input.Password),
		UserStatus:   1,
		UserRole:     role,
		Name:         input.Name,
	}

	if err := validate.Struct(user); err != nil {
		return core.Error(400, "validation error", utils.ValidatorErrors(err), nil), nil
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return core.Error(500, "create user failed", err.Error(), nil), nil
	}

	user.PasswordHash = ""
	return core.Success(200, "ok", user, nil), nil
}

func (s *AuthServiceImpl) SignIn(ctx context.Context, input *models.SignIn) (*core.ApiResponse, error) {
	validate := utils.NewValidator()
	if err := validate.Struct(input); err != nil {
		return core.Error(400, "validation error", utils.ValidatorErrors(err), nil), nil
	}

	user, err := s.userRepo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return core.Error(404, "user not found", err.Error(), nil), nil
	}

	if !utils.ComparePasswords(user.PasswordHash, input.Password) {
		return core.Error(400, "wrong email or password", nil, nil), nil
	}

	creds, err := utils.GetCredentialsByRole(user.UserRole)
	if err != nil {
		return core.Error(400, "credentials error", err.Error(), nil), nil
	}

	tokens, err := utils.GenerateNewTokens(user.UserId, creds)
	if err != nil {
		return core.Error(500, "token generation error", err.Error(), nil), nil
	}

	if err := s.cacheService.Set(user.UserId, tokens.Refresh, 0); err != nil {
		return core.Error(500, "cache token failed", err.Error(), nil), nil
	}

	return core.Success(200, "ok", fiber.Map{
		"access":  tokens.Access,
		"refresh": tokens.Refresh,
	}, nil), nil
}

func (s *AuthServiceImpl) SignOut(ctx context.Context, c any) (*core.ApiResponse, error) {
	cc := c.(*fiber.Ctx)
	claims, err := utils.ExtractTokenMetadata(cc)
	if err != nil {
		return core.Error(500, "token parse error", err.Error(), nil), nil
	}

	if err := s.cacheService.Delete(claims.UserID.String()); err != nil {
		return core.Error(500, "invalidate token failed", err.Error(), nil), nil
	}
	return core.Success(204, "signed out", nil, nil), nil
}
