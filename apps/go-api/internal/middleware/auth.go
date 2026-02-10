package middleware

import (
	"fmt"
	"strings"

	"go-api/internal/apperror"
	"go-api/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const AuthUserKey = "authUser"

// RequireAuth validates JWT and sets user info in context
func RequireAuth(jwtSecret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return apperror.Unauthorized("missing or invalid Authorization header")
			}

			tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return jwtSecret, nil
			})
			if err != nil || !token.Valid {
				return apperror.Unauthorized("invalid or expired token")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return apperror.Unauthorized("invalid token claims")
			}

			sub, ok := claims["sub"].(float64)
			if !ok {
				return apperror.Unauthorized("invalid subject")
			}
			email, _ := claims["email"].(string)

			user := model.AuthUser{
				ID:    int64(sub),
				Email: email,
			}

			c.Set(AuthUserKey, user)
			return next(c)
		}
	}
}

// GetAuthUser retrieves authenticated user from context
func GetAuthUser(c echo.Context) (model.AuthUser, bool) {
	v := c.Get(AuthUserKey)
	if v == nil {
		return model.AuthUser{}, false
	}
	user, ok := v.(model.AuthUser)
	return user, ok
}
