package database

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

type DBer interface {
	CreateIntegration(ctx context.Context, integration *EmailIntegration) error
	GetUserIntegrations(ctx context.Context, userID string) ([]EmailIntegration, error)
	DeleteIntegration(ctx context.Context, userID, integrationID string) error
	GetIntegrationsForSync(ctx context.Context, limit int) ([]EmailIntegration, error)
	UpdateLastSync(ctx context.Context, integrationID string) error
	EmailExists(ctx context.Context, userID, messageID string) (bool, error)
	SaveEmail(ctx context.Context, email *EmailRaw) error
}

func NewDB(url string) (*DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) CreateIntegration(ctx context.Context, integration *EmailIntegration) error {
	query := `INSERT INTO email_integrations (id, user_id, email_address, imap_host, imap_port, use_ssl, password, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())`
	_, err := db.ExecContext(ctx, query,
		integration.ID, integration.UserID, integration.EmailAddress,
		integration.ImapHost, integration.ImapPort, integration.UseSSL, integration.Password)
	return err
}

func (db *DB) GetUserIntegrations(ctx context.Context, userID string) ([]EmailIntegration, error) {
	query := `SELECT id, user_id, email_address, imap_host, imap_port, use_ssl, created_at, updated_at, last_sync_at
              FROM email_integrations WHERE user_id = $1`
	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var integrations []EmailIntegration
	for rows.Next() {
		var integration EmailIntegration
		err := rows.Scan(&integration.ID, &integration.UserID, &integration.EmailAddress,
			&integration.ImapHost, &integration.ImapPort, &integration.UseSSL,
			&integration.CreatedAt, &integration.UpdatedAt, &integration.LastSyncAt)
		if err != nil {
			return nil, err
		}
		integrations = append(integrations, integration)
	}
	return integrations, nil
}

func (db *DB) DeleteIntegration(ctx context.Context, userID, integrationID string) error {
	query := `DELETE FROM email_integrations WHERE user_id = $1 AND id = $2`
	result, err := db.ExecContext(ctx, query, userID, integrationID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrIntegrationNotFound
	}

	return nil
}

func (db *DB) GetIntegrationsForSync(ctx context.Context, limit int) ([]EmailIntegration, error) {
	query := `SELECT id, user_id, email_address, imap_host, imap_port, use_ssl, password, last_sync_at
              FROM email_integrations
              ORDER BY last_sync_at ASC NULLS FIRST
              LIMIT $1`
	rows, err := db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var integrations []EmailIntegration
	for rows.Next() {
		var integration EmailIntegration
		err := rows.Scan(&integration.ID, &integration.UserID, &integration.EmailAddress,
			&integration.ImapHost, &integration.ImapPort, &integration.UseSSL,
			&integration.Password, &integration.LastSyncAt)
		if err != nil {
			return nil, err
		}
		integrations = append(integrations, integration)
	}
	return integrations, nil
}

func (db *DB) UpdateLastSync(ctx context.Context, integrationID string) error {
	query := `UPDATE email_integrations SET last_sync_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := db.ExecContext(ctx, query, integrationID)
	return err
}

func (db *DB) EmailExists(ctx context.Context, userID, messageID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM emails_raw WHERE user_id = $1 AND message_id = $2)`
	err := db.QueryRowContext(ctx, query, userID, messageID).Scan(&exists)
	return exists, err
}

func (db *DB) SaveEmail(ctx context.Context, email *EmailRaw) error {
	query := `INSERT INTO emails_raw (id, user_id, message_id, from_address, subject, body_text, date_received, processed, created_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`
	_, err := db.ExecContext(ctx, query,
		email.ID, email.UserID, email.MessageID, email.FromAddress,
		email.Subject, email.BodyText, email.DateReceived, email.Processed)
	return err
}
