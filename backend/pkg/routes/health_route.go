package routes

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/create-go-app/fiber-go-template/pkg/core"
	"github.com/create-go-app/fiber-go-template/pkg/di"
)

func HealthRoute(a *fiber.App, container *di.Container) {
	handler := func(c *fiber.Ctx) error {
		ctx := context.Background()

		redisStatus := "healthy"
		if err := container.Cache.GetClient().HealthCheck(ctx); err != nil {
			redisStatus = "unhealthy"
		}

		dbStatus := "healthy"
		sqlDB, err := container.DB.DB()
		if err != nil || sqlDB == nil {
			dbStatus = "unhealthy"
		} else if err := sqlDB.PingContext(ctx); err != nil {
			dbStatus = "unhealthy"
		}

		overall := "healthy"
		if redisStatus != "healthy" || dbStatus != "healthy" {
			overall = "unhealthy"
		}

		return c.Status(fiber.StatusOK).JSON(core.Success(fiber.StatusOK, "health", fiber.Map{
			"status": overall,
			"redis":  redisStatus,
			"db":     dbStatus,
		}, nil))
	}
	a.Get("/health", handler)
	a.Get("/api/v1/health", handler)
}
