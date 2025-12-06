package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const ContextKeyUserID = "user_id"

func UserIDAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Request().Header.Get("X-User-ID")
		if userID == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "X-User-ID header required")
		}

		if !isValidUUID(userID) {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID format")
		}

		c.Set(ContextKeyUserID, userID)
		return next(c)
	}
}

func isValidUUID(u string) bool {
	if len(u) != 36 {
		return false
	}

	for i, r := range u {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if r != '-' {
				return false
			}
		} else if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}

	return true
}
