package api

import (
	"net/http"
	"strings"

	"core/internal/api/response"
	"core/internal/database"
	"core/internal/security"
	"reminder-hub/pkg/logger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.in/go-playground/validator.v9"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func SetupRoutes(e *echo.Echo, db *database.DB, encryptor security.Encryptor, internalToken string, log *logger.CurrentLogger) {

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.Validator = &CustomValidator{validator: validator.New()}

	handler := NewHandler(db, encryptor, log)

	api := e.Group("/api")

	api.GET("/health", handler.HealthCheck)

	integrations := api.Group("/integrations")
	integrations.Use(InternalAuth(internalToken), UserIDAuth)

	integrations.POST("", handler.CreateIntegration)
	integrations.GET("/:user_id", handler.GetUserIntegrations)
	integrations.DELETE("/:id", handler.DeleteIntegration)
}

func InternalAuth(internalToken string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "Authorization header required"})
			}

			if len(authHeader) < 8 || !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "Invalid authorization format"})
			}

			token := authHeader[7:]
			if token != internalToken {
				return c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "Invalid token"})
			}

			return next(c)
		}
	}
}
