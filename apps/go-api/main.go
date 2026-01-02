package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type PingResponse struct {
	Message string `json:"message"`
	From    string `json:"from"`
}
type HealthResponse struct {
	Status string `json:"status"`
}
type User struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

var jwtSecret []byte

func generateJWT(user User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ユーザー作成（サインアップ）用ハンドラ
func handleUserSignup(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req CreateUserRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid request body",
			})
		}

		if strings.TrimSpace(req.Email) == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "email is required",
			})
		}

		if strings.TrimSpace(req.Password) == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "password is required",
			})
		}

		// パスワードをハッシュ化
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Println("failed to hash password:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to hash password",
			})
		}

		var user User
		err = db.QueryRow(
			`INSERT INTO users (email, password_hash)
             VALUES ($1, $2)
             RETURNING id, email, created_at`,
			req.Email,
			string(hashed),
		).Scan(&user.ID, &user.Email, &user.CreatedAt)
		if err != nil {
			log.Println("failed to insert user:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to create user",
			})
		}

		return c.JSON(http.StatusCreated, user)
	}
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// コンテキストに載せる認証済みユーザー情報
type AuthUser struct {
	ID    int64
	Email string
}

// JWT を検証し、ユーザー情報をコンテキストに載せるミドルウェア
func requireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "missing or invalid Authorization header",
			})
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			// 予期しない署名アルゴリズムを弾く
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid or expired token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid token claims",
			})
		}

		// sub は JSON 上は数値なので float64 経由で取り出す
		sub, ok := claims["sub"].(float64)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid subject",
			})
		}
		email, _ := claims["email"].(string)

		user := AuthUser{
			ID:    int64(sub),
			Email: email,
		}

		// コンテキストに保存
		c.Set("authUser", user)

		return next(c)
	}
}

// コンテキストから認証済みユーザーを取り出すヘルパー
func getAuthUser(c echo.Context) (AuthUser, bool) {
	v := c.Get("authUser")
	if v == nil {
		return AuthUser{}, false
	}
	user, ok := v.(AuthUser)
	return user, ok
}

func findUserByID(db *sql.DB, id int64) (User, error) {
	var u User
	err := db.QueryRow(
		`SELECT id, email, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.CreatedAt)
	return u, err
}

func main() {
	jwtSecretStr := os.Getenv("JWT_SECRET")
	if strings.TrimSpace(jwtSecretStr) == "" {
		log.Fatal("JWT_SECRET is required")
	}
	jwtSecret = []byte(jwtSecretStr)

	dsn := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(dsn) == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal("failed to open database:", err)
	}

	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	e := echo.New()

	e.GET("/ping", func(c echo.Context) error {
		return c.JSON(http.StatusOK, PingResponse{
			Message: "pong",
			From:    "go-api",
		})
	})

	e.GET("/ping/:name", func(c echo.Context) error {
		name := c.Param("name")
		if name == "" {
			name = "unknown"
		}
		return c.JSON(http.StatusOK, PingResponse{
			Message: "pong ",
			From:    name,
		})
	})

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, HealthResponse{
			Status: "ok",
		})
	})

	signupHandler := handleUserSignup(db)

	e.POST("/users", signupHandler)       // 既存のエンドポイント
	e.POST("/auth/signup", signupHandler) // 認証用のサインアップAPI

	// GET /users: ユーザー一覧取得
	e.GET("/users", func(c echo.Context) error {
		rows, err := db.Query("SELECT id, email, created_at FROM users ORDER BY id")
		if err != nil {
			log.Println("failed to query users:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch users"})
		}
		defer rows.Close()

		var users []User
		for rows.Next() {
			var user User
			if err := rows.Scan(&user.ID, &user.Email, &user.CreatedAt); err != nil {
				log.Println("failed to scan user:", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch users"})
			}
			users = append(users, user)
		}

		return c.JSON(http.StatusOK, users)
	})

	// JWT 発行のデモ用エンドポイント
	e.GET("/auth/token-demo", func(c echo.Context) error {
		// 本当は DB からユーザーを取るが、
		// まずは「JWTを発行できるか」のデモなので固定値でOK
		user := User{
			ID:    1,
			Email: "demo@example.com",
		}

		token, err := generateJWT(user)
		if err != nil {
			log.Println("failed to generate jwt:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to generate token",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"token": token,
		})
	})

	// POST /auth/login: email + password でログインして JWT を発行
	e.POST("/auth/login", func(c echo.Context) error {
		var req LoginRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid request body",
			})
		}

		if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "email and password are required",
			})
		}

		var user User
		var passwordHash string

		err := db.QueryRow(
			`SELECT id, email, password_hash, created_at
             FROM users
             WHERE email = $1`,
			req.Email,
		).Scan(&user.ID, &user.Email, &passwordHash, &user.CreatedAt)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// ユーザーが見つからない → 認証失敗
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid email or password",
			})
		case err != nil:
			log.Println("failed to query user:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to login",
			})
		}

		// パスワードを照合
		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
			// ハッシュと一致しない → 認証失敗
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid email or password",
			})
		}

		// JWT を発行
		token, err := generateJWT(user)
		if err != nil {
			log.Println("failed to generate jwt:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to generate token",
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"token": token,
			"user":  user,
		})
	})

	// GET /auth/me: トークンから自分の情報を返す
	e.GET("/auth/me", func(c echo.Context) error {
		au, ok := getAuthUser(c)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		u, err := findUserByID(db, au.ID)
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "user not found"})
		}
		if err != nil {
			log.Println("failed to query user:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		}

		return c.JSON(http.StatusOK, u)
	}, requireAuth)

	// GET /me/profile: 認証が必要な保護ルートの例
	e.GET("/me/profile", func(c echo.Context) error {
		au, ok := getAuthUser(c)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		u, err := findUserByID(db, au.ID)
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "user not found"})
		}
		if err != nil {
			log.Println("failed to query user:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"id":    u.ID,
			"email": u.Email,
			"bio":   "This is a sample profile.",
		})
	}, requireAuth)

	// サーバー起動
	e.Logger.Fatal(e.Start(":8080"))
}
