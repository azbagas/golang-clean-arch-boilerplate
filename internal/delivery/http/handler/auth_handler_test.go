package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/azbagas/golang-clean-arch-boilerplate/internal/delivery/http/handler"
	"github.com/azbagas/golang-clean-arch-boilerplate/internal/domain"
	"github.com/azbagas/golang-clean-arch-boilerplate/internal/mocks"
	"github.com/azbagas/golang-clean-arch-boilerplate/pkg/response"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupAuthTestApp() (*fiber.App, *mocks.MockAuthUsecase, *mocks.MockUserUsecase) {
	app := fiber.New()
	mockAuthUsecase := new(mocks.MockAuthUsecase)
	mockUserUsecase := new(mocks.MockUserUsecase)
	h := handler.NewAuthHandler(mockAuthUsecase, mockUserUsecase, 7*24*time.Hour, false)

	app.Post("/auth/register", h.Register)
	app.Post("/auth/login", h.Login)
	app.Post("/auth/refresh", h.Refresh)
	app.Post("/auth/logout", h.Logout)
	// Simulate JWT middleware by injecting locals
	app.Get("/auth/current", func(c *fiber.Ctx) error {
		c.Locals("user_id", float64(1))
		c.Locals("email", "john@example.com")
		return c.Next()
	}, h.GetCurrent)

	return app, mockAuthUsecase, mockUserUsecase
}

func TestAuthHandler_Register(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app, mockUsecase, _ := setupAuthTestApp()
		mockUsecase.On("Register", mock.Anything, mock.AnythingOfType("*domain.User")).
			Return(nil).Once()

		body := `{"name":"John Doe","email":"john@example.com","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var res response.SuccessResponse
		respBody, _ := io.ReadAll(resp.Body)
		json.Unmarshal(respBody, &res)
		assert.True(t, res.Success)
		mockUsecase.AssertExpectations(t)
	})

	t.Run("conflict", func(t *testing.T) {
		app, mockUsecase, _ := setupAuthTestApp()
		mockUsecase.On("Register", mock.Anything, mock.AnythingOfType("*domain.User")).
			Return(domain.ErrConflict).Once()

		body := `{"name":"John Doe","email":"existing@example.com","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
		mockUsecase.AssertExpectations(t)
	})

	t.Run("invalid body", func(t *testing.T) {
		app, _, _ := setupAuthTestApp()
		body := `not valid json`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app, mockUsecase, _ := setupAuthTestApp()
		tokenPair := &domain.TokenPair{
			AccessToken:  "access-token-here",
			RefreshToken: "refresh-token-here",
		}
		mockUsecase.On("Login", mock.Anything, "john@example.com", "password123").
			Return(tokenPair, nil).Once()

		body := `{"email":"john@example.com","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res response.SuccessResponse
		respBody, _ := io.ReadAll(resp.Body)
		json.Unmarshal(respBody, &res)
		assert.True(t, res.Success)

		// Check Set-Cookie header for refresh token
		cookies := resp.Cookies()
		var refreshCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == "refresh_token" {
				refreshCookie = c
			}
		}
		assert.NotNil(t, refreshCookie)
		assert.Equal(t, "refresh-token-here", refreshCookie.Value)
		assert.True(t, refreshCookie.HttpOnly)
		assert.Equal(t, "/api/v1/auth", refreshCookie.Path)

		mockUsecase.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		app, mockUsecase, _ := setupAuthTestApp()
		mockUsecase.On("Login", mock.Anything, "john@example.com", "wrongpassword").
			Return(nil, domain.ErrUnauthorized).Once()

		body := `{"email":"john@example.com","password":"wrongpassword"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		mockUsecase.AssertExpectations(t)
	})

	t.Run("invalid body", func(t *testing.T) {
		app, _, _ := setupAuthTestApp()
		body := `not valid json`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAuthHandler_Refresh(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app, mockUsecase, _ := setupAuthTestApp()
		newPair := &domain.TokenPair{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
		}
		mockUsecase.On("Refresh", mock.Anything, "old-refresh-token").
			Return(newPair, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "old-refresh-token"})

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		cookies := resp.Cookies()
		var refreshCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == "refresh_token" {
				refreshCookie = c
			}
		}
		assert.NotNil(t, refreshCookie)
		assert.Equal(t, "new-refresh-token", refreshCookie.Value)

		mockUsecase.AssertExpectations(t)
	})

	t.Run("no cookie", func(t *testing.T) {
		app, _, _ := setupAuthTestApp()
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("invalid token", func(t *testing.T) {
		app, mockUsecase, _ := setupAuthTestApp()
		mockUsecase.On("Refresh", mock.Anything, "invalid-token").
			Return(nil, domain.ErrUnauthorized).Once()

		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "invalid-token"})

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		mockUsecase.AssertExpectations(t)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app, mockUsecase, _ := setupAuthTestApp()
		mockUsecase.On("Logout", mock.Anything, "some-refresh-token").
			Return(nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "some-refresh-token"})

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		cookies := resp.Cookies()
		var refreshCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == "refresh_token" {
				refreshCookie = c
			}
		}
		assert.NotNil(t, refreshCookie)
		assert.Empty(t, refreshCookie.Value)

		mockUsecase.AssertExpectations(t)
	})

	t.Run("no cookie", func(t *testing.T) {
		app, mockUsecase, _ := setupAuthTestApp()
		mockUsecase.On("Logout", mock.Anything, "").
			Return(nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockUsecase.AssertExpectations(t)
	})
}

func TestAuthHandler_GetCurrent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app, _, mockUserUsecase := setupAuthTestApp()
		user := &domain.User{
			ID:    1,
			Name:  "John Doe",
			Email: "john@example.com",
		}
		mockUserUsecase.On("GetByID", mock.Anything, uint(1)).Return(user, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/auth/current", nil)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockUserUsecase.AssertExpectations(t)
	})
}
