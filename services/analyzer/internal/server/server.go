package server

import (
	"context"
	"errors"
	"net/http"
	"yfp/internal/logger"
	"yfp/services/analyzer/internal/config"
	"yfp/services/analyzer/internal/server/echoserver"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

func RunServers(lc fx.Lifecycle, log *logger.CurrentLogger, e *echo.Echo, ctx context.Context, cfg *config.Config) error {

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				if err := echoserver.RunEchoServer(ctx, e, log, cfg.Echo); !errors.Is(err, http.ErrServerClosed) {
					log.Fatal(ctx, "error running http server: %v", err)
				}
			}()

			e.GET("/", func(c echo.Context) error {
				return c.String(http.StatusOK, config.GetMicroserviceName(cfg.ServiceName))
			})

			return nil
		},
		OnStop: func(_ context.Context) error {
			log.Info(ctx, "all servers shutdown gracefully...")
			return nil
		},
	})

	return nil
}
