package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type PingResponse struct {
	Message string `json:"message"`
	From    string `json:"go-api"`
}
type HealthResponse struct {
	Status string `json:"status"`
}

func main() {
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

	e.Logger.Fatal(e.Start(":8080"))
}
