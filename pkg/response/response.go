package response

import (
	"errors"
	"net/http"

	"github.com/azbagas/golang-clean-arch-boilerplate/internal/domain"
	"github.com/gofiber/fiber/v2"
)

// SuccessResponse is the schema for 2xx responses
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data"`
}

// ErrorResponse is the schema for 4xx and 5xx responses
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"error message"`
}

// Success sends a successful JSON response.
func Success(c *fiber.Ctx, statusCode int, data interface{}) error {
	return c.Status(statusCode).JSON(SuccessResponse{
		Success: true,
		Data:    data,
	})
}

// Error sends an error JSON response, mapping domain errors to HTTP status codes.
func Error(c *fiber.Ctx, err error) error {
	statusCode := mapErrorToStatus(err)
	return c.Status(statusCode).JSON(ErrorResponse{
		Success: false,
		Error:   err.Error(),
	})
}

// ErrorWithMessage sends an error JSON response with a custom message.
func ErrorWithMessage(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(ErrorResponse{
		Success: false,
		Error:   message,
	})
}

// mapErrorToStatus maps domain errors to HTTP status codes.
func mapErrorToStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrBadRequest):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
