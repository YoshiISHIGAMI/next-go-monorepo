package handler

import (
	"database/sql"
	"log/slog"
	"net/http"

	"go-api/internal/apperror"
	"go-api/internal/model"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	db *sql.DB
}

func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{db: db}
}

// List returns all users
func (h *UserHandler) List(c echo.Context) error {
	rows, err := h.db.Query("SELECT id, email, name, created_at FROM users ORDER BY id")
	if err != nil {
		slog.Error("failed to query users", "error", err)
		return apperror.Internal("failed to fetch users")
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt); err != nil {
			slog.Error("failed to scan user", "error", err)
			return apperror.Internal("failed to fetch users")
		}
		users = append(users, user)
	}

	if users == nil {
		users = []model.User{}
	}

	return c.JSON(http.StatusOK, users)
}
