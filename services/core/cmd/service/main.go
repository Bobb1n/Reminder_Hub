package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"core/internal/api"
	"core/internal/config"
	"core/internal/database"
	"core/internal/imap"
	"core/internal/logger"
	"core/internal/rabbitmq"
	"core/internal/security"
	scheduler "core/internal/sheduler"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

func main() {
	logger.Init()

	cfg := config.Load()
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

	rabbit, err := rabbitmq.NewProducer(cfg.RabbitURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to RabbitMQ")
	}
	defer rabbit.Close()
	log.Info().Msg("RabbitMQ connected")

	encryptor := security.NewEncryptor(cfg.EncryptionKey)

	syncer := imap.NewSyncer(db, rabbit, encryptor, cfg.IMAPTimeout)

	sched := scheduler.NewScheduler(db, syncer, cfg.MaxWorkers, cfg.BatchSize, cfg.SyncInterval)
	sched.Start()
	defer sched.Stop()

	e := echo.New()

	api.SetupRoutes(e, db, encryptor, cfg.InternalAPIToken)

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
