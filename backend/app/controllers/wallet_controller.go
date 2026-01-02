package controllers

import (
	"github.com/create-go-app/fiber-go-template/app/dto"
	"github.com/create-go-app/fiber-go-template/app/interfaces/services"
	"github.com/create-go-app/fiber-go-template/pkg/core"
	"github.com/create-go-app/fiber-go-template/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

type WalletController struct {
	walletService services.WalletService
}

func NewWalletController(s services.WalletService) *WalletController {
	return &WalletController{s}
}

// CreateWallet godoc
// @Summary Create a new wallet
// @Description Create a new crypto wallet with generated mnemonic and first blockchain address
// @Tags Wallet
// @Accept json
// @Produce json
// @Param data body dto.CreateWalletReq true "Create wallet payload"
// @Success 201 {object} dto.CreateWalletRes
// @Failure 400 {object} core.ApiResponse "Invalid request"
// @Failure 500 {object} core.ApiResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /v1/wallet [post]
func (c *WalletController) CreateWallet(ctx *fiber.Ctx) error {

	req := new(dto.CreateWalletReq)
	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(400).JSON(
			core.Error(400, "invalid body", err.Error(), nil),
		)
	}

	res, err := c.walletService.CreateWallet(
		ctx.Context(),
		req,
	)
	if err != nil {
		return ctx.Status(500).JSON(
			core.Error(500, "create wallet failed", err.Error(), nil),
		)
	}

	return ctx.Status(201).JSON(
		core.Success(201, "wallet created", res, nil),
	)
}

// RestoreWallet godoc
// @Summary Restore / Access existing wallet
// @Description Restore access to an existing wallet using secret phrase and optional passphrase.
// @Description If the wallet was protected with a passphrase, the correct passphrase must be provided.
// @Description On success, returns wallet identifier and associated blockchain addresses.
// @Tags Wallet
// @Accept json
// @Produce json
// @Param data body dto.RestoreWalletReq true "Restore wallet payload (secret phrase and optional passphrase)"
// @Success 200 {object} core.ApiResponse{data=dto.RestoreWalletRes} "Wallet restored successfully"
// @Failure 400 {object} core.ApiResponse "Invalid secret phrase or passphrase"
// @Failure 500 {object} core.ApiResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /v1/wallet/restore [post]
func (ctl *WalletController) RestoreWallet(c *fiber.Ctx) error {
	var req dto.RestoreWalletReq

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(
			core.Error(400, "invalid body", err.Error(), nil),
		)
	}

	validate := utils.NewValidator()
	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(
			core.Error(400, "validation error", err.Error(), nil),
		)
	}

	resp, err := ctl.walletService.RestoreWallet(c.Context(), &req)
	if err != nil {
		return c.Status(500).JSON(
			core.Error(500, "internal error", err.Error(), nil),
		)
	}

	return c.Status(resp.Code).JSON(resp)
}
