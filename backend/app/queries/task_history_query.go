package queries

import (
	"github.com/create-go-app/fiber-go-template/app/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// TaskHistoryQueries struct for queries from TaskHistory model.
type TaskHistoryQueries struct {
	*sqlx.DB
}

// GetTaskHistories method for getting all task histories.
func (q *TaskHistoryQueries) GetTaskHistories() ([]models.TaskHistory, error) {
	// Define histories variable.
	histories := []models.TaskHistory{}

	// Define query string.
	query := `SELECT * FROM task_histories`

	// Send query to database.
	err := q.Select(&histories, query)
	if err != nil {
		// Return empty object and error.
		return histories, err
	}

	// Return query result.
	return histories, nil
}

// GetTaskHistoriesByTaskID method for getting histories by task ID.
func (q *TaskHistoryQueries) GetTaskHistoriesByTaskID(taskID uuid.UUID) ([]models.TaskHistory, error) {
	// Define histories variable.
	histories := []models.TaskHistory{}

	// Define query string.
	query := `SELECT * FROM task_histories WHERE task_id = $1 ORDER BY created_at DESC`

	// Send query to database.
	err := q.Select(&histories, query, taskID)
	if err != nil {
		// Return empty object and error.
		return histories, err
	}

	// Return query result.
	return histories, nil
}

// GetTaskHistory method for getting one task history by ID.
func (q *TaskHistoryQueries) GetTaskHistory(id uuid.UUID) (models.TaskHistory, error) {
	// Define history variable.
	history := models.TaskHistory{}

	// Define query string.
	query := `SELECT * FROM task_histories WHERE id = $1`

	// Send query to database.
	err := q.Get(&history, query, id)
	if err != nil {
		// Return empty object and error.
		return history, err
	}

	// Return query result.
	return history, nil
}

// CreateTaskHistory method for creating task history by given TaskHistory object.
func (q *TaskHistoryQueries) CreateTaskHistory(h *models.TaskHistory) error {
	// Define query string.
	query := `
		INSERT INTO task_histories 
		VALUES ($1, $2, $3, $4, $5)
	`

	// Send query to database.
	_, err := q.Exec(
		query,
		h.ID,
		h.TaskID,
		h.Action,
		h.CreatedBy,
		h.CreatedAt,
	)
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}
