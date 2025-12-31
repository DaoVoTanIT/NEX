package middleware

import (
	"time"

	"github.com/create-go-app/fiber-go-template/pkg/core"
	"github.com/create-go-app/fiber-go-template/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

func RequireCredentials(required ...string) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		claims, err := utils.ExtractTokenMetadata(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(core.Error(fiber.StatusUnauthorized, "unauthorized", err.Error(), nil))
		}
		if time.Now().Unix() > claims.Expires {
			return c.Status(fiber.StatusUnauthorized).JSON(core.Error(fiber.StatusUnauthorized, "token expired", nil, nil))
		}
		for _, cred := range required {
			if !claims.Credentials[cred] {
				return c.Status(fiber.StatusForbidden).JSON(core.Error(fiber.StatusForbidden, "permission denied", nil, nil))
			}
		}
		c.Locals("userID", claims.UserID)
		c.Locals("claims", claims)
		return c.Next()
	}
}
