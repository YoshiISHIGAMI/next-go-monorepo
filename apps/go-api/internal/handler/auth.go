package handler

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"go-api/internal/apperror"
	"go-api/internal/middleware"
	"go-api/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordLen = 8
	maxPasswordLen = 72
)

type AuthHandler struct {
	db        *sql.DB
	jwtSecret []byte
}

func NewAuthHandler(db *sql.DB, jwtSecret []byte) *AuthHandler {
	return &AuthHandler{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

// Signup creates a new user with email and password
func (h *AuthHandler) Signup(c echo.Context) error {
	var req model.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return apperror.BadRequest("invalid request body")
	}

	req.Email = normalizeEmail(req.Email)

	if err := validateEmail(req.Email); err != nil {
		return apperror.BadRequest(err.Error())
	}
	if err := validatePassword(req.Password); err != nil {
		return apperror.BadRequest(err.Error())
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("failed to hash password", "error", err)
		return apperror.Internal("failed to hash password")
	}

	var user model.User
	err = h.db.QueryRow(
		`INSERT INTO users (email, password_hash)
		 VALUES ($1, $2)
		 RETURNING id, email, created_at`,
		req.Email, string(hashed),
	).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperror.EmailAlreadyExists()
		}
		slog.Error("failed to insert user", "error", err)
		return apperror.Internal("failed to create user")
	}

	return c.JSON(http.StatusCreated, user)
}

// Login authenticates user and returns JWT
func (h *AuthHandler) Login(c echo.Context) error {
	var req model.LoginRequest
	if err := c.Bind(&req); err != nil {
		return apperror.BadRequest("invalid request body")
	}

	req.Email = normalizeEmail(req.Email)

	if err := validateEmail(req.Email); err != nil {
		return apperror.BadRequest(err.Error())
	}
	if err := validatePassword(req.Password); err != nil {
		return apperror.BadRequest(err.Error())
	}

	var user model.User
	var passwordHash string

	err := h.db.QueryRow(
		`SELECT id, email, password_hash, created_at
		 FROM users WHERE email = $1`,
		req.Email,
	).Scan(&user.ID, &user.Email, &passwordHash, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return apperror.InvalidCredentials()
	}
	if err != nil {
		slog.Error("failed to query user", "error", err)
		return apperror.Internal("failed to login")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		return apperror.InvalidCredentials()
	}

	token, err := h.generateJWT(user)
	if err != nil {
		slog.Error("failed to generate jwt", "error", err)
		return apperror.Internal("failed to generate token")
	}

	return c.JSON(http.StatusOK, model.LoginResponse{
		Token: token,
		User:  user,
	})
}

// OAuthCallback handles OAuth provider callback
func (h *AuthHandler) OAuthCallback(c echo.Context) error {
	var req model.OAuthCallbackRequest
	if err := c.Bind(&req); err != nil {
		return apperror.BadRequest("invalid request body")
	}

	if req.Provider == "" || req.ProviderAccountID == "" {
		return apperror.BadRequest("provider and provider_account_id are required")
	}

	req.Email = normalizeEmail(req.Email)

	// Check if auth_identity exists
	var userID int64
	err := h.db.QueryRow(
		`SELECT user_id FROM auth_identities
		 WHERE provider = $1 AND provider_account_id = $2`,
		req.Provider, req.ProviderAccountID,
	).Scan(&userID)

	if err == nil {
		user, err := h.findUserByID(userID)
		if err != nil {
			slog.Error("failed to find user", "error", err)
			return apperror.Internal("failed to fetch user")
		}
		return c.JSON(http.StatusOK, model.OAuthCallbackResponse{
			User:      user,
			IsNewUser: false,
		})
	}

	if !errors.Is(err, sql.ErrNoRows) {
		slog.Error("failed to query auth_identity", "error", err)
		return apperror.Internal("failed to check auth identity")
	}

	// Create new user
	tx, err := h.db.Begin()
	if err != nil {
		slog.Error("failed to begin transaction", "error", err)
		return apperror.Internal("failed to create user")
	}
	defer tx.Rollback()

	var user model.User
	var name *string
	if req.Name != "" {
		name = &req.Name
	}

	err = tx.QueryRow(
		`INSERT INTO users (email, name)
		 VALUES ($1, $2)
		 ON CONFLICT (email) DO UPDATE SET name = COALESCE(EXCLUDED.name, users.name)
		 RETURNING id, email, name, created_at`,
		req.Email, name,
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
	if err != nil {
		slog.Error("failed to insert user", "error", err)
		return apperror.Internal("failed to create user")
	}

	_, err = tx.Exec(
		`INSERT INTO auth_identities (user_id, provider, provider_account_id)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (provider, provider_account_id) DO NOTHING`,
		user.ID, req.Provider, req.ProviderAccountID,
	)
	if err != nil {
		slog.Error("failed to insert auth_identity", "error", err)
		return apperror.Internal("failed to create auth identity")
	}

	if err := tx.Commit(); err != nil {
		slog.Error("failed to commit transaction", "error", err)
		return apperror.Internal("failed to create user")
	}

	return c.JSON(http.StatusCreated, model.OAuthCallbackResponse{
		User:      user,
		IsNewUser: true,
	})
}

// Me returns current authenticated user
func (h *AuthHandler) Me(c echo.Context) error {
	au, ok := middleware.GetAuthUser(c)
	if !ok {
		return apperror.Unauthorized("")
	}

	user, err := h.findUserByID(au.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return apperror.NotFound("user")
	}
	if err != nil {
		slog.Error("failed to query user", "error", err)
		return apperror.Internal("failed to fetch user")
	}

	return c.JSON(http.StatusOK, user)
}

// Profile returns user profile (demo endpoint)
func (h *AuthHandler) Profile(c echo.Context) error {
	au, ok := middleware.GetAuthUser(c)
	if !ok {
		return apperror.Unauthorized("")
	}

	user, err := h.findUserByID(au.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return apperror.NotFound("user")
	}
	if err != nil {
		slog.Error("failed to query user", "error", err)
		return apperror.Internal("failed to fetch user")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
		"bio":   "This is a sample profile.",
	})
}

// TokenDemo generates a demo JWT (for testing)
func (h *AuthHandler) TokenDemo(c echo.Context) error {
	user := model.User{
		ID:    1,
		Email: "demo@example.com",
	}

	token, err := h.generateJWT(user)
	if err != nil {
		slog.Error("failed to generate jwt", "error", err)
		return apperror.Internal("failed to generate token")
	}

	return c.JSON(http.StatusOK, model.TokenResponse{Token: token})
}

// Helper functions

func (h *AuthHandler) generateJWT(user model.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}

func (h *AuthHandler) findUserByID(id int64) (model.User, error) {
	var u model.User
	err := h.db.QueryRow(
		`SELECT id, email, name, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt)
	return u, err
}

func normalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func validateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return errors.New("invalid email")
	}
	return nil
}

func validatePassword(pw string) error {
	pw = strings.TrimSpace(pw)
	if pw == "" {
		return errors.New("password is required")
	}
	if len([]byte(pw)) < minPasswordLen {
		return errors.New("password must be at least 8 characters")
	}
	if len([]byte(pw)) > maxPasswordLen {
		return errors.New("password must be at most 72 characters")
	}
	return nil
}
