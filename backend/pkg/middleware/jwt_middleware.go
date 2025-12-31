package middleware

import (
	"github.com/gofiber/fiber/v2"

	jwtMiddleware "github.com/gofiber/contrib/jwt"
)

type JWTConfig struct {
	SecretKey string
}

// NewJWTProtected func for specify routes group with JWT authentication.
// See: https://github.com/gofiber/contrib/jwt
func NewJWTProtected(cfg JWTConfig) func(*fiber.Ctx) error {
	// Create config for JWT authentication middleware.
	config := jwtMiddleware.Config{
		SigningKey:   jwtMiddleware.SigningKey{Key: []byte(cfg.SecretKey)},
		ContextKey:   "jwt", // used in private routes
		ErrorHandler: jwtError,
	}

	return jwtMiddleware.New(config)
}

func jwtError(c *fiber.Ctx, err error) error {
	// Return status 401 and failed authentication error.
	if err.Error() == "Missing or malformed JWT" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Return status 401 and failed authentication error.
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": true,
		"msg":   err.Error(),
	})
}

// func jwtError(c *fiber.Ctx, err error) error {
// 	return c.Status(400).JSON(fiber.Map{
// 		"error": true,
// 		"msg":   err.Error(),
// 		"auth":  c.Get("Authorization"),
// 	})
// }
