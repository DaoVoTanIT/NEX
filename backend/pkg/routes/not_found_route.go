package routes

import (
	"github.com/create-go-app/fiber-go-template/pkg/core"
	"github.com/gofiber/fiber/v2"
)

// NotFoundRoute func for describe 404 Error route.
func NotFoundRoute(a *fiber.App) {
	// Register new special route.
	a.Use(
		// Anonymous function.
		func(c *fiber.Ctx) error {
			// Return HTTP 404 status and JSON response.
			return c.Status(fiber.StatusNotFound).JSON(core.Error(fiber.StatusNotFound, "sorry, endpoint is not found", nil, nil))
		},
	)
}
