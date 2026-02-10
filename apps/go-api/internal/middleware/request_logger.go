package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

// RequestLogger logs HTTP requests in JSON format
func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			req := c.Request()
			res := c.Response()

			slog.Info("request",
				"method", req.Method,
				"path", req.URL.Path,
				"status", res.Status,
				"latency_ms", time.Since(start).Milliseconds(),
				"remote_ip", c.RealIP(),
				"user_agent", req.UserAgent(),
			)

			return err
		}
	}
}
