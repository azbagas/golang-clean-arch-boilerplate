package domain

import "errors"

// Domain-specific sentinel errors.
var (
	ErrNotFound       = errors.New("resource not found")
	ErrConflict       = errors.New("resource already exists")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrBadRequest     = errors.New("bad request")
	ErrInternalServer = errors.New("internal server error")
)
