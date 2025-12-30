package queries

import (
	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskHistoryQueries struct {
	*gorm.DB
}

// =========================
// GET ALL TASK HISTORIES
// =========================
func (q *TaskHistoryQueries) GetTaskHistories() ([]models.TaskHistory, error) {
	var histories []models.TaskHistory

	err := q.
		Order("created_at DESC").
		Find(&histories).Error

	if err != nil {
		return nil, err
	}

	return histories, nil
}

// =========================
// GET TASK HISTORIES BY TASK ID
// =========================
func (q *TaskHistoryQueries) GetTaskHistoriesByTaskID(taskID uuid.UUID) ([]models.TaskHistory, error) {
	var histories []models.TaskHistory

	err := q.
		Where("task_id = ?", taskID).
		Order("created_at DESC").
		Find(&histories).Error

	if err != nil {
		return nil, err
	}

	return histories, nil
}

// =========================
// GET TASK HISTORY BY ID
// =========================
func (q *TaskHistoryQueries) GetTaskHistory(id uuid.UUID) (models.TaskHistory, error) {
	var history models.TaskHistory

	err := q.
		Where("id = ?", id).
		First(&history).Error

	if err != nil {
		return history, err
	}

	return history, nil
}

// =========================
// CREATE TASK HISTORY
// =========================
func (q *TaskHistoryQueries) CreateTaskHistory(h *models.TaskHistory) error {
	return q.Create(h).Error
}
