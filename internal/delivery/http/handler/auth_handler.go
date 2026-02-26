package handler

import (
	"net/http"
	"time"

	"github.com/azbagas/golang-clean-arch-boilerplate/internal/delivery/http/dto"
	"github.com/azbagas/golang-clean-arch-boilerplate/internal/domain"
	"github.com/azbagas/golang-clean-arch-boilerplate/pkg/response"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// AuthHandler handles HTTP requests for authentication.
type AuthHandler struct {
	authUsecase   domain.AuthUsecase
	userUsecase   domain.UserUsecase
	validate      *validator.Validate
	refreshExpiry time.Duration
	secureCookie  bool
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(authUsecase domain.AuthUsecase, userUsecase domain.UserUsecase, refreshExpiry time.Duration, secureCookie bool) *AuthHandler {
	return &AuthHandler{
		authUsecase:   authUsecase,
		userUsecase:   userUsecase,
		validate:      validator.New(),
		refreshExpiry: refreshExpiry,
		secureCookie:  secureCookie,
	}
}

// Register godoc
// @Summary      Register a new user
// @Description  Register a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "Register Request"
// @Success      201 {object} response.SuccessResponse{data=dto.RegisterResponse}
// @Failure      400 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
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

	if err := h.authUsecase.Register(c.Context(), user); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, http.StatusCreated, dto.RegisterResponse{
		Message: "user registered successfully",
	})
}

// Login godoc
// @Summary      Login
// @Description  Authenticate user and return access token. Refresh token is set as HttpOnly cookie.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Login Request"
// @Success      200 {object} response.SuccessResponse{data=dto.LoginResponse}
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorWithMessage(c, http.StatusBadRequest, "invalid request body")
	}

	if err := h.validate.Struct(&req); err != nil {
		return response.ErrorWithMessage(c, http.StatusBadRequest, err.Error())
	}

	tokenPair, err := h.authUsecase.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return response.Error(c, err)
	}

	// Set refresh token as HttpOnly cookie
	h.setRefreshTokenCookie(c, tokenPair.RefreshToken)

	return response.Success(c, http.StatusOK, dto.LoginResponse{
		AccessToken: tokenPair.AccessToken,
	})
}

// Refresh godoc
// @Summary      Refresh access token
// @Description  Exchange refresh token (from HttpOnly cookie) for a new access token and refresh token.
// @Tags         auth
// @Produce      json
// @Success      200 {object} response.SuccessResponse{data=dto.LoginResponse}
// @Failure      401 {object} response.ErrorResponse
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return response.ErrorWithMessage(c, http.StatusUnauthorized, "refresh token not found")
	}

	tokenPair, err := h.authUsecase.Refresh(c.Context(), refreshToken)
	if err != nil {
		// Clear the invalid cookie
		h.clearRefreshTokenCookie(c)
		return response.Error(c, err)
	}

	// Set new refresh token cookie (rotation)
	h.setRefreshTokenCookie(c, tokenPair.RefreshToken)

	return response.Success(c, http.StatusOK, dto.LoginResponse{
		AccessToken: tokenPair.AccessToken,
	})
}

// Logout godoc
// @Summary      Logout
// @Description  Revoke the current refresh token and clear the cookie. Requires valid access token.
// @Tags         auth
// @Produce      json
// @Success      200 {object} response.SuccessResponse{data=dto.LogoutResponse}
// @Security     BearerAuth
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")

	if err := h.authUsecase.Logout(c.Context(), refreshToken); err != nil {
		return response.Error(c, err)
	}

	// Clear the refresh token cookie
	h.clearRefreshTokenCookie(c)

	return response.Success(c, http.StatusOK, dto.LogoutResponse{
		Message: "User logged out successfully",
	})
}

// GetCurrent godoc
// @Summary      Get current user
// @Description  Get the profile of the currently authenticated user
// @Tags         auth
// @Produce      json
// @Success      200 {object} response.SuccessResponse{data=dto.UserResponse}
// @Failure      401 {object} response.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/auth/current [get]
func (h *AuthHandler) GetCurrent(c *fiber.Ctx) error {
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return response.ErrorWithMessage(c, http.StatusUnauthorized, "invalid user identity")
	}
	userID := uint(userIDFloat)

	user, err := h.userUsecase.GetByID(c.Context(), userID)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, http.StatusOK, dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	})
}

// setRefreshTokenCookie sets the refresh token as an HttpOnly cookie.
func (h *AuthHandler) setRefreshTokenCookie(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    token,
		HTTPOnly: true,
		Secure:   h.secureCookie,
		SameSite: "Lax",
		Path:     "/api/v1/auth",
		MaxAge:   int(h.refreshExpiry.Seconds()),
	})
}

// clearRefreshTokenCookie clears the refresh token cookie.
func (h *AuthHandler) clearRefreshTokenCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   h.secureCookie,
		SameSite: "Lax",
		Path:     "/api/v1/auth",
		MaxAge:   -1,
	})
}
