package di

import (
	"os"

	"github.com/create-go-app/fiber-go-template/app/controllers"
	apprepos "github.com/create-go-app/fiber-go-template/app/interfaces/repositories"
	"github.com/create-go-app/fiber-go-template/app/interfaces/services"
	"github.com/create-go-app/fiber-go-template/app/repository"
	serviceimpl "github.com/create-go-app/fiber-go-template/app/services"
	"github.com/create-go-app/fiber-go-template/pkg/crypto"
	"github.com/create-go-app/fiber-go-template/pkg/middleware"
	"github.com/create-go-app/fiber-go-template/platform/cache"
	"github.com/create-go-app/fiber-go-template/platform/database"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Container struct {
	DB               *gorm.DB
	Cache            *cache.CacheService
	UserRepo         apprepos.UserRepository
	AuthService      services.AuthService
	TokenService     services.TokenService
	AuthController   *controllers.AuthController
	TokenController  *controllers.TokenController
	WalletService    services.WalletService
	WalletController *controllers.WalletController
	JWTMiddleware    func(*fiber.Ctx) error
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
	var txManager apprepos.TransactionManager = database.NewGormTransactionManager(gormDB)

	var userRepo apprepos.UserRepository = repository.NewUserRepository(gormDB)

	authService := serviceimpl.NewAuthService(userRepo, cacheService)
	tokenService := serviceimpl.NewTokenService(userRepo, cacheService)

	authCtrl := controllers.NewAuthController(authService)
	tokenCtrl := controllers.NewTokenController(tokenService)

	jwtConfig := middleware.JWTConfig{
		SecretKey: os.Getenv("JWT_SECRET_KEY"),
	}
	jwtMiddleware := middleware.NewJWTProtected(jwtConfig)
	// Wallet
	cryptoService := crypto.NewCryptoService()
	walletRepo := repository.NewWalletRepository(gormDB)
	addressRepo := repository.NewBlockchainAddressRepository(gormDB)

	walletService := serviceimpl.NewWalletService(
		walletRepo,
		addressRepo,
		cryptoService,
		txManager,
	)

	walletController := controllers.NewWalletController(walletService)
	return &Container{
		DB:               gormDB,
		Cache:            cacheService,
		UserRepo:         userRepo,
		AuthService:      authService,
		TokenService:     tokenService,
		AuthController:   authCtrl,
		TokenController:  tokenCtrl,
		JWTMiddleware:    jwtMiddleware,
		WalletService:    walletService,
		WalletController: walletController,
	}, nil
}
