package queries

import (
	"github.com/create-go-app/fiber-go-template/app/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskQueries struct for queries from Task model using GORM.
type TaskQueries struct {
	*gorm.DB
}

// GetTasks method for getting all tasks.
func (q *TaskQueries) GetTasks() ([]models.Task, error) {
	// Define tasks variable.
	var tasks []models.Task

	// Use GORM to find all tasks
	err := q.Find(&tasks).Error
	if err != nil {
		// Return empty slice and error.
		return tasks, err
	}

	// Return query result.
	return tasks, nil
}

// GetTask method for getting one task by given ID.
func (q *TaskQueries) GetTask(id uuid.UUID) (models.Task, error) {
	// Define task variable.
	var task models.Task

	// Use GORM to find task by ID
	err := q.Where("id = ?", id).First(&task).Error
	if err != nil {
		// Return empty object and error.
		return task, err
	}

	// Return query result.
	return task, nil
}

// GetTasksByStatus method for getting all tasks by status.
func (q *TaskQueries) GetTasksByStatus(status string) ([]models.Task, error) {
	// Define tasks variable.
	var tasks []models.Task

	// Use GORM to find tasks by status
	err := q.Where("status = ?", status).Find(tasks).Error
	if err != nil {
		// Return empty slice and error.
		return tasks, err
	}

	// Return query result.
	return tasks, nil
}

// CreateTask method for creating task by given Task object.
func (q *TaskQueries) CreateTask(t *models.Task) error {
	// Use GORM to create task
	err := q.Create(t).Error
	return err
}

// UpdateTask method for updating task by given Task object.
func (q *TaskQueries) UpdateTask(id uuid.UUID, t *models.Task) error {
	// Use GORM to update specific fields of the task
	err := q.Model(&models.Task{}).Where("id = ?", id).Updates(map[string]interface{}{
		"title":      t.Title,
		"status":     t.Status,
		"updated_at": t.UpdatedAt,
	}).Error

	return err
}

// DeleteTask method for delete task by given ID.
func (q *TaskQueries) DeleteTask(id uuid.UUID) error {
	// Use GORM to delete task by ID
	err := q.Where("id = ?", id).Delete(&models.Task{}).Error
	return err
}
