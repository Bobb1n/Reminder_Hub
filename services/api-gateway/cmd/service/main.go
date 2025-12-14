package main

import (
	"log"

	"api-gateway/internal/config"
	auth "api-gateway/internal/middleware"
	"api-gateway/internal/proxy"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg := config.Load()

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
	}))

	e.Use(auth.AuthMiddleware(cfg.AuthServiceURL))

	authProxy, err := proxy.AuthProxy(cfg.AuthServiceURL)
	if err != nil {
		log.Fatalf("Failed to create auth proxy: %v", err)
	}

	coreProxy, err := proxy.NewServiceProxy(cfg.CoreServiceURL, cfg.InternalToken)
	if err != nil {
		log.Fatalf("Failed to create core proxy: %v", err)
	}

	collectorProxy, err := proxy.NewServiceProxy(cfg.CollectorServiceURL, cfg.InternalToken)
	if err != nil {
		log.Fatalf("Failed to create collector proxy: %v", err)
	}

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"status":  "healthy",
			"service": "api-gateway",
		})
	})

	authGroup := e.Group("/auth")
	authGroup.Any("/*", authProxy)

	api := e.Group("/api/v1")
	{
		// Применяем AutoIMAPMiddleware только к integrations
		integrations := api.Group("/integrations/email")
		integrations.Use(auth.AutoIMAPMiddleware()) // Добавляем middleware
		integrations.Any("/*", coreProxy.Proxy)

		reminders := api.Group("/reminders")
		reminders.Any("/*", coreProxy.Proxy)
	}

	internal := e.Group("/internal")
	internal.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			providedToken := c.Request().Header.Get("X-Internal-Token")
			if providedToken != cfg.InternalToken {
				return echo.NewHTTPError(403, "Forbidden - internal access only")
			}
			return next(c)
		}
	})
	{
		internal.Any("/reminders/*", collectorProxy.Proxy)
	}

	log.Printf("Starting API Gateway on port %s", cfg.ServerPort)
	if err := e.Start(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start gateway: %v", err)
	}
}
