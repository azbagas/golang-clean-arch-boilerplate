package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/azbagas/golang-clean-arch-boilerplate/internal/domain"
	"github.com/azbagas/golang-clean-arch-boilerplate/internal/mocks"
	"github.com/azbagas/golang-clean-arch-boilerplate/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserUsecase_Create(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)
	ctx := context.Background()

	user := &domain.User{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.On("Create", ctx, user).Return(nil).Once()

		err := uc.Create(ctx, user)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("conflict", func(t *testing.T) {
		mockRepo.On("Create", ctx, user).Return(domain.ErrConflict).Once()

		err := uc.Create(ctx, user)
		assert.ErrorIs(t, err, domain.ErrConflict)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserUsecase_GetByID(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)
	ctx := context.Background()

	expectedUser := &domain.User{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
	}

	t.Run("found", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, uint(1)).Return(expectedUser, nil).Once()

		user, err := uc.GetByID(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, uint(999)).Return(nil, domain.ErrNotFound).Once()

		user, err := uc.GetByID(ctx, 999)
		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserUsecase_GetAll(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)
	ctx := context.Background()

	expectedUsers := []domain.User{
		{ID: 1, Name: "John Doe", Email: "john@example.com"},
		{ID: 2, Name: "Jane Doe", Email: "jane@example.com"},
	}

	params := domain.PaginationParams{Page: 1, PerPage: 10}
	expectedResult := domain.NewPaginatedResult(expectedUsers, 2, params)

	mockRepo.On("GetAll", ctx, params).Return(expectedResult, nil).Once()

	result, err := uc.GetAll(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	assert.Equal(t, 1, result.Pagination.Page)
	assert.Equal(t, int64(2), result.Pagination.TotalItems)
	mockRepo.AssertExpectations(t)
}

func TestUserUsecase_Update(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)
	ctx := context.Background()

	existingUser := &domain.User{
		ID:        1,
		Name:      "John Doe",
		Email:     "john@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		updateUser := &domain.User{
			ID:    1,
			Name:  "John Updated",
			Email: "john.updated@example.com",
		}

		mockRepo.On("GetByID", ctx, uint(1)).Return(existingUser, nil).Once()
		mockRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(nil).Once()

		err := uc.Update(ctx, updateUser)
		assert.NoError(t, err)
		assert.Equal(t, existingUser.Password, updateUser.Password) // password preserved
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		updateUser := &domain.User{ID: 999, Name: "Nobody"}
		mockRepo.On("GetByID", ctx, uint(999)).Return(nil, domain.ErrNotFound).Once()

		err := uc.Update(ctx, updateUser)
		assert.ErrorIs(t, err, domain.ErrNotFound)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserUsecase_Delete(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		existingUser := &domain.User{ID: 1, Name: "John Doe"}
		mockRepo.On("GetByID", ctx, uint(1)).Return(existingUser, nil).Once()
		mockRepo.On("Delete", ctx, uint(1)).Return(nil).Once()

		err := uc.Delete(ctx, 1)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, uint(999)).Return(nil, domain.ErrNotFound).Once()

		err := uc.Delete(ctx, 999)
		assert.ErrorIs(t, err, domain.ErrNotFound)
		mockRepo.AssertExpectations(t)
	})
}
