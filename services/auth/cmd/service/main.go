package main

import (
	"context"

	"auth/internal/config"
	"auth/internal/usecase/service"
	"auth/internal/repository/postgres"
	"auth/internal/transport/http"
	postgresDB "auth/pkg/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		cfg.Logger.Fatal(context.Background(), "Failed to load config", "error", err)
	}

	ctx := context.Background()
	cfg.Logger.Info(ctx, "Loading configuration", "port", cfg.Port)

	db, err := postgresDB.New(&cfg.Config)
	if err != nil {
		cfg.Logger.Fatal(ctx, "Failed to connect to database", "error", err)
	}

	cfg.Logger.Info(ctx, "Database connected successfully")

	userRepo := postgres.NewUserRepo(db.Pool)
	blacklistRepo := postgres.NewBlacklistRepo(db.Pool)

	authUsecase := service.NewAuthService(userRepo, blacklistRepo, cfg.JWTSecret)

	server := http.NewServer(cfg.Port, authUsecase, cfg.Logger)

	if err := server.Start(); err != nil {
		cfg.Logger.Fatal(ctx, "Server failed", "error", err)
	}
}
