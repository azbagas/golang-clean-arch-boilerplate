package postgres

import (
	"context"
	"errors"

	"github.com/azbagas/golang-clean-arch-boilerplate/internal/domain"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository instance.
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return domain.ErrConflict
		}
		return result.Error
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	result := r.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *userRepository) GetAll(ctx context.Context, params domain.UserListParams) (*domain.PaginatedResult, error) {
	var users []domain.User
	var totalItems int64

	query := r.db.WithContext(ctx).Model(&domain.User{})

	// Apply search filter
	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ?", search, search)
	}

	// Count total items (with filters applied)
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, err
	}

	// Apply sorting
	if params.SortBy != "" {
		query = query.Order(params.SortBy + " " + params.SortOrder)
	}

	// Apply pagination
	result := query.
		Offset(params.GetOffset()).
		Limit(params.PerPage).
		Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	return domain.NewPaginatedResult(users, totalItems, params.PaginationParams), nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&domain.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
