package routes

import (
	"github.com/create-go-app/fiber-go-template/app/controllers"
	mw "github.com/create-go-app/fiber-go-template/pkg/middleware"
	"github.com/create-go-app/fiber-go-template/pkg/repository"
	"github.com/gofiber/fiber/v2"
)

// PrivateRoutes func for describe group of private routes.
func PrivateRoutes(a *fiber.App, jwtMiddleware func(*fiber.Ctx) error, auth *controllers.AuthController, token *controllers.TokenController, task *controllers.TaskController) {
	// Create routes group.
	route := a.Group("/api/v1")

	// Routes for POST method:
	route.Post("/user/sign/out", jwtMiddleware, auth.UserSignOut)
	route.Post("/token/renew", jwtMiddleware, token.RenewTokens)

	// Routes for Task management:
	route.Post("/task", jwtMiddleware, mw.RequireCredentials(repository.TaskCreateCredential), task.CreateTask)
	route.Put("/task/:id", jwtMiddleware, mw.RequireCredentials(repository.TaskUpdateCredential), task.UpdateTask)
	route.Delete("/task/:id", jwtMiddleware, mw.RequireCredentials(repository.TaskDeleteCredential), task.DeleteTask)
	route.Get("/tasks", jwtMiddleware, mw.RequireCredentials(repository.TaskViewCredential), task.GetTasks)
	route.Get("/task/:id", jwtMiddleware, mw.RequireCredentials(repository.TaskViewCredential), task.GetTask)
	//route.Get("/tasks/history/:id", middleware, controllers.GetTaskHistory) // get task history by task ID

}
