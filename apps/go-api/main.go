package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v4"
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
type CreateUserRequest struct {
	Email string `json:"email"`
}

func main() {
	dsn := "postgres://nextgo:nextgo@localhost:5432/nextgo_dev?sslmode=disable"

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

	// POST /users: ユーザーを1件作成
	e.POST("/users", func(c echo.Context) error {
		var req CreateUserRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		}

		// とりあえず password_hash はダミー。後で bcrypt に差し替える
		const dummyPasswordHash = "not-implemented-yet"

		var user User
		err := db.QueryRow(
			"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email, created_at",
			req.Email,
			dummyPasswordHash,
		).Scan(&user.ID, &user.Email, &user.CreatedAt)
		if err != nil {
			log.Println("failed to insert user:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
		}

		return c.JSON(http.StatusCreated, user)
	})

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

	// サーバー起動
	e.Logger.Fatal(e.Start(":8080"))
}
