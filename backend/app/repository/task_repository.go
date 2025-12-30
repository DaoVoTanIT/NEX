package repository

import (
	"context"
	"time"

	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/create-go-app/fiber-go-template/app/interfaces/repositories"
	"github.com/create-go-app/fiber-go-template/platform/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskRepositoryImpl struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) repositories.TaskRepository {
	return &TaskRepositoryImpl{db: db}
}

func (r *TaskRepositoryImpl) getDB(ctx context.Context) *gorm.DB {
	if tx := database.GetTx(ctx); tx != nil {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *TaskRepositoryImpl) GetTasks(ctx context.Context) ([]models.Task, error) {
	var tasks []models.Task
	err := r.getDB(ctx).
		Preload("Creator").
		Preload("Assignee").
		Order("created_at DESC").
		Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepositoryImpl) GetTask(ctx context.Context, id uuid.UUID) (models.Task, error) {
	var task models.Task
	err := r.getDB(ctx).
		Preload("Creator").
		Preload("Assignee").
		First(&task, "id = ?", id).Error
	return task, err
}

func (r *TaskRepositoryImpl) GetTasksByStatus(ctx context.Context, status string) ([]models.Task, error) {
	var tasks []models.Task
	err := r.getDB(ctx).
		Where("status = ?", status).
		Preload("Creator").
		Preload("Assignee").
		Order("created_at DESC").
		Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepositoryImpl) CreateTask(ctx context.Context, t *models.Task) error {
	now := time.Now()
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	t.UpdatedAt = now
	return r.getDB(ctx).Create(t).Error
}

func (r *TaskRepositoryImpl) UpdateTask(ctx context.Context, id uuid.UUID, t *models.Task) error {
	return r.getDB(ctx).
		Model(&models.Task{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"title":       t.Title,
			"description": t.Description,
			"status":      t.Status,
			"assigned_to": t.AssignedTo,
			"updated_at":  time.Now(),
		}).Error
}

func (r *TaskRepositoryImpl) DeleteTask(ctx context.Context, id uuid.UUID) error {
	return r.getDB(ctx).Delete(&models.Task{}, "id = ?", id).Error
}
