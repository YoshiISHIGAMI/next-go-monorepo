package apperror

import (
	"fmt"
	"net/http"
)

// APIError represents a standardized API error response
type APIError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// New creates a custom API error
func New(statusCode int, code, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

// Common errors

func BadRequest(message string) *APIError {
	return &APIError{
		StatusCode: http.StatusBadRequest,
		Code:       "BAD_REQUEST",
		Message:    message,
	}
}

func Unauthorized(message string) *APIError {
	if message == "" {
		message = "unauthorized"
	}
	return &APIError{
		StatusCode: http.StatusUnauthorized,
		Code:       "UNAUTHORIZED",
		Message:    message,
	}
}

func Forbidden(message string) *APIError {
	if message == "" {
		message = "forbidden"
	}
	return &APIError{
		StatusCode: http.StatusForbidden,
		Code:       "FORBIDDEN",
		Message:    message,
	}
}

func NotFound(resource string) *APIError {
	message := "not found"
	if resource != "" {
		message = resource + " not found"
	}
	return &APIError{
		StatusCode: http.StatusNotFound,
		Code:       "NOT_FOUND",
		Message:    message,
	}
}

func Conflict(message string) *APIError {
	return &APIError{
		StatusCode: http.StatusConflict,
		Code:       "CONFLICT",
		Message:    message,
	}
}

func Internal(message string) *APIError {
	if message == "" {
		message = "internal server error"
	}
	return &APIError{
		StatusCode: http.StatusInternalServerError,
		Code:       "INTERNAL_ERROR",
		Message:    message,
	}
}

// Domain-specific errors

func EmailAlreadyExists() *APIError {
	return &APIError{
		StatusCode: http.StatusConflict,
		Code:       "EMAIL_EXISTS",
		Message:    "email already exists",
	}
}

func InvalidCredentials() *APIError {
	return &APIError{
		StatusCode: http.StatusUnauthorized,
		Code:       "INVALID_CREDENTIALS",
		Message:    "invalid email or password",
	}
}
