package routes

import (
	"github.com/create-go-app/fiber-go-template/app/controllers"
	"github.com/gofiber/fiber/v2"
)

// PublicRoutes func for describe group of public routes.
func PublicRoutes(a *fiber.App, auth *controllers.AuthController, wallet *controllers.WalletController) {
	// Create routes group.
	route := a.Group("/api/v1")

	// Routes for POST method:
	route.Post("/user/sign/up", auth.UserSignUp)
	route.Post("/user/sign/in", auth.UserSignIn)
	route.Post("/wallet", wallet.CreateWallet)
	route.Post("/wallet/restore", wallet.RestoreWallet)

}
