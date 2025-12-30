package queries

import (
	"time"

	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserQueries struct {
	*gorm.DB
}

// =========================
// GET USER BY ID
// =========================
func (q *UserQueries) GetUserByID(id uuid.UUID) (models.User, error) {
	var user models.User

	err := q.
		Where("id = ?", id).
		First(&user).Error

	if err != nil {
		return user, err
	}

	return user, nil
}

// =========================
// GET USER BY EMAIL
// =========================
func (q *UserQueries) GetUserByEmail(email string) (models.User, error) {
	var user models.User

	err := q.
		Where("email = ?", email).
		First(&user).Error

	if err != nil {
		return user, err
	}

	return user, nil
}

// =========================
// CREATE USER
// =========================
func (q *UserQueries) CreateUser(u *models.User) error {
	return q.Create(u).Error
}

// =========================
// UPDATE USER
// =========================
func (q *UserQueries) UpdateUser(id uuid.UUID, u *models.User) error {
	return q.
		Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"name":        u.Name,
			"email":       u.Email,
			"user_status": u.UserStatus,
			"user_role":   u.UserRole,
			"updated_at":  time.Now(),
		}).Error
}

// =========================
// DELETE USER
// =========================
func (q *UserQueries) DeleteUser(id uuid.UUID) error {
	return q.Where("id = ?", id).Delete(&models.User{}).Error
}
