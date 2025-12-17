package main

import (
	"context"

	"auth/internal/config"
	"auth/internal/repository/postgres"
	"auth/internal/repository/redis"
	"auth/internal/transport/http"
	"auth/internal/usecase/service"
	postgresDB "auth/pkg/postgres"
	redisClient "reminder-hub/pkg/redis"
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

	redisCfg := &redisClient.Config{
		Host:     cfg.RedisHost,
		Port:     cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}

	redisClient, err := redisClient.New(ctx, redisCfg)
	if err != nil {
		cfg.Logger.Fatal(ctx, "Failed to connect to Redis", "error", err)
	}
	defer redisClient.Close()

	cfg.Logger.Info(ctx, "Redis connected successfully")

	userRepo := postgres.NewUserRepo(db.Pool)
	blacklistRepo := redis.NewBlacklistRepo(redisClient.GetClient())

	authUsecase := service.NewAuthService(userRepo, blacklistRepo, cfg.JWTSecret)

	server := http.NewServer(cfg.Port, authUsecase, cfg.Logger)

	if err := server.Start(); err != nil {
		cfg.Logger.Fatal(ctx, "Server failed", "error", err)
	}
}
