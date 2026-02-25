package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/azbagas/golang-clean-arch-boilerplate/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type authUsecase struct {
	userRepo      domain.UserRepository
	refreshRepo   domain.RefreshTokenRepository
	jwtSecret     string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewAuthUsecase creates a new AuthUsecase instance.
func NewAuthUsecase(
	userRepo domain.UserRepository,
	refreshRepo domain.RefreshTokenRepository,
	jwtSecret string,
	accessExpiry time.Duration,
	refreshExpiry time.Duration,
) domain.AuthUsecase {
	return &authUsecase{
		userRepo:      userRepo,
		refreshRepo:   refreshRepo,
		jwtSecret:     jwtSecret,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

func (u *authUsecase) Register(ctx context.Context, user *domain.User) error {
	// Check if email already exists
	existing, err := u.userRepo.GetByEmail(ctx, user.Email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return err
	}
	if existing != nil {
		return domain.ErrConflict
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	return u.userRepo.Create(ctx, user)
}

func (u *authUsecase) Login(ctx context.Context, email, password string) (*domain.TokenPair, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, domain.ErrUnauthorized
	}

	return u.generateTokenPair(ctx, user.ID, user.Email)
}

func (u *authUsecase) Refresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	// Look up the refresh token in DB
	storedToken, err := u.refreshRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}

	// Check if expired
	if time.Now().After(storedToken.ExpiresAt) {
		// Delete the expired token
		_ = u.refreshRepo.DeleteByToken(ctx, refreshToken)
		return nil, domain.ErrUnauthorized
	}

	// Delete the old refresh token (rotation)
	if err := u.refreshRepo.DeleteByToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	// Get user to include email in access token claims
	user, err := u.userRepo.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, err
	}

	return u.generateTokenPair(ctx, user.ID, user.Email)
}

func (u *authUsecase) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil // No refresh token to revoke
	}
	err := u.refreshRepo.DeleteByToken(ctx, refreshToken)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return err
	}
	return nil
}

// generateTokenPair creates a new access token + refresh token and stores the refresh token in DB.
func (u *authUsecase) generateTokenPair(ctx context.Context, userID uint, email string) (*domain.TokenPair, error) {
	// Generate access token (JWT)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(u.accessExpiry).Unix(),
		"iat":     time.Now().Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(u.jwtSecret))
	if err != nil {
		return nil, err
	}

	// Generate refresh token (random opaque string)
	refreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		return nil, err
	}
	refreshTokenString := hex.EncodeToString(refreshTokenBytes)

	// Store refresh token in DB
	rt := &domain.RefreshToken{
		UserID:    userID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(u.refreshExpiry),
	}
	if err := u.refreshRepo.Create(ctx, rt); err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}
