package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func AuthMiddleware(authServiceURL string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Path() == "/health" || strings.HasPrefix(c.Path(), "/auth/") {
				return next(c)
			}

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Authorization header required")
			}
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authorization format")
			}

			token := parts[1]
			userID, err := validateToken(authServiceURL, token)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
			}
			c.Set("user_id", userID)
			return next(c)
		}
	}
}

func validateToken(authServiceURL, token string) (string, error) {
	// Реализация проверки токена через Auth Service POST /auth/validate с токеном. В реализации здесь будет HTTP запрос к Auth Service
	return "user-uuid", nil
}
