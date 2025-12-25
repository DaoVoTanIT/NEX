package queries

import (
	"github.com/create-go-app/fiber-go-template/app/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// TaskQueries struct for queries from Task model.
type TaskQueries struct {
	*sqlx.DB
}

// GetTasks method for getting all tasks.
func (q *TaskQueries) GetTasks() ([]models.Task, error) {
	// Define tasks variable.
	tasks := []models.Task{}

	// Define query string.
	query := `SELECT * FROM tasks`

	// Send query to database.
	err := q.Select(&tasks, query)
	if err != nil {
		// Return empty object and error.
		return tasks, err
	}

	// Return query result.
	return tasks, nil
}

// GetTask method for getting one task by given ID.
func (q *TaskQueries) GetTask(id uuid.UUID) (models.Task, error) {
	// Define task variable.
	task := models.Task{}

	// Define query string.
	query := `SELECT * FROM tasks WHERE id = $1`

	// Send query to database.
	err := q.Get(&task, query, id)
	if err != nil {
		// Return empty object and error.
		return task, err
	}

	// Return query result.
	return task, nil
}

// GetTasksByStatus method for getting all tasks by status.
func (q *TaskQueries) GetTasksByStatus(status int) ([]models.Task, error) {
	// Define tasks variable.
	tasks := []models.Task{}

	// Define query string.
	query := `SELECT * FROM tasks WHERE status = $1`

	// Send query to database.
	err := q.Select(&tasks, query, status)
	if err != nil {
		// Return empty object and error.
		return tasks, err
	}

	// Return query result.
	return tasks, nil
}

// CreateTask method for creating task by given Task object.
func (q *TaskQueries) CreateTask(t *models.Task) error {
	query := `
		INSERT INTO tasks (
			id,
			title,
			description,
			status,
			created_by,
			assigned_to,
			meta,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := q.Exec(
		query,
		t.ID,
		t.Title,
		t.Description,
		t.Status,
		t.CreatedBy,
		t.AssignedTo, // *uuid.UUID hoặc nil
		t.Meta,       // map[string]interface{} hoặc []byte
		t.CreatedAt,
		t.UpdatedAt,
	)

	return err
}

// UpdateTask method for updating task by given Task object.
func (q *TaskQueries) UpdateTask(id uuid.UUID, t *models.Task) error {
	// Define query string.
	query := `
		UPDATE tasks 
		SET updated_at = $2, title = $3, status = $4
		WHERE id = $1
	`

	// Send query to database.
	_, err := q.Exec(
		query,
		id,
		t.UpdatedAt,
		t.Title,
		t.Status,
	)
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}

// DeleteTask method for delete task by given ID.
func (q *TaskQueries) DeleteTask(id uuid.UUID) error {
	// Define query string.
	query := `DELETE FROM tasks WHERE id = $1`

	// Send query to database.
	_, err := q.Exec(query, id)
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}
