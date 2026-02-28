package domain

import (
	"context"
	"time"
)

// User represents the core user entity.
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"type:varchar(255);not null"`
	Email     string    `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"type:varchar(255);not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RefreshToken represents a stored refresh token in the database.
type RefreshToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index;not null"`
	Token     string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
}

// TokenPair holds an access token and a refresh token.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// UserAllowedSortFields defines the fields that users can be sorted by.
var UserAllowedSortFields = map[string]bool{
	"name":       true,
	"email":      true,
	"created_at": true,
}

// UserListParams holds all query parameters for listing users.
type UserListParams struct {
	PaginationParams
	SortParams
	Search string
}

// UserRepository defines the interface for user data access.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetAll(ctx context.Context, params UserListParams) (*PaginatedResult, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uint) error
}

// RefreshTokenRepository defines the interface for refresh token data access.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByToken(ctx context.Context, token string) (*RefreshToken, error)
	DeleteByToken(ctx context.Context, token string) error
}

// UserUsecase defines the interface for user business logic.
type UserUsecase interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uint) (*User, error)
	GetAll(ctx context.Context, params UserListParams) (*PaginatedResult, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uint) error
}

// AuthUsecase defines the interface for authentication business logic.
type AuthUsecase interface {
	Register(ctx context.Context, user *User) error
	Login(ctx context.Context, email, password string) (*TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
}
