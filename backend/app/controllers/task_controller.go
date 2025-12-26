package controllers

import (
	"time"

	"github.com/create-go-app/fiber-go-template/app/dto"
	"github.com/create-go-app/fiber-go-template/app/models"
	"github.com/create-go-app/fiber-go-template/pkg/repository"
	"github.com/create-go-app/fiber-go-template/pkg/utils"
	"github.com/create-go-app/fiber-go-template/platform/database"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GetTasks func gets all exists tasks.
// @Summary get all exists tasks
// @Tags Tasks
// @Accept json
// @Produce json
// @Success 200 {array} models.Task
// @Security ApiKeyAuth
// @Router /v1/tasks [get]
func GetTasks(c *fiber.Ctx) error {
	db, err := database.OpenDBConnection()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "msg": err.Error()})
	}

	tasks, err := db.GetTasks()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": true,
			"msg":   "tasks not found",
			"count": 0,
			"tasks": nil,
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"msg":   nil,
		"count": len(tasks),
		"tasks": tasks,
	})
}

// GetTask func gets task by ID.
// @Tags Task
// @Security ApiKeyAuth
// @Param id path string true "Task ID"
// @Accept json
// @Produce json
// @Success 200 {object} dto.TaskRes
// @Router /v1/task/{id} [get]
func GetTask(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": true, "msg": err.Error()})
	}

	db, err := database.OpenDBConnection()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "msg": err.Error()})
	}

	task, err := db.GetTask(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": true,
			"msg":   "task not found",
			"task":  nil,
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"msg":   nil,
		"task":  task,
	})
}

// CreateTask func creates a new task.
// @Tags Task
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param task body dto.CreateTaskReq true "Create task request"
// @Security ApiKeyAuth
// @Router /v1/task [post]
func CreateTask(c *fiber.Ctx) error {
	claims, err := utils.ExtractTokenMetadata(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": true, "msg": err.Error()})
	}

	req := new(dto.CreateTaskReq)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": true, "msg": err.Error()})
	}

	validate := utils.NewValidator()
	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": true, "msg": utils.ValidatorErrors(err)})
	}

	task := &models.Task{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		Status:      "NEW",
		CreatedBy:   claims.UserID,
		CreatedAt:   time.Now(),
		AssignedTo:  req.AssignedTo,
	}

	db, err := database.OpenDBConnection()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": true, "msg": err.Error()})
	}

	if err := db.CreateTask(task); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	return c.JSON(dto.TaskRes{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedBy:   task.CreatedBy,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	})
}

// UpdateTask func updates task.
// @Tags Task
// @Security ApiKeyAuth
// @Router /v1/task [put]
func UpdateTask(c *fiber.Ctx) error {
	now := time.Now().Unix()

	claims, err := utils.ExtractTokenMetadata(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "msg": err.Error()})
	}

	if now > claims.Expires {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": true, "msg": "token expired"})
	}

	if !claims.Credentials[repository.TaskUpdateCredential] {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": true, "msg": "permission denied"})
	}

	task := &models.Task{}
	if err := c.BodyParser(task); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": true, "msg": err.Error()})
	}

	db, err := database.OpenDBConnection()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "msg": err.Error()})
	}

	oldTask, err := db.GetTask(task.ID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": true, "msg": "task not found"})
	}

	if oldTask.CreatedBy != claims.UserID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": true, "msg": "only creator can update"})
	}

	task.UpdatedAt = time.Now()

	if err := db.UpdateTask(task.ID, task); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": true, "msg": err.Error()})
	}

	_ = db.CreateTaskHistory(&models.TaskHistory{
		ID:        uuid.New(),
		TaskID:    task.ID,
		Action:    "update",
		CreatedBy: claims.UserID,
		CreatedAt: time.Now(),
	})

	return c.Status(201).JSON(fiber.Map{"error": false})
}

// DeleteTask func deletes task.
// @Tags Task
// @Security ApiKeyAuth
// @Router /v1/task [delete]
func DeleteTask(c *fiber.Ctx) error {
	now := time.Now().Unix()

	claims, err := utils.ExtractTokenMetadata(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "msg": err.Error()})
	}

	if now > claims.Expires {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": true, "msg": "token expired"})
	}

	if !claims.Credentials[repository.TaskDeleteCredential] {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": true, "msg": "permission denied"})
	}

	task := &models.Task{}
	if err := c.BodyParser(task); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": true, "msg": err.Error()})
	}

	db, err := database.OpenDBConnection()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "msg": err.Error()})
	}

	oldTask, err := db.GetTask(task.ID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": true, "msg": "task not found"})
	}

	if oldTask.CreatedBy != claims.UserID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": true, "msg": "only creator can delete"})
	}

	if err := db.DeleteTask(task.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "msg": err.Error()})
	}

	_ = db.CreateTaskHistory(&models.TaskHistory{
		ID:        uuid.New(),
		TaskID:    task.ID,
		Action:    "delete",
		CreatedBy: claims.UserID,
		CreatedAt: time.Now(),
	})

	return c.SendStatus(fiber.StatusNoContent)
}
