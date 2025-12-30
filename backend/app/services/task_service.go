package services

import (
	"context"
	"time"

	"github.com/create-go-app/fiber-go-template/app/dto"
	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/create-go-app/fiber-go-template/app/interfaces/repositories"
	"github.com/create-go-app/fiber-go-template/app/interfaces/services"
	"github.com/create-go-app/fiber-go-template/pkg/core"
	genmapper "github.com/create-go-app/fiber-go-template/pkg/mappers/generated"
	"github.com/create-go-app/fiber-go-template/pkg/repository"
	"github.com/create-go-app/fiber-go-template/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TaskServiceImpl struct {
	taskRepo        repositories.TaskRepository
	taskHistoryRepo repositories.TaskHistoryRepository
	txManager       repositories.TransactionManager
}

func NewTaskService(taskRepo repositories.TaskRepository, taskHistoryRepo repositories.TaskHistoryRepository, txManager repositories.TransactionManager) services.TaskService {
	return &TaskServiceImpl{
		taskRepo:        taskRepo,
		taskHistoryRepo: taskHistoryRepo,
		txManager:       txManager,
	}
}

func (s *TaskServiceImpl) GetTasks(ctx context.Context) (*core.ApiResponse, error) {
	tasks, err := s.taskRepo.GetTasks(ctx)
	if err != nil {
		return core.Error(fiber.StatusNotFound, "tasks not found", err.Error(), fiber.Map{
			"count": 0,
		}), nil
	}
	mapper := &genmapper.TaskMapperImpl{}
	res := mapper.EntitiesToResList(tasks)
	return core.Success(fiber.StatusOK, "ok", res, nil), nil
}

func (s *TaskServiceImpl) GetTask(ctx context.Context, id string) (*core.ApiResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return core.Error(fiber.StatusBadRequest, "bad request", err.Error(), nil), nil
	}
	task, err := s.taskRepo.GetTask(ctx, uid)
	if err != nil {
		return core.Error(fiber.StatusNotFound, "task not found", err.Error(), nil), nil
	}
	mapper := &genmapper.TaskMapperImpl{}
	res := mapper.EntityToRes(task)
	return core.Success(fiber.StatusOK, "ok", res, nil), nil
}

func (s *TaskServiceImpl) Create(ctx context.Context, c any, req *dto.CreateTaskReq) (*core.ApiResponse, error) {
	cc := c.(*fiber.Ctx)

	claims, err := utils.ExtractTokenMetadata(cc)
	if err != nil {
		return core.Error(fiber.StatusUnauthorized, "unauthorized", err.Error(), nil), nil
	}

	validate := utils.NewValidator()
	if err := validate.Struct(req); err != nil {
		return core.Error(fiber.StatusBadRequest, "validation error", utils.ValidatorErrors(err), nil), nil
	}

	mapper := &genmapper.TaskMapperImpl{}
	taskEntity := mapper.CreateReqToEntity(*req)
	taskEntity.ID = uuid.New()
	taskEntity.Status = "NEW"
	taskEntity.CreatedBy = claims.UserID
	taskEntity.CreatedAt = time.Now()
	task := &taskEntity

	if err := s.taskRepo.CreateTask(ctx, task); err != nil {
		return core.Error(fiber.StatusInternalServerError, "create task failed", err.Error(), nil), nil
	}

	res := mapper.EntityToRes(*task)
	return core.Success(fiber.StatusOK, "ok", res, nil), nil
}

func (s *TaskServiceImpl) Update(ctx context.Context, c any, task *models.Task) (*core.ApiResponse, error) {
	cc := c.(*fiber.Ctx)

	now := time.Now().Unix()

	claims, err := utils.ExtractTokenMetadata(cc)
	if err != nil {
		return core.Error(fiber.StatusInternalServerError, "token parse error", err.Error(), nil), nil
	}
	if now > claims.Expires {
		return core.Error(fiber.StatusUnauthorized, "token expired", nil, nil), nil
	}
	if !claims.Credentials[repository.TaskUpdateCredential] {
		return core.Error(fiber.StatusForbidden, "permission denied", nil, nil), nil
	}

	oldTask, err := s.taskRepo.GetTask(ctx, task.ID)
	if err != nil {
		return core.Error(fiber.StatusNotFound, "task not found", err.Error(), nil), nil
	}

	if oldTask.CreatedBy != claims.UserID {
		return core.Error(fiber.StatusForbidden, "only creator can update", nil, nil), nil
	}

	task.UpdatedAt = time.Now()

	if err := s.txManager.Do(ctx, func(ctx context.Context) error {
		if err := s.taskRepo.UpdateTask(ctx, task.ID, task); err != nil {
			return err
		}

		return s.taskHistoryRepo.CreateTaskHistory(ctx, &models.TaskHistory{
			ID:        uuid.New(),
			TaskID:    task.ID,
			Action:    "update",
			CreatedBy: claims.UserID,
			CreatedAt: time.Now(),
		})
	}); err != nil {
		return core.Error(fiber.StatusInternalServerError, "update failed", err.Error(), nil), nil
	}

	return core.Success(201, "updated", nil, nil), nil
}

func (s *TaskServiceImpl) Delete(ctx context.Context, c any, id string) (*core.ApiResponse, error) {
	cc := c.(*fiber.Ctx)

	now := time.Now().Unix()

	claims, err := utils.ExtractTokenMetadata(cc)
	if err != nil {
		return core.Error(fiber.StatusInternalServerError, "token parse error", err.Error(), nil), nil
	}
	if now > claims.Expires {
		return core.Error(fiber.StatusUnauthorized, "token expired", nil, nil), nil
	}
	if !claims.Credentials[repository.TaskDeleteCredential] {
		return core.Error(fiber.StatusForbidden, "permission denied", nil, nil), nil
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return core.Error(fiber.StatusBadRequest, "bad request", err.Error(), nil), nil
	}

	oldTask, err := s.taskRepo.GetTask(ctx, uid)
	if err != nil {
		return core.Error(fiber.StatusNotFound, "task not found", err.Error(), nil), nil
	}

	if oldTask.CreatedBy != claims.UserID {
		return core.Error(fiber.StatusForbidden, "only creator can delete", nil, nil), nil
	}

	if err := s.txManager.Do(ctx, func(ctx context.Context) error {
		if err := s.taskRepo.DeleteTask(ctx, uid); err != nil {
			return err
		}

		return s.taskHistoryRepo.CreateTaskHistory(ctx, &models.TaskHistory{
			ID:        uuid.New(),
			TaskID:    uid,
			Action:    "delete",
			CreatedBy: claims.UserID,
			CreatedAt: time.Now(),
		})
	}); err != nil {
		return core.Error(fiber.StatusInternalServerError, "delete failed", err.Error(), nil), nil
	}

	return core.Success(fiber.StatusNoContent, "deleted", nil, nil), nil
}
