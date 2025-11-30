package echoserver

import (
	"context"
	"time"
	"yfp/internal/logger"

	"github.com/labstack/echo/v4"
)

const (
	ReadTimeout    = 15 * time.Second
	WriteTimeout   = 15 * time.Second
	MaxHeaderBytes = 1 << 20
)

type EchoConfig struct {
	Port               string `env:"PORT" env-default:"5000"`
	BasePath           string `env:"PATH" env-default:"/analyzer/v1" `
	DebugErrorResponse bool   `env:"DEBUG" env-default:"true"`
	Timeout            int    `env:"TIMEOUT" env-default:"30"`
	Host               string `env:"HOST" env-default:"http://localhost"`
}

func NewEchoServer() *echo.Echo {
	e := echo.New()
	return e
}

func RunEchoServer(ctx context.Context, echo *echo.Echo, log logger.CurrentLogger, cfg *EchoConfig) error {
	echo.Server.ReadTimeout = ReadTimeout
	echo.Server.WriteTimeout = WriteTimeout
	echo.Server.MaxHeaderBytes = MaxHeaderBytes
	savedCtx := ctx
	go func() {
		<-ctx.Done()
		log.Info(savedCtx, "shutting down Http PORT: {%s}", cfg.Port)
		err := echo.Shutdown(ctx)
		if err != nil {
			log.Error(savedCtx, "Shutdown error {%v}", err)
			return
		}
		log.Info(savedCtx, "server exited properly")
	}()

	err := echo.Start(cfg.Host + ":" + cfg.Port)

	return err
}
