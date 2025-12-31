package repository

import (
	"context"

	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/create-go-app/fiber-go-template/app/interfaces/repositories"
	"github.com/create-go-app/fiber-go-template/platform/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskHistoryRepositoryImpl struct {
	db *gorm.DB
}

func NewTaskHistoryRepository(db *gorm.DB) repositories.TaskHistoryRepository {
	return &TaskHistoryRepositoryImpl{db: db}
}

func (r *TaskHistoryRepositoryImpl) getDB(ctx context.Context) *gorm.DB {
	if tx := database.GetTx(ctx); tx != nil {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *TaskHistoryRepositoryImpl) GetTaskHistories(ctx context.Context) ([]models.TaskHistory, error) {
	var histories []models.TaskHistory
	err := r.getDB(ctx).Order("created_at DESC").Find(&histories).Error
	return histories, err
}

func (r *TaskHistoryRepositoryImpl) GetTaskHistoriesByTaskID(ctx context.Context, taskID uuid.UUID) ([]models.TaskHistory, error) {
	var histories []models.TaskHistory
	err := r.getDB(ctx).Where("task_id = ?", taskID).Order("created_at DESC").Find(&histories).Error
	return histories, err
}

func (r *TaskHistoryRepositoryImpl) GetTaskHistory(ctx context.Context, id uuid.UUID) (models.TaskHistory, error) {
	var history models.TaskHistory
	err := r.getDB(ctx).Where("id = ?", id).First(&history).Error
	return history, err
}

func (r *TaskHistoryRepositoryImpl) CreateTaskHistory(ctx context.Context, h *models.TaskHistory) error {
	return r.getDB(ctx).Create(h).Error
}
