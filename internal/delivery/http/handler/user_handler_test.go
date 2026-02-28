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

func setupUserTestApp() (*fiber.App, *mocks.MockUserUsecase) {
	app := fiber.New()
	mockUsecase := new(mocks.MockUserUsecase)
	h := handler.NewUserHandler(mockUsecase)

	app.Post("/users", h.Create)
	app.Get("/users", h.GetAll)
	app.Get("/users/:id", h.GetByID)
	app.Put("/users/:id", h.Update)
	app.Delete("/users/:id", h.Delete)

	return app, mockUsecase
}

func TestUserHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app, mockUsecase := setupUserTestApp()
		mockUsecase.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
			Return(nil).Once()

		body := `{"name":"John Doe","email":"john@example.com","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte(body)))
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

	t.Run("invalid body", func(t *testing.T) {
		app, _ := setupUserTestApp()
		body := `not json`
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestUserHandler_GetAll(t *testing.T) {
	t.Run("success with default pagination", func(t *testing.T) {
		app, mockUsecase := setupUserTestApp()

		users := []domain.User{
			{ID: 1, Name: "John", Email: "john@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: 2, Name: "Jane", Email: "jane@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}

		expectedParams := domain.PaginationParams{Page: 1, PerPage: 10}
		paginatedResult := domain.NewPaginatedResult(users, 2, expectedParams)

		mockUsecase.On("GetAll", mock.Anything, expectedParams).Return(paginatedResult, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res response.PaginatedResponse
		respBody, _ := io.ReadAll(resp.Body)
		json.Unmarshal(respBody, &res)
		assert.True(t, res.Success)
		assert.Equal(t, 1, res.Pagination.Page)
		assert.Equal(t, 10, res.Pagination.PerPage)
		assert.Equal(t, int64(2), res.Pagination.TotalItems)
		assert.Equal(t, 1, res.Pagination.TotalPages)
		assert.False(t, res.Pagination.HasNext)
		assert.False(t, res.Pagination.HasPrev)
		mockUsecase.AssertExpectations(t)
	})

	t.Run("success with custom pagination", func(t *testing.T) {
		app, mockUsecase := setupUserTestApp()

		users := []domain.User{
			{ID: 3, Name: "Alice", Email: "alice@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}

		expectedParams := domain.PaginationParams{Page: 2, PerPage: 5}
		paginatedResult := domain.NewPaginatedResult(users, 6, expectedParams)

		mockUsecase.On("GetAll", mock.Anything, expectedParams).Return(paginatedResult, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/users?page=2&page_size=5", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res response.PaginatedResponse
		respBody, _ := io.ReadAll(resp.Body)
		json.Unmarshal(respBody, &res)
		assert.True(t, res.Success)
		assert.Equal(t, 2, res.Pagination.Page)
		assert.Equal(t, 5, res.Pagination.PerPage)
		assert.Equal(t, int64(6), res.Pagination.TotalItems)
		assert.Equal(t, 2, res.Pagination.TotalPages)
		assert.False(t, res.Pagination.HasNext)
		assert.True(t, res.Pagination.HasPrev)
		mockUsecase.AssertExpectations(t)
	})
}

func TestUserHandler_GetByID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		app, mockUsecase := setupUserTestApp()
		user := &domain.User{ID: 1, Name: "John", Email: "john@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()}
		mockUsecase.On("GetByID", mock.Anything, uint(1)).Return(user, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockUsecase.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		app, mockUsecase := setupUserTestApp()
		mockUsecase.On("GetByID", mock.Anything, uint(999)).Return(nil, domain.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/users/999", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockUsecase.AssertExpectations(t)
	})

	t.Run("invalid id returns bad request", func(t *testing.T) {
		app, _ := setupUserTestApp()
		// Fiber routes "abc" as a param value, strconv.ParseUint will fail
		req := httptest.NewRequest(http.MethodGet, "/users/abc", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestUserHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app, mockUsecase := setupUserTestApp()
		mockUsecase.On("Delete", mock.Anything, uint(1)).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockUsecase.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		app, mockUsecase := setupUserTestApp()
		mockUsecase.On("Delete", mock.Anything, uint(999)).Return(domain.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodDelete, "/users/999", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockUsecase.AssertExpectations(t)
	})
}
