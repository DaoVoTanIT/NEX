package controllers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/create-go-app/fiber-go-template/app/dto"
	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/create-go-app/fiber-go-template/pkg/core"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTaskService is a mock implementation of services.TaskService
type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) GetTasks(ctx context.Context) (*core.ApiResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.ApiResponse), args.Error(1)
}

func (m *MockTaskService) GetTask(ctx context.Context, id string) (*core.ApiResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.ApiResponse), args.Error(1)
}

func (m *MockTaskService) Create(ctx context.Context, c any, req *dto.CreateTaskReq) (*core.ApiResponse, error) {
	args := m.Called(ctx, c, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.ApiResponse), args.Error(1)
}

func (m *MockTaskService) Update(ctx context.Context, c any, task *models.Task) (*core.ApiResponse, error) {
	args := m.Called(ctx, c, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.ApiResponse), args.Error(1)
}

func (m *MockTaskService) Delete(ctx context.Context, c any, id string) (*core.ApiResponse, error) {
	args := m.Called(ctx, c, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.ApiResponse), args.Error(1)
}

func TestGetTasks(t *testing.T) {
	// Setup
	app := fiber.New()
	mockService := new(MockTaskService)
	controller := NewTaskController(mockService)

	app.Get("/tasks", controller.GetTasks)

	// Define expected response
	expectedResponse := &core.ApiResponse{
		Code:    200,
		Message: "success",
		Data:    []dto.TaskRes{},
	}

	// Mock expectation
	mockService.On("GetTasks", mock.Anything).Return(expectedResponse, nil)

	// Request
	req := httptest.NewRequest("GET", "/tasks", nil)
	resp, _ := app.Test(req)

	// Assertions
	assert.Equal(t, 200, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestGetTask(t *testing.T) {
	// Setup
	app := fiber.New()
	mockService := new(MockTaskService)
	controller := NewTaskController(mockService)

	app.Get("/task/:id", controller.GetTask)

	taskID := "123e4567-e89b-12d3-a456-426614174000"
	taskUUID, _ := uuid.Parse(taskID)
	expectedResponse := &core.ApiResponse{
		Code:    200,
		Message: "success",
		Data:    dto.TaskRes{ID: taskUUID},
	}

	// Mock expectation
	mockService.On("GetTask", mock.Anything, taskID).Return(expectedResponse, nil)

	// Request
	req := httptest.NewRequest("GET", "/task/"+taskID, nil)
	resp, _ := app.Test(req)

	// Assertions
	assert.Equal(t, 200, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestCreateTask(t *testing.T) {
	// Setup
	app := fiber.New()
	mockService := new(MockTaskService)
	controller := NewTaskController(mockService)

	app.Post("/task", controller.CreateTask)

	createReq := &dto.CreateTaskReq{
		Title: "New Task",
	}
	body, _ := json.Marshal(createReq)

	expectedResponse := &core.ApiResponse{
		Code:    201,
		Message: "created",
		Data:    dto.TaskRes{Title: "New Task"},
	}

	// Mock expectation - Note: the second argument 'c' is the fiber context, using mock.Anything
	mockService.On("Create", mock.Anything, mock.Anything, createReq).Return(expectedResponse, nil)

	// Request
	req := httptest.NewRequest("POST", "/task", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	// Assertions
	assert.Equal(t, 201, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestUpdateTask(t *testing.T) {
	// Setup
	app := fiber.New()
	mockService := new(MockTaskService)
	controller := NewTaskController(mockService)

	app.Put("/task", controller.UpdateTask)

	taskID := "123e4567-e89b-12d3-a456-426614174000"
	// Create UUID for test
	// We need to match the struct passed to Update
	// The controller parses body into models.Task

	// updateTask.ID, _ = uuid.Parse(taskID)
	// For JSON unmarshalling in controller, we need to send valid JSON matching models.Task fields.
	// However, models.Task uses uuid.UUID which expects string format in JSON if properly tagged or handled.

	// Let's keep it simple with JSON string
	reqBody := `{"id":"` + taskID + `", "title":"Updated Task"}`

	expectedResponse := &core.ApiResponse{
		Code:    200,
		Message: "updated",
	}

	// Mock expectation
	// Since controller parses body into a struct pointer, we should match loosely or use a custom matcher if needed.
	// For simplicity, we match mock.AnythingOfType("*models.Task")
	mockService.On("Update", mock.Anything, mock.Anything, mock.MatchedBy(func(t *models.Task) bool {
		return t.ID.String() == taskID && t.Title == "Updated Task"
	})).Return(expectedResponse, nil)

	// Request
	req := httptest.NewRequest("PUT", "/task", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	// Assertions
	assert.Equal(t, 200, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestDeleteTask(t *testing.T) {
	// Setup
	app := fiber.New()
	mockService := new(MockTaskService)
	controller := NewTaskController(mockService)

	app.Delete("/task", controller.DeleteTask)

	taskID := "123e4567-e89b-12d3-a456-426614174000"
	reqBody := `{"id":"` + taskID + `"}`

	expectedResponse := &core.ApiResponse{
		Code:    204,
		Message: "deleted",
	}

	// Mock expectation
	mockService.On("Delete", mock.Anything, mock.Anything, taskID).Return(expectedResponse, nil)

	// Request
	req := httptest.NewRequest("DELETE", "/task", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	// Assertions
	assert.Equal(t, 204, resp.StatusCode)
	mockService.AssertExpectations(t)
}
