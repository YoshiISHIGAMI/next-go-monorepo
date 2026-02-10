package middleware

import (
	"net/http"

	"go-api/internal/apperror"

	"github.com/labstack/echo/v4"
)

// ErrorHandler is a custom error handler that returns standardized error responses
func ErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	// Handle our custom APIError
	if apiErr, ok := err.(*apperror.APIError); ok {
		c.JSON(apiErr.StatusCode, map[string]string{
			"code":    apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}

	// Handle Echo's HTTPError
	if he, ok := err.(*echo.HTTPError); ok {
		code := "ERROR"
		message := "an error occurred"

		switch he.Code {
		case http.StatusBadRequest:
			code = "BAD_REQUEST"
		case http.StatusUnauthorized:
			code = "UNAUTHORIZED"
		case http.StatusForbidden:
			code = "FORBIDDEN"
		case http.StatusNotFound:
			code = "NOT_FOUND"
		case http.StatusConflict:
			code = "CONFLICT"
		case http.StatusInternalServerError:
			code = "INTERNAL_ERROR"
		}

		if m, ok := he.Message.(string); ok {
			message = m
		}

		c.JSON(he.Code, map[string]string{
			"code":    code,
			"message": message,
		})
		return
	}

	// Default to internal server error
	c.JSON(http.StatusInternalServerError, map[string]string{
		"code":    "INTERNAL_ERROR",
		"message": "internal server error",
	})
}
