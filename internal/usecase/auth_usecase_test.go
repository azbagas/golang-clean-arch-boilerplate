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
	"golang.org/x/crypto/bcrypt"
)

func newAuthUsecase(userRepo *mocks.MockUserRepository, refreshRepo *mocks.MockRefreshTokenRepository) domain.AuthUsecase {
	return usecase.NewAuthUsecase(userRepo, refreshRepo, "test-secret", 15*time.Minute, 7*24*time.Hour)
}

func TestAuthUsecase_Register(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockRefreshRepo := new(mocks.MockRefreshTokenRepository)
		uc := newAuthUsecase(mockUserRepo, mockRefreshRepo)
		ctx := context.Background()

		user := &domain.User{
			Name:     "John Doe",
			Email:    "john@example.com",
			Password: "password123",
		}

		mockUserRepo.On("GetByEmail", ctx, "john@example.com").Return(nil, domain.ErrNotFound).Once()
		mockUserRepo.On("Create", ctx, user).Return(nil).Once()

		err := uc.Register(ctx, user)
		assert.NoError(t, err)
		assert.NotEqual(t, "password123", user.Password) // password was hashed
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("duplicate email", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockRefreshRepo := new(mocks.MockRefreshTokenRepository)
		uc := newAuthUsecase(mockUserRepo, mockRefreshRepo)
		ctx := context.Background()

		user := &domain.User{
			Name:     "Jane Doe",
			Email:    "existing@example.com",
			Password: "password123",
		}

		existingUser := &domain.User{ID: 1, Email: "existing@example.com"}
		mockUserRepo.On("GetByEmail", ctx, "existing@example.com").Return(existingUser, nil).Once()

		err := uc.Register(ctx, user)
		assert.ErrorIs(t, err, domain.ErrConflict)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthUsecase_Login(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	t.Run("success", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockRefreshRepo := new(mocks.MockRefreshTokenRepository)
		uc := newAuthUsecase(mockUserRepo, mockRefreshRepo)
		ctx := context.Background()

		existingUser := &domain.User{
			ID:       1,
			Email:    "john@example.com",
			Password: string(hashedPassword),
		}

		mockUserRepo.On("GetByEmail", ctx, "john@example.com").Return(existingUser, nil).Once()
		mockRefreshRepo.On("Create", ctx, mock.AnythingOfType("*domain.RefreshToken")).Return(nil).Once()

		tokenPair, err := uc.Login(ctx, "john@example.com", "password123")
		assert.NoError(t, err)
		assert.NotEmpty(t, tokenPair.AccessToken)
		assert.NotEmpty(t, tokenPair.RefreshToken)
		mockUserRepo.AssertExpectations(t)
		mockRefreshRepo.AssertExpectations(t)
	})

	t.Run("wrong password", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockRefreshRepo := new(mocks.MockRefreshTokenRepository)
		uc := newAuthUsecase(mockUserRepo, mockRefreshRepo)
		ctx := context.Background()

		existingUser := &domain.User{
			ID:       1,
			Email:    "john@example.com",
			Password: string(hashedPassword),
		}

		mockUserRepo.On("GetByEmail", ctx, "john@example.com").Return(existingUser, nil).Once()

		tokenPair, err := uc.Login(ctx, "john@example.com", "wrongpassword")
		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.Nil(t, tokenPair)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockRefreshRepo := new(mocks.MockRefreshTokenRepository)
		uc := newAuthUsecase(mockUserRepo, mockRefreshRepo)
		ctx := context.Background()

		mockUserRepo.On("GetByEmail", ctx, "notfound@example.com").Return(nil, domain.ErrNotFound).Once()

		tokenPair, err := uc.Login(ctx, "notfound@example.com", "password123")
		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.Nil(t, tokenPair)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthUsecase_Refresh(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockRefreshRepo := new(mocks.MockRefreshTokenRepository)
		uc := newAuthUsecase(mockUserRepo, mockRefreshRepo)
		ctx := context.Background()

		storedToken := &domain.RefreshToken{
			ID:        1,
			UserID:    1,
			Token:     "valid-refresh-token",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		user := &domain.User{ID: 1, Email: "john@example.com"}

		mockRefreshRepo.On("GetByToken", ctx, "valid-refresh-token").Return(storedToken, nil).Once()
		mockRefreshRepo.On("DeleteByToken", ctx, "valid-refresh-token").Return(nil).Once()
		mockUserRepo.On("GetByID", ctx, uint(1)).Return(user, nil).Once()
		mockRefreshRepo.On("Create", ctx, mock.AnythingOfType("*domain.RefreshToken")).Return(nil).Once()

		tokenPair, err := uc.Refresh(ctx, "valid-refresh-token")
		assert.NoError(t, err)
		assert.NotEmpty(t, tokenPair.AccessToken)
		assert.NotEmpty(t, tokenPair.RefreshToken)
		assert.NotEqual(t, "valid-refresh-token", tokenPair.RefreshToken) // rotated
		mockRefreshRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("invalid token", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockRefreshRepo := new(mocks.MockRefreshTokenRepository)
		uc := newAuthUsecase(mockUserRepo, mockRefreshRepo)
		ctx := context.Background()

		mockRefreshRepo.On("GetByToken", ctx, "invalid-token").Return(nil, domain.ErrNotFound).Once()

		tokenPair, err := uc.Refresh(ctx, "invalid-token")
		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.Nil(t, tokenPair)
		mockRefreshRepo.AssertExpectations(t)
	})

	t.Run("expired token", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockRefreshRepo := new(mocks.MockRefreshTokenRepository)
		uc := newAuthUsecase(mockUserRepo, mockRefreshRepo)
		ctx := context.Background()

		storedToken := &domain.RefreshToken{
			ID:        1,
			UserID:    1,
			Token:     "expired-token",
			ExpiresAt: time.Now().Add(-1 * time.Hour), // expired
		}

		mockRefreshRepo.On("GetByToken", ctx, "expired-token").Return(storedToken, nil).Once()
		mockRefreshRepo.On("DeleteByToken", ctx, "expired-token").Return(nil).Once()

		tokenPair, err := uc.Refresh(ctx, "expired-token")
		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.Nil(t, tokenPair)
		mockRefreshRepo.AssertExpectations(t)
	})
}

func TestAuthUsecase_Logout(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockRefreshRepo := new(mocks.MockRefreshTokenRepository)
		uc := newAuthUsecase(mockUserRepo, mockRefreshRepo)
		ctx := context.Background()

		mockRefreshRepo.On("DeleteByToken", ctx, "some-refresh-token").Return(nil).Once()

		err := uc.Logout(ctx, "some-refresh-token")
		assert.NoError(t, err)
		mockRefreshRepo.AssertExpectations(t)
	})

	t.Run("empty token", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockRefreshRepo := new(mocks.MockRefreshTokenRepository)
		uc := newAuthUsecase(mockUserRepo, mockRefreshRepo)
		ctx := context.Background()

		err := uc.Logout(ctx, "")
		assert.NoError(t, err) // no-op, no error
	})
}
