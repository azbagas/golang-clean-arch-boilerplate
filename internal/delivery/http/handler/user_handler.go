package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/azbagas/golang-clean-arch-boilerplate/internal/delivery/http/dto"
	"github.com/azbagas/golang-clean-arch-boilerplate/internal/domain"
	"github.com/azbagas/golang-clean-arch-boilerplate/pkg/response"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// UserHandler handles HTTP requests for user operations.
type UserHandler struct {
	userUsecase domain.UserUsecase
	validate    *validator.Validate
}

// NewUserHandler creates a new UserHandler instance.
func NewUserHandler(userUsecase domain.UserUsecase) *UserHandler {
	return &UserHandler{
		userUsecase: userUsecase,
		validate:    validator.New(),
	}
}

// Create godoc
// @Summary      Create a new user
// @Description  Create a new user with the provided information
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateUserRequest true "Create User Request"
// @Success      201 {object} response.SuccessResponse{data=dto.UserResponse}
// @Failure      400 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/users [post]
func (h *UserHandler) Create(c *fiber.Ctx) error {
	var req dto.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorWithMessage(c, http.StatusBadRequest, "invalid request body")
	}

	if err := h.validate.Struct(&req); err != nil {
		return response.ErrorWithMessage(c, http.StatusBadRequest, err.Error())
	}

	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	if err := h.userUsecase.Create(c.Context(), user); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, http.StatusCreated, toUserResponse(user))
}

// GetAll godoc
// @Summary      Get all users
// @Description  Retrieve a list of all users
// @Tags         users
// @Produce      json
// @Success      200 {object} response.SuccessResponse{data=[]dto.UserResponse}
// @Security     BearerAuth
// @Router       /api/v1/users [get]
func (h *UserHandler) GetAll(c *fiber.Ctx) error {
	users, err := h.userUsecase.GetAll(c.Context())
	if err != nil {
		return response.Error(c, err)
	}

	var userResponses []dto.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, toUserResponse(&user))
	}

	return response.Success(c, http.StatusOK, userResponses)
}

// GetByID godoc
// @Summary      Get user by ID
// @Description  Retrieve a user by their ID
// @Tags         users
// @Produce      json
// @Param        id path int true "User ID"
// @Success      200 {object} response.SuccessResponse{data=dto.UserResponse}
// @Failure      404 {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/users/{id} [get]
func (h *UserHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return response.ErrorWithMessage(c, http.StatusBadRequest, "invalid user id")
	}

	user, err := h.userUsecase.GetByID(c.Context(), uint(id))
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, http.StatusOK, toUserResponse(user))
}

// Update godoc
// @Summary      Update a user
// @Description  Update user information by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id path int true "User ID"
// @Param        request body dto.UpdateUserRequest true "Update User Request"
// @Success      200 {object} response.SuccessResponse{data=dto.UserResponse}
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/users/{id} [put]
func (h *UserHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return response.ErrorWithMessage(c, http.StatusBadRequest, "invalid user id")
	}

	var req dto.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorWithMessage(c, http.StatusBadRequest, "invalid request body")
	}

	if err := h.validate.Struct(&req); err != nil {
		return response.ErrorWithMessage(c, http.StatusBadRequest, err.Error())
	}

	user := &domain.User{
		ID:    uint(id),
		Name:  req.Name,
		Email: req.Email,
	}

	if err := h.userUsecase.Update(c.Context(), user); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, http.StatusOK, toUserResponse(user))
}

// Delete godoc
// @Summary      Delete a user
// @Description  Delete a user by their ID
// @Tags         users
// @Produce      json
// @Param        id path int true "User ID"
// @Success      200 {object} response.SuccessResponse{data=dto.DeleteUserResponse}
// @Failure      404 {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/users/{id} [delete]
func (h *UserHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return response.ErrorWithMessage(c, http.StatusBadRequest, "invalid user id")
	}

	if err := h.userUsecase.Delete(c.Context(), uint(id)); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, http.StatusOK, dto.DeleteUserResponse{
		Message: "User deleted successfully",
	})
}

// toUserResponse converts a domain.User to a dto.UserResponse.
func toUserResponse(user *domain.User) dto.UserResponse {
	return dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}
