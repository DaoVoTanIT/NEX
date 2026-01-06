package main

import (
	"context"
	"os"

	"github.com/create-go-app/fiber-go-template/pkg/configs"
	"github.com/create-go-app/fiber-go-template/pkg/di"
	"github.com/create-go-app/fiber-go-template/pkg/middleware"
	"github.com/create-go-app/fiber-go-template/pkg/routes"
	"github.com/create-go-app/fiber-go-template/pkg/utils"
	"github.com/gofiber/fiber/v2"

	_ "github.com/create-go-app/fiber-go-template/docs"
	_ "github.com/joho/godotenv/autoload"
)

// @title API
// @version 1.0
// @description This is an auto-generated API Docs.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email your@mail.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /api
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// Define Fiber config.
	config := configs.FiberConfig()

	// Define a new Fiber app with config.
	app := fiber.New(config)

	ctx := context.Background()
	container, err := di.NewContainer(ctx)
	if err != nil {
		panic(err)
	}

	// Middlewares.
	middleware.FiberMiddleware(app) // Register Fiber's middleware for app.

	// Routes.
	routes.SwaggerRoute(app) // Register a route for API Docs (Swagger).
	routes.HealthRoute(app, container)
	routes.PublicRoutes(app, container.AuthController, container.WalletController)
	routes.PrivateRoutes(app, container.JWTMiddleware, container.AuthController, container.TokenController, container.WalletController)
	routes.NotFoundRoute(app) // Register route for 404 Error.

	// Start server (with or without graceful shutdown).
	if os.Getenv("STAGE_STATUS") == "dev" {
		utils.StartServer(app)
	} else {
		utils.StartServerWithGracefulShutdown(app)
	}
}
