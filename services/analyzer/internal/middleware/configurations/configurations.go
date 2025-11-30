package configurations

import (
	"strings"
	echomiddleware "yfp/services/analyzer/internal/middleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	GzipLevel = 5
	BodyLimit = "4M"
)

func ConfigMiddlewares(e *echo.Echo) {
	e.Use(middleware.Logger())
	e.Use(echomiddleware.CorrelationIdMiddleware)
	e.Use(middleware.RequestID())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: GzipLevel,
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Request().URL.Path, "swagger")
		},
	}))
	e.Use(middleware.BodyLimit(BodyLimit))
}
