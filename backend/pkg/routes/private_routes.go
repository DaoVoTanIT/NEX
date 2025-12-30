package routes

import (
	"github.com/create-go-app/fiber-go-template/app/controllers"
	"github.com/gofiber/fiber/v2"
)

// PrivateRoutes func for describe group of private routes.
func PrivateRoutes(a *fiber.App, middleware func(*fiber.Ctx) error, auth *controllers.AuthController, token *controllers.TokenController, task *controllers.TaskController) {
	// Create routes group.
	route := a.Group("/api/v1")

	// Routes for POST method:
	route.Post("/user/sign/out", middleware, auth.UserSignOut)
	route.Post("/token/renew", middleware, token.RenewTokens)

	// Routes for Task management:
	route.Post("/task", middleware, task.CreateTask)
	route.Put("/task/:id", middleware, task.UpdateTask)
	route.Delete("/task/:id", middleware, task.DeleteTask)
	route.Get("/tasks", middleware, task.GetTasks)
	route.Get("/task/:id", middleware, task.GetTask)
	//route.Get("/tasks/history/:id", middleware, controllers.GetTaskHistory) // get task history by task ID

}
