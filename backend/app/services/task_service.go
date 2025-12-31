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
	"github.com/create-go-app/fiber-go-template/pkg/utils"
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
		return core.Error(404, "tasks not found", err.Error(), map[string]any{
			"count": 0,
		}), nil
	}
	mapper := &genmapper.TaskMapperImpl{}
	res := mapper.EntitiesToResList(tasks)
	return core.Success(200, "ok", res, nil), nil
}

func (s *TaskServiceImpl) GetTask(ctx context.Context, id string) (*core.ApiResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return core.Error(400, "bad request", err.Error(), nil), nil
	}
	task, err := s.taskRepo.GetTask(ctx, uid)
	if err != nil {
		return core.Error(404, "task not found", err.Error(), nil), nil
	}
	mapper := &genmapper.TaskMapperImpl{}
	res := mapper.EntityToRes(task)
	return core.Success(200, "ok", res, nil), nil
}

func (s *TaskServiceImpl) Create(ctx context.Context, userID uuid.UUID, req *dto.CreateTaskReq) (*core.ApiResponse, error) {
	validate := utils.NewValidator()
	if err := validate.Struct(req); err != nil {
		return core.Error(400, "validation error", utils.ValidatorErrors(err), nil), nil
	}

	mapper := &genmapper.TaskMapperImpl{}
	taskEntity := mapper.CreateReqToEntity(*req)
	taskEntity.ID = uuid.New()
	taskEntity.Status = "NEW"
	taskEntity.CreatedBy = userID
	taskEntity.CreatedAt = time.Now()
	task := &taskEntity

	// Start a transaction
	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		// Create the task
		if err := s.taskRepo.CreateTask(ctx, task); err != nil {
			return err
		}

		// Create the history record after the task is created
		if err := s.taskHistoryRepo.CreateTaskHistory(ctx, &models.TaskHistory{
			ID:        uuid.New(),
			TaskID:    task.ID,
			Action:    "create",
			CreatedBy: userID,
			CreatedAt: time.Now(),
			OldValue:  "new",
			NewValue:  "new",
		}); err != nil {
			return err
		}

		// If everything is successful, return nil to commit the transaction
		return nil
	})

	// Check for transaction errors
	if err != nil {
		return core.Error(500, "create task failed", err.Error(), nil), nil
	}

	res := mapper.EntityToRes(*task)
	return core.Success(201, "created", res, nil), nil
}

func (s *TaskServiceImpl) Update(ctx context.Context, userID uuid.UUID, task *models.Task) (*core.ApiResponse, error) {
	oldTask, err := s.taskRepo.GetTask(ctx, task.ID)
	if err != nil {
		return core.Error(404, "task not found", err.Error(), nil), nil
	}

	if oldTask.CreatedBy != userID {
		return core.Error(403, "only creator can update", nil, nil), nil
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
			CreatedBy: userID,
			CreatedAt: time.Now(),
		})
	}); err != nil {
		return core.Error(500, "update failed", err.Error(), nil), nil
	}

	return core.Success(200, "updated", nil, nil), nil
}

func (s *TaskServiceImpl) Delete(ctx context.Context, userID uuid.UUID, id string) (*core.ApiResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return core.Error(400, "bad request", err.Error(), nil), nil
	}

	oldTask, err := s.taskRepo.GetTask(ctx, uid)
	if err != nil {
		return core.Error(404, "task not found", err.Error(), nil), nil
	}

	if oldTask.CreatedBy != userID {
		return core.Error(403, "only creator can delete", nil, nil), nil
	}

	if err := s.txManager.Do(ctx, func(ctx context.Context) error {
		if err := s.taskRepo.DeleteTask(ctx, uid); err != nil {
			return err
		}

		return s.taskHistoryRepo.CreateTaskHistory(ctx, &models.TaskHistory{
			ID:        uuid.New(),
			TaskID:    uid,
			Action:    "delete",
			CreatedBy: userID,
			CreatedAt: time.Now(),
		})
	}); err != nil {
		return core.Error(500, "delete failed", err.Error(), nil), nil
	}

	return core.Success(204, "deleted", nil, nil), nil
}
