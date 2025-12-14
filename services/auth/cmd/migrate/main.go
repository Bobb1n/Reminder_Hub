package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	username := getEnv("POSTGRES_USER", "postgres")
	password := getEnv("POSTGRES_PASSWORD", "postgres")
	host := getEnv("POSTGRES_HOST", "postgres")
	port := getEnv("POSTGRES_PORT", "5432")
	dbName := getEnv("POSTGRES_DB", "reminderhub")

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		username, password, host, port, dbName)

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Failed to close database: %v", closeErr)
		}
	}()

	ctx := context.Background()
	if pingErr := db.PingContext(ctx); pingErr != nil {
		log.Fatalf("Failed to ping database: %v", pingErr)
	}

	migrationsDir := "/app/migrations"

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		log.Fatalf("Failed to read migrations: %v", err)
	}

	if len(files) == 0 {
		log.Println("No migration files found")
		return
	}

	sort.Strings(files)

	log.Printf("Found %d migration files", len(files))

	createMigrationTable(db)

	for _, file := range files {
		filename := filepath.Base(file)
		if err := runMigration(db, filename, file); err != nil {
			log.Fatalf("Migration %s failed: %v", filename, err)
		}
	}

	log.Println("All migrations completed successfully")
}

func createMigrationTable(db *sql.DB) {
	ctx := context.Background()
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP NOT NULL DEFAULT NOW()
	);
	`
	if _, err := db.ExecContext(ctx, query); err != nil {
		log.Fatalf("Failed to create migration table: %v", err)
	}
}

func runMigration(db *sql.DB, filename, migrationPath string) error {
	ctx := context.Background()
	var count int
	err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM schema_migrations WHERE version = $1",
		filename,
	).Scan(&count)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check migration status: %w", err)
	}

	if count > 0 {
		log.Printf("Migration %s already applied, skipping", filename)
		return nil
	}

	sqlContent, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("read migration file: %w", err)
	}

	sqlQuery := string(sqlContent)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && err == nil {
			log.Printf("Failed to rollback transaction: %v", rollbackErr)
		}
	}()

	queries := strings.Split(sqlQuery, ";")
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" || strings.HasPrefix(query, "--") {
			continue
		}
		if _, execErr := tx.ExecContext(ctx, query); execErr != nil {
			return fmt.Errorf("execute query: %w", execErr)
		}
	}

	if _, err := tx.ExecContext(ctx,
		"INSERT INTO schema_migrations (version) VALUES ($1)",
		filename,
	); err != nil {
		return fmt.Errorf("record migration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	log.Printf("Migration %s applied successfully", filename)
	return nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
