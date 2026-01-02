package repositories

import (
	"context"

	models "github.com/create-go-app/fiber-go-template/app/entities"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id string) (models.Users, error)
	GetUserByEmail(ctx context.Context, email string) (models.Users, error)
	CreateUser(ctx context.Context, u *models.Users) error
	UpdateUser(ctx context.Context, id string, u *models.Users) error
	DeleteUser(ctx context.Context, id string) error
}
