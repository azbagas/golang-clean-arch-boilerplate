package dto

// RegisterRequest represents the request body for user registration.
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=255" example:"John Doe"`
	Email    string `json:"email" validate:"required,email" example:"john@example.com"`
	Password string `json:"password" validate:"required,min=6,max=128" example:"password123"`
}

// RegisterResponse represents the response body for a successful registration.
type RegisterResponse struct {
	Message string `json:"message" example:"User registered successfully"`
}

// LoginRequest represents the request body for user login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"john@example.com"`
	Password string `json:"password" validate:"required" example:"password123"`
}

// LoginResponse represents the response body for a successful login.
// The refresh token is NOT included here — it is set as an HttpOnly cookie.
type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

// LogoutResponse represents the response body for a successful logout.
type LogoutResponse struct {
	Message string `json:"message" example:"User logged out successfully"`
}