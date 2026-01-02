package repository

import (
	"context"

	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/create-go-app/fiber-go-template/app/interfaces/repositories"
	"github.com/create-go-app/fiber-go-template/platform/database"
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

func (r *UserRepositoryImpl) GetUserByID(
	ctx context.Context,
	userId string,
) (models.Users, error) {

	var user models.Users

	err := r.getDB(ctx).
		Where(&models.Users{UserId: userId}).
		First(&user).
		Error

	return user, err
}

func (r *UserRepositoryImpl) GetUserByEmail(
	ctx context.Context,
	email string,
) (models.Users, error) {

	var user models.Users

	err := r.getDB(ctx).
		Where(&models.Users{Email: email}).
		First(&user).
		Error

	return user, err
}

func (r *UserRepositoryImpl) CreateUser(
	ctx context.Context,
	user *models.Users,
) error {
	return r.getDB(ctx).Create(user).Error
}

func (r *UserRepositoryImpl) UpdateUser(
	ctx context.Context,
	userId string,
	user *models.Users,
) error {

	return r.getDB(ctx).
		Model(&models.Users{}).
		Where(&models.Users{UserId: userId}).
		Select(
			"Name",
			"Email",
			"UserStatus",
			"UserRole",
			"UpdateDate",
		).
		Updates(user).
		Error
}

func (r *UserRepositoryImpl) DeleteUser(
	ctx context.Context,
	userId string,
) error {

	return r.getDB(ctx).
		Where(&models.Users{UserId: userId}).
		Delete(&models.Users{}).
		Error
}
