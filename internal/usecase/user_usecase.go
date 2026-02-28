package usecase

import (
	"context"

	"github.com/azbagas/golang-clean-arch-boilerplate/internal/domain"
)

type userUsecase struct {
	userRepo domain.UserRepository
}

// NewUserUsecase creates a new UserUsecase instance.
func NewUserUsecase(userRepo domain.UserRepository) domain.UserUsecase {
	return &userUsecase{userRepo: userRepo}
}

func (u *userUsecase) Create(ctx context.Context, user *domain.User) error {
	return u.userRepo.Create(ctx, user)
}

func (u *userUsecase) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	return u.userRepo.GetByID(ctx, id)
}

func (u *userUsecase) GetAll(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedResult, error) {
	return u.userRepo.GetAll(ctx, params)
}

func (u *userUsecase) Update(ctx context.Context, user *domain.User) error {
	// Verify the user exists first
	existing, err := u.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return err
	}

	// Preserve fields that shouldn't be updated directly
	user.CreatedAt = existing.CreatedAt
	if user.Password == "" {
		user.Password = existing.Password
	}

	return u.userRepo.Update(ctx, user)
}

func (u *userUsecase) Delete(ctx context.Context, id uint) error {
	// Verify the user exists first
	_, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return u.userRepo.Delete(ctx, id)
}
