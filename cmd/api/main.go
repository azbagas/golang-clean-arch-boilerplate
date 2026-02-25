package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/azbagas/golang-clean-arch-boilerplate/internal/delivery/http"
	"github.com/azbagas/golang-clean-arch-boilerplate/internal/delivery/http/handler"
	"github.com/azbagas/golang-clean-arch-boilerplate/internal/domain"
	postgresRepo "github.com/azbagas/golang-clean-arch-boilerplate/internal/repository/postgres"
	"github.com/azbagas/golang-clean-arch-boilerplate/internal/usecase"
	"github.com/azbagas/golang-clean-arch-boilerplate/pkg/config"
	"github.com/azbagas/golang-clean-arch-boilerplate/pkg/database"
	"github.com/azbagas/golang-clean-arch-boilerplate/pkg/logger"
	"github.com/gofiber/fiber/v2"

	_ "github.com/azbagas/golang-clean-arch-boilerplate/docs"
)

// @title           Golang Clean Architecture API
// @version         1.0
// @description     A production-ready REST API boilerplate following clean architecture principles.
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load config
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Init logger
	appLogger := logger.NewLogger(cfg.Logger.Level, cfg.Server.Mode)
	appLogger.Info().Msg("starting application...")

	// Connect to database
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("failed to connect to database")
	}
	appLogger.Info().Msg("database connected successfully")

	// Auto migrate
	if err := db.AutoMigrate(&domain.User{}, &domain.RefreshToken{}); err != nil {
		appLogger.Fatal().Err(err).Msg("failed to run auto migration")
	}
	appLogger.Info().Msg("database migration completed")

	// Init layers (manual dependency injection)
	userRepo := postgresRepo.NewUserRepository(db)
	refreshTokenRepo := postgresRepo.NewRefreshTokenRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo)
	authUsecase := usecase.NewAuthUsecase(
		userRepo,
		refreshTokenRepo,
		cfg.JWT.Secret,
		cfg.JWT.AccessExpiry,
		cfg.JWT.RefreshExpiry,
	)
	userHandler := handler.NewUserHandler(userUsecase)
	secureCookie := cfg.Server.Mode == "production"
	authHandler := handler.NewAuthHandler(authUsecase, cfg.JWT.RefreshExpiry, secureCookie)

	// Setup Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Golang Clean Architecture API",
	})

	// Register routes
	http.NewRouter(app, appLogger, cfg.JWT.Secret, userHandler, authHandler)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Listen(fmt.Sprintf(":%s", cfg.Server.Port)); err != nil {
			appLogger.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	appLogger.Info().Str("port", cfg.Server.Port).Msg("server started")

	<-quit
	appLogger.Info().Msg("shutting down server...")

	if err := app.Shutdown(); err != nil {
		appLogger.Fatal().Err(err).Msg("server forced to shutdown")
	}

	appLogger.Info().Msg("server exited gracefully")
}
