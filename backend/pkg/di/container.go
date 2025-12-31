package di

import (
	"os"

	"github.com/create-go-app/fiber-go-template/app/controllers"
	apprepos "github.com/create-go-app/fiber-go-template/app/interfaces/repositories"
	"github.com/create-go-app/fiber-go-template/app/interfaces/services"
	"github.com/create-go-app/fiber-go-template/app/repository"
	serviceimpl "github.com/create-go-app/fiber-go-template/app/services"
	"github.com/create-go-app/fiber-go-template/pkg/middleware"
	"github.com/create-go-app/fiber-go-template/platform/cache"
	"github.com/create-go-app/fiber-go-template/platform/database"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Container struct {
	DB              *gorm.DB
	Cache           *cache.CacheService
	UserRepo        apprepos.UserRepository
	TaskRepo        apprepos.TaskRepository
	TaskHistoryRepo apprepos.TaskHistoryRepository
	AuthService     services.AuthService
	TokenService    services.TokenService
	TaskService     services.TaskService
	AuthController  *controllers.AuthController
	TokenController *controllers.TokenController
	TaskController  *controllers.TaskController
	JWTMiddleware   func(*fiber.Ctx) error
}

func NewContainer() (*Container, error) {
	gormDB, err := database.OpenGORMDBConnection()
	if err != nil {
		return nil, err
	}

	cacheService, err := cache.NewCacheService()
	if err != nil {
		return nil, err
	}

	var userRepo apprepos.UserRepository = repository.NewUserRepository(gormDB)
	var taskRepo apprepos.TaskRepository = repository.NewTaskRepository(gormDB)
	var taskHistoryRepo apprepos.TaskHistoryRepository = repository.NewTaskHistoryRepository(gormDB)
	var txManager apprepos.TransactionManager = database.NewGormTransactionManager(gormDB)

	authService := serviceimpl.NewAuthService(userRepo, cacheService)
	tokenService := serviceimpl.NewTokenService(userRepo, cacheService)
	taskService := serviceimpl.NewTaskService(taskRepo, taskHistoryRepo, txManager)

	authCtrl := controllers.NewAuthController(authService)
	tokenCtrl := controllers.NewTokenController(tokenService)
	taskCtrl := controllers.NewTaskController(taskService)

	jwtConfig := middleware.JWTConfig{
		SecretKey: os.Getenv("JWT_SECRET_KEY"),
	}
	jwtMiddleware := middleware.NewJWTProtected(jwtConfig)

	return &Container{
		DB:              gormDB,
		Cache:           cacheService,
		UserRepo:        userRepo,
		TaskRepo:        taskRepo,
		TaskHistoryRepo: taskHistoryRepo,
		AuthService:     authService,
		TokenService:    tokenService,
		TaskService:     taskService,
		AuthController:  authCtrl,
		TokenController: tokenCtrl,
		TaskController:  taskCtrl,
		JWTMiddleware:   jwtMiddleware,
	}, nil
}
