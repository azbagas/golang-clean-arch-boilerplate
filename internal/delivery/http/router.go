package http

import (
	"github.com/azbagas/golang-clean-arch-boilerplate/internal/delivery/http/handler"
	"github.com/azbagas/golang-clean-arch-boilerplate/internal/delivery/http/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/rs/zerolog"
)

// NewRouter sets up all routes and returns the Fiber app.
func NewRouter(
	app *fiber.App,
	log zerolog.Logger,
	jwtSecret string,
	userHandler *handler.UserHandler,
	authHandler *handler.AuthHandler,
) {
	// Global middleware
	app.Use(middleware.CORSMiddleware())
	app.Use(middleware.LoggerMiddleware(log))

	// Swagger docs
	app.Get("/swagger/*", swagger.HandlerDefault)

	// API v1
	v1 := app.Group("/api/v1")

	// Auth routes (public)
	auth := v1.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)

	// Logout and current user require valid access token
	auth.Post("/logout", middleware.JWTMiddleware(jwtSecret), authHandler.Logout)
	auth.Get("/current", middleware.JWTMiddleware(jwtSecret), authHandler.GetCurrent)

	// User routes (protected)
	users := v1.Group("/users", middleware.JWTMiddleware(jwtSecret))
	users.Post("/", userHandler.Create)
	users.Get("/", userHandler.GetAll)
	users.Get("/:id", userHandler.GetByID)
	users.Put("/:id", userHandler.Update)
	users.Delete("/:id", userHandler.Delete)
}
