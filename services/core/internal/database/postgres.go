package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func NewPostgresConnection(dbURL string) (*DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Connected to PostgreSQL")

	return &DB{db}, nil
}

func (db *DB) Init() error {
	// Создание таблицы email_integrations
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS email_integrations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			email_address VARCHAR NOT NULL,
			imap_host VARCHAR NOT NULL,
			imap_port INTEGER NOT NULL,
			use_ssl BOOLEAN NOT NULL DEFAULT true,
			password VARCHAR NOT NULL,
			last_sync_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create email_integrations table: %w", err)
	}

	// Создание таблицы emails_raw
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS emails_raw (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			message_id VARCHAR NOT NULL,
			from_address VARCHAR NOT NULL,
			subject VARCHAR NOT NULL,
			body_text TEXT NOT NULL,
			date_received TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			processed BOOLEAN NOT NULL DEFAULT false,
			UNIQUE(message_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create emails_raw table: %w", err)
	}

	log.Println("✅ Database tables initialized")
	return nil
}
