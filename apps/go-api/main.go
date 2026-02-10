package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"go-api/internal/handler"
	"go-api/internal/logger"
	"go-api/internal/middleware"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v4"
)

func main() {
	// Initialize structured logging
	logger.Setup()

	// Load configuration
	jwtSecret := mustEnv("JWT_SECRET")
	dsn := mustEnv("DATABASE_URL")

	// Connect to database
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	slog.Info("connected to database")

	// Initialize Echo
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = middleware.ErrorHandler

	// Middleware
	e.Use(middleware.RequestLogger())

	// Initialize handlers
	authHandler := handler.NewAuthHandler(db, []byte(jwtSecret))
	userHandler := handler.NewUserHandler(db)

	// Auth middleware
	authMiddleware := middleware.RequireAuth([]byte(jwtSecret))

	// Routes
	setupRoutes(e, authHandler, userHandler, authMiddleware)

	// Start server
	port := os.Getenv("PORT")
	if strings.TrimSpace(port) == "" {
		port = "8080"
	}

	slog.Info("starting server", "port", port)
	if err := e.Start(":" + port); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func setupRoutes(
	e *echo.Echo,
	auth *handler.AuthHandler,
	user *handler.UserHandler,
	authMiddleware echo.MiddlewareFunc,
) {
	// Health check
	e.GET("/ping", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "pong"})
	})

	// Auth
	e.POST("/auth/signup", auth.Signup)
	e.POST("/auth/login", auth.Login)
	e.POST("/auth/oauth/callback", auth.OAuthCallback)
	e.GET("/auth/token-demo", auth.TokenDemo)
	e.GET("/auth/me", auth.Me, authMiddleware)

	// Users
	e.POST("/users", auth.Signup) // Alias for signup
	e.GET("/users", user.List)

	// Protected routes
	e.GET("/me/profile", auth.Profile, authMiddleware)
}

func mustEnv(key string) string {
	value := os.Getenv(key)
	if strings.TrimSpace(value) == "" {
		slog.Error("required environment variable not set", "key", key)
		os.Exit(1)
	}
	return value
}
