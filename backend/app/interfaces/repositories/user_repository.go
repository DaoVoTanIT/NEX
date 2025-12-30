package repositories

import (
	"context"

	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	CreateUser(ctx context.Context, u *models.User) error
	UpdateUser(ctx context.Context, id uuid.UUID, u *models.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
}
