package queries

import (
	"time"

	entities "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskQueries struct {
	*gorm.DB
}

// =========================
// GET LIST TASKS
// =========================
func (q *TaskQueries) GetTasks() ([]entities.Task, error) {
	var tasks []entities.Task

	err := q.
		Preload("Creator").
		Preload("Assignee").
		Order("created_at DESC").
		Find(&tasks).Error

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// =========================
// GET TASK BY ID
// =========================
func (q *TaskQueries) GetTask(id uuid.UUID) (entities.Task, error) {
	var task entities.Task

	err := q.
		Preload("Creator").
		Preload("Assignee").
		First(&task, "id = ?", id).Error

	if err != nil {
		return entities.Task{}, err
	}

	return task, nil
}

// =========================
// GET TASKS BY STATUS
// =========================
func (q *TaskQueries) GetTasksByStatus(status string) ([]entities.Task, error) {
	var tasks []entities.Task

	err := q.
		Where("status = ?", status).
		Preload("Creator").
		Preload("Assignee").
		Order("created_at DESC").
		Find(&tasks).Error

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// =========================
// CREATE TASK
// =========================
func (q *TaskQueries) CreateTask(t *entities.Task) error {
	now := time.Now()
	t.ID = uuid.New()
	t.CreatedAt = now
	t.UpdatedAt = now

	return q.Create(t).Error
}

// =========================
// UPDATE TASK
// =========================
func (q *TaskQueries) UpdateTask(id uuid.UUID, t *entities.Task) error {
	return q.
		Model(&entities.Task{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"title":       t.Title,
			"description": t.Description,
			"status":      t.Status,
			"assigned_to": t.AssignedTo,
			"updated_at":  time.Now(),
		}).Error
}

// =========================
// DELETE TASK
// =========================
func (q *TaskQueries) DeleteTask(id uuid.UUID) error {
	return q.Delete(&entities.Task{}, "id = ?", id).Error
}
