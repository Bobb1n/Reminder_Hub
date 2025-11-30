package http

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func NewContext() context.Context {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-ctx.Done()
		log.Println("Shutdown signal received, gracefully shutting down...")
		cancel()
	}()
	return ctx
}
