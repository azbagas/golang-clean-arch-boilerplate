package dto

// CreateUserRequest represents the request body for creating a user.
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=255" example:"John Doe"`
	Email    string `json:"email" validate:"required,email" example:"john@example.com"`
	Password string `json:"password" validate:"required,min=6,max=128" example:"password123"`
}

// UpdateUserRequest represents the request body for updating a user.
type UpdateUserRequest struct {
	Name  string `json:"name" validate:"omitempty,min=2,max=255" example:"John Doe"`
	Email string `json:"email" validate:"omitempty,email" example:"john@example.com"`
}

// UserResponse represents the response body for a user (excludes password).
type UserResponse struct {
	ID        uint   `json:"id" example:"1"`
	Name      string `json:"name" example:"John Doe"`
	Email     string `json:"email" example:"john@example.com"`
	CreatedAt string `json:"created_at" example:"2026-01-01T00:00:00Z"`
	UpdatedAt string `json:"updated_at" example:"2026-01-01T00:00:00Z"`
}

// DeleteUserResponse represents the response body for a successful user deletion.
type DeleteUserResponse struct {
	Message string `json:"message" example:"User deleted successfully"`
}