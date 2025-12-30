package controllers

import (
	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/create-go-app/fiber-go-template/app/services"
	"github.com/create-go-app/fiber-go-template/pkg/core"

	"github.com/gofiber/fiber/v2"
)

// RenewTokens method for renew access and refresh tokens.
// @Description Renew access and refresh tokens.
// @Summary renew access and refresh tokens
// @Tags Token
// @Accept json
// @Produce json
// @Param refresh_token body string true "Refresh token"
// @Success 200 {string} status "ok"
// @Security ApiKeyAuth
// @Router /v1/token/renew [post]
func RenewTokens(c *fiber.Ctx) error {
	// Create a new renew refresh token struct.
	renew := &models.Renew{}

	// Checking received data from JSON body.
	if err := c.BodyParser(renew); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(core.Error(400, "bad request", err.Error(), nil))
	}

	resp, err := (&services.DefaultTokenService{}).Renew(c.Context(), c, renew.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(core.Error(500, "internal error", err.Error(), nil))
	}
	return c.Status(resp.Code).JSON(resp)
}
