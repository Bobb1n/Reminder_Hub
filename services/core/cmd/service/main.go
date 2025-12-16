package main

import (
	"context"
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
	pkgrabbitmq "reminder-hub/pkg/rabbitmq"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
)

func main() {
	ctx := context.Background()

	
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	appLogger := logger.Init(env)

	cfg := config.Load(appLogger)
	appLogger.Info(ctx, "Configuration loaded")

	db, err := database.NewDB(cfg.DBURL)
	if err != nil {
		appLogger.Fatal(ctx, "Failed to connect to database", "error", err)
	}
	defer db.Close()
	appLogger.Info(ctx, "Database connected")

	migrationPath := "internal/database/migrations"
	absPath, err := filepath.Abs(migrationPath)
	if err != nil {
		appLogger.Fatal(ctx, "Failed to get absolute path for migrations", "error", err)
	}

	appLogger.Info(ctx, "Using migrations", "path", absPath)

	
	migrationURL := cfg.DBURL
	if migrationURL[len(migrationURL)-1] != '?' && migrationURL[len(migrationURL)-1] != '&' {
		if strings.Contains(migrationURL, "?") {
			migrationURL += "&x-migrations-table=core_schema_migrations"
		} else {
			migrationURL += "?x-migrations-table=core_schema_migrations"
		}
	}

	m, err := migrate.New(
		"file://"+absPath,
		migrationURL,
	)
	if err != nil {
		appLogger.Fatal(ctx, "Failed to create migrate instance", "error", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		appLogger.Fatal(ctx, "Failed to run migrations", "error", err)
	}
	appLogger.Info(ctx, "Migrations completed")

	
	rabbitConn, err := pkgrabbitmq.NewRabbitMQConn(cfg.Rabbitmq, ctx, appLogger)
	if err != nil {
		appLogger.Fatal(ctx, "Failed to connect to RabbitMQ", "error", err)
	}
	defer rabbitConn.Close()
	appLogger.Info(ctx, "RabbitMQ connected")

	
	rabbit, err := rabbitmq.NewProducerWithConn(rabbitConn, cfg.Rabbitmq, appLogger, ctx)
	if err != nil {
		appLogger.Fatal(ctx, "Failed to create RabbitMQ producer", "error", err)
	}
	defer rabbit.Close()

	encryptor := security.NewEncryptor(cfg.EncryptionKey)

	syncer := imap.NewSyncer(db, rabbit, encryptor, cfg.IMAPTimeout, appLogger)

	sched := scheduler.NewScheduler(db, syncer, cfg.MaxWorkers, cfg.BatchSize, cfg.SyncInterval, appLogger)
	sched.Start()
	defer sched.Stop()

	e := echo.New()

	api.SetupRoutes(e, db, encryptor, cfg.InternalAPIToken, appLogger)

	go func() {
		if err := e.Start(":" + cfg.ServerPort); err != nil {
			appLogger.Info(ctx, "Server stopped", "error", err)
		}
	}()
	appLogger.Info(ctx, "Server started", "port", cfg.ServerPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	appLogger.Info(ctx, "Shutting down server...")
}
