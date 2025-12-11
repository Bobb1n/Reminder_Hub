package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/collector/internal/api"
	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/collector/internal/config"
	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/collector/internal/database"
	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/collector/internal/rabbitmq"
	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/collector/internal/service"
	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/collector/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

func main() {
	logger.Init()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load conig")
	}
	log.Info().Msg("Configuration loaded")

	db, err := database.NewDB(cfg.DBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()
	log.Info().Msg("Database connected")

	migrationPath := "internal/database/migrations"
	absPath, err := filepath.Abs(migrationPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get absolute path for migrations")
	}

	log.Info().Msgf("Using migrations from: %s", absPath)

	m, err := migrate.New(
		"file://"+absPath,
		cfg.DBURL,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create migrate instance")
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}
	log.Info().Msg("Migrations completed")

	service := service.NewTaskService(db)
	rabbit, err := rabbitmq.NewConsumer(cfg.RabbitURL, cfg.QueueName, service.HandleEmailMessage)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to RabbitMQ")
	}
	defer rabbit.Close()
	log.Info().Msg("RabbitMQ connected")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := rabbit.Start(ctx); err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("RabbitMQ consumer stopped with error")
		}
	}()
	log.Info().Msg("RabbitMQ consumer started")

	e := echo.New()

	api.SetupRoutes(e, service, cfg.InternalAPIToken)

	go func() {
		if err := e.Start(":" + cfg.ServerPort); err != nil {
			log.Info().Err(err).Msg("Server stopped")
		}
	}()
	log.Info().Msgf("Server started on port %s", cfg.ServerPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
}
