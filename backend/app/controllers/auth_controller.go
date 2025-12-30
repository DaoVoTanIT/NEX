package controllers

import (
	"context"

	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/create-go-app/fiber-go-template/app/services"
	"github.com/create-go-app/fiber-go-template/pkg/core"

	"github.com/gofiber/fiber/v2"
)

// UserSignUp method to create a new user.
// @Summary create a new user
// @Description Create a new user
// @Tags User
// @Accept json
// @Produce json
// @Param data body models.SignUp true "Sign up payload"
// @Success 200 {object} models.User
// @Router /v1/user/sign/up [post]
func UserSignUp(c *fiber.Ctx) error {
	signUp := &models.SignUp{}

	if err := c.BodyParser(signUp); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(core.Error(fiber.StatusBadRequest, "bad request", err.Error(), nil))
	}

	resp, err := (&services.DefaultAuthService{}).SignUp(context.Background(), signUp)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(core.Error(fiber.StatusInternalServerError, "internal error", err.Error(), nil))
	}
	return c.Status(resp.Code).JSON(resp)
}

// UserSignIn godoc
// @Summary auth user and return access and refresh token
// @Description Auth user and return access and refresh token
// @Tags User
// @Accept json
// @Produce json
// @Param data body models.SignIn true "Sign in payload"
// @Success 200 {object} map[string]interface{}
// @Router /v1/user/sign/in [post]
func UserSignIn(c *fiber.Ctx) error {
	signIn := &models.SignIn{}

	if err := c.BodyParser(signIn); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(core.Error(fiber.StatusBadRequest, "bad request", err.Error(), nil))
	}

	resp, err := (&services.DefaultAuthService{}).SignIn(context.Background(), signIn)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(core.Error(fiber.StatusInternalServerError, "internal error", err.Error(), nil))
	}
	return c.Status(resp.Code).JSON(resp)
}

// UserSignOut method to de-authorize user and delete refresh token from Redis.
// @Description De-authorize user and delete refresh token from Redis.
// @Summary de-authorize user and delete refresh token from Redis
// @Tags User
// @Accept json
// @Produce json
// @Success 204 {string} status "ok"
// @Security ApiKeyAuth
// @Router /v1/user/sign/out [post]
func UserSignOut(c *fiber.Ctx) error {
	resp, err := (&services.DefaultAuthService{}).SignOut(context.Background(), c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(core.Error(fiber.StatusInternalServerError, "internal error", err.Error(), nil))
	}
	return c.Status(resp.Code).JSON(resp)
}
