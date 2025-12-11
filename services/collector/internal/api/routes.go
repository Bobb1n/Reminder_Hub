package api

import (
	"net/http"
	"strings"

	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/collector/internal/api/response"
	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/collector/internal/service"
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

func SetupRoutes(e *echo.Echo, taskService *service.TaskService, internalToken string) {

	e.Validator = &CustomValidator{validator: validator.New()}

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	taskHandler := NewTaskHandler(taskService)

	e.GET("/health", taskHandler.HealthCheck)

	api := e.Group("/api/v1", InternalAuth(internalToken))
	api.Use(UserIDAuth)
	{

		api.GET("/tasks", taskHandler.GetTasks)
		api.GET("/tasks/:id", taskHandler.GetTask)
		api.PUT("/tasks/:id", taskHandler.UpdateTask)
		api.DELETE("/tasks/:id", taskHandler.DeleteTask)
		api.POST("/tasks/:id/complete", taskHandler.CompleteTask)
		api.GET("/tasks/stats", taskHandler.GetStats)
	}
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
