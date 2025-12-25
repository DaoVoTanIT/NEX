package routes

import (
	"github.com/create-go-app/fiber-go-template/app/controllers"
	"github.com/create-go-app/fiber-go-template/pkg/middleware"
	"github.com/gofiber/fiber/v2"
)

// PrivateRoutes func for describe group of private routes.
func PrivateRoutes(a *fiber.App) {
	// Create routes group.
	route := a.Group("/api/v1")

	// Routes for POST method:
	route.Post("/book", middleware.JWTProtected(), controllers.CreateBook)           // create a new book
	route.Post("/user/sign/out", middleware.JWTProtected(), controllers.UserSignOut) // de-authorization user
	route.Post("/token/renew", middleware.JWTProtected(), controllers.RenewTokens)   // renew Access & Refresh tokens

	// Routes for PUT method:
	route.Put("/book", middleware.JWTProtected(), controllers.UpdateBook) // update one book by ID

	// Routes for DELETE method:
	route.Delete("/book", middleware.JWTProtected(), controllers.DeleteBook) // delete one book by ID
	// Routes for Task management:
	route.Post("/task", middleware.JWTProtected(), controllers.CreateTask)       // create a new task
	route.Put("/task/:id", middleware.JWTProtected(), controllers.UpdateTask)    // update task by ID
	route.Delete("/task/:id", middleware.JWTProtected(), controllers.DeleteTask) // delete task by ID
	route.Get("/tasks", middleware.JWTProtected(), controllers.GetTasks)         // get all exists tasks
	route.Get("/task/:id", middleware.JWTProtected(), controllers.GetTask)       // get task by ID
	//route.Get("/tasks/history/:id", middleware.JWTProtected(), controllers.GetTaskHistory) // get task history by task ID

}
