package main

import (
	"log"

	"auth/internal/config"
	"auth/internal/usecase/service"
	"auth/internal/repository/postgres"
	"auth/internal/transport/http"
	postgresDB "auth/pkg/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := postgresDB.New(&cfg.Config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	userRepo := postgres.NewUserRepo(db.Pool)
	blacklistRepo := postgres.NewBlacklistRepo(db.Pool)

	authUsecase := service.NewAuthService(userRepo, blacklistRepo, cfg.JWTSecret)

	server := http.NewServer(cfg.Port, authUsecase)

	if err := server.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
