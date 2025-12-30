package repository

import (
	"context"

	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/create-go-app/fiber-go-template/app/interfaces/repositories"
	"github.com/create-go-app/fiber-go-template/platform/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repositories.UserRepository {
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) getDB(ctx context.Context) *gorm.DB {
	if tx := database.GetTx(ctx); tx != nil {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *UserRepositoryImpl) GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	var user models.User
	err := r.getDB(ctx).Where("id = ?", id).First(&user).Error
	return user, err
}

func (r *UserRepositoryImpl) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := r.getDB(ctx).Where("email = ?", email).First(&user).Error
	return user, err
}

func (r *UserRepositoryImpl) CreateUser(ctx context.Context, u *models.User) error {
	return r.getDB(ctx).Create(u).Error
}

func (r *UserRepositoryImpl) UpdateUser(ctx context.Context, id uuid.UUID, u *models.User) error {
	return r.getDB(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"name":        u.Name,
			"email":       u.Email,
			"user_status": u.UserStatus,
			"user_role":   u.UserRole,
			"updated_at":  u.UpdatedAt,
		}).Error
}

func (r *UserRepositoryImpl) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return r.getDB(ctx).Where("id = ?", id).Delete(&models.User{}).Error
}
