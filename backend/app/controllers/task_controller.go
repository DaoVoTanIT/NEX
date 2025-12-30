package controllers

import (
	"github.com/create-go-app/fiber-go-template/app/dto"
	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/create-go-app/fiber-go-template/app/interfaces/services"
	"github.com/create-go-app/fiber-go-template/pkg/core"
	"github.com/gofiber/fiber/v2"
)

type TaskController struct {
	taskService services.TaskService
}

func NewTaskController(taskService services.TaskService) *TaskController {
	return &TaskController{taskService: taskService}
}

// GetTasks func gets all exists tasks.
// @Summary get all exists tasks
// @Tags Tasks
// @Accept json
// @Produce json
// @Success 200 {object} core.ApiResponse{data=[]dto.TaskRes}
// @Security ApiKeyAuth
// @Router /v1/tasks [get]
func (ctl *TaskController) GetTasks(c *fiber.Ctx) error {
	resp, err := ctl.taskService.GetTasks(c.Context())
	if err != nil {
		return c.Status(500).JSON(core.Error(500, "internal error", err.Error(), nil))
	}
	return c.Status(resp.Code).JSON(resp)
}

// GetTask func gets task by ID.
// @Tags Task
// @Security ApiKeyAuth
// @Param id path string true "Task ID"
// @Accept json
// @Produce json
// @Success 200 {object} dto.TaskRes
// @Router /v1/task/{id} [get]
func (ctl *TaskController) GetTask(c *fiber.Ctx) error {
	resp, err := ctl.taskService.GetTask(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(500).JSON(core.Error(500, "internal error", err.Error(), nil))
	}
	return c.Status(resp.Code).JSON(resp)
}

// CreateTask func creates a new task.
// @Tags Task
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param task body dto.CreateTaskReq true "Create task request"
// @Security ApiKeyAuth
// @Router /v1/task [post]
func (ctl *TaskController) CreateTask(c *fiber.Ctx) error {
	req := new(dto.CreateTaskReq)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(core.Error(fiber.StatusBadRequest, "bad request", err.Error(), nil))
	}

	resp, err := ctl.taskService.Create(c.Context(), c, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(core.Error(fiber.StatusInternalServerError, "internal error", err.Error(), nil))
	}
	return c.Status(resp.Code).JSON(resp)
}

// UpdateTask func updates task.
// @Tags Task
// @Security ApiKeyAuth
// @Router /v1/task [put]
func (ctl *TaskController) UpdateTask(c *fiber.Ctx) error {
	task := &models.Task{}
	if err := c.BodyParser(task); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(core.Error(fiber.StatusBadRequest, "bad request", err.Error(), nil))
	}

	resp, err := ctl.taskService.Update(c.Context(), c, task)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(core.Error(fiber.StatusInternalServerError, "internal error", err.Error(), nil))
	}
	return c.Status(resp.Code).JSON(resp)
}

// DeleteTask func deletes task.
// @Tags Task
// @Security ApiKeyAuth
// @Router /v1/task [delete]
func (ctl *TaskController) DeleteTask(c *fiber.Ctx) error {
	var req struct {
		ID string `json:"id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(core.Error(fiber.StatusBadRequest, "bad request", err.Error(), nil))
	}
	resp, err := ctl.taskService.Delete(c.Context(), c, req.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(core.Error(fiber.StatusInternalServerError, "internal error", err.Error(), nil))
	}
	return c.Status(resp.Code).JSON(resp)
}
