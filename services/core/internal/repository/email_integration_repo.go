package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/core/internal/models"
	"github.com/google/uuid"
)

type EmailIntegrationRepository struct {
	db *sql.DB
}

func NewEmailIntegrationRepository(db *sql.DB) *EmailIntegrationRepository {
	return &EmailIntegrationRepository{db: db}
}

func (r *EmailIntegrationRepository) Create(ctx context.Context, integration *models.EmailIntegration) error {
	query := `
		INSERT INTO email_integrations (
			id, user_id, email_address, imap_host, imap_port, use_ssl, 
			password, last_sync_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		integration.ID,
		integration.UserID,
		integration.EmailAddress,
		integration.IMAPHost,
		integration.IMAPPort,
		integration.UseSSL,
		integration.Password,
		integration.LastSyncAt,
		integration.CreatedAt,
		integration.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create email integration: %w", err)
	}

	return nil
}

func (r *EmailIntegrationRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.EmailIntegration, error) {
	query := `
		SELECT id, user_id, email_address, imap_host, imap_port, use_ssl,
		       password, last_sync_at, created_at, updated_at
		FROM email_integrations 
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query email integrations: %w", err)
	}
	defer rows.Close()

	var integrations []*models.EmailIntegration
	for rows.Next() {
		var integration models.EmailIntegration
		var lastSyncAt sql.NullTime

		err := rows.Scan(
			&integration.ID,
			&integration.UserID,
			&integration.EmailAddress,
			&integration.IMAPHost,
			&integration.IMAPPort,
			&integration.UseSSL,
			&integration.Password,
			&lastSyncAt,
			&integration.CreatedAt,
			&integration.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email integration: %w", err)
		}

		if lastSyncAt.Valid {
			integration.LastSyncAt = &lastSyncAt.Time
		}

		integrations = append(integrations, &integration)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return integrations, nil
}

func (r *EmailIntegrationRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.EmailIntegration, error) {
	query := `
		SELECT id, user_id, email_address, imap_host, imap_port, use_ssl,
		       password, last_sync_at, created_at, updated_at
		FROM email_integrations 
		WHERE id = $1
	`

	var integration models.EmailIntegration
	var lastSyncAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&integration.ID,
		&integration.UserID,
		&integration.EmailAddress,
		&integration.IMAPHost,
		&integration.IMAPPort,
		&integration.UseSSL,
		&integration.Password,
		&lastSyncAt,
		&integration.CreatedAt,
		&integration.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query email integration: %w", err)
	}

	if lastSyncAt.Valid {
		integration.LastSyncAt = &lastSyncAt.Time
	}

	return &integration, nil
}

func (r *EmailIntegrationRepository) Update(ctx context.Context, integration *models.EmailIntegration) error {
	query := `
		UPDATE email_integrations 
		SET email_address = $1, imap_host = $2, imap_port = $3, use_ssl = $4,
		    password = $5, last_sync_at = $6, updated_at = $7
		WHERE id = $8
	`

	_, err := r.db.ExecContext(ctx, query,
		integration.EmailAddress,
		integration.IMAPHost,
		integration.IMAPPort,
		integration.UseSSL,
		integration.Password,
		integration.LastSyncAt,
		time.Now(),
		integration.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update email integration: %w", err)
	}

	return nil
}

func (r *EmailIntegrationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM email_integrations WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete email integration: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("email integration not found")
	}

	return nil
}

func (r *EmailIntegrationRepository) UpdateLastSync(ctx context.Context, id uuid.UUID, syncTime time.Time) error {
	query := `UPDATE email_integrations SET last_sync_at = $1, updated_at = $2 WHERE id = $3`

	_, err := r.db.ExecContext(ctx, query, syncTime, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update last sync time: %w", err)
	}

	return nil
}

// GetAllActiveIntegrations возвращает все активные интеграции для polling
func (r *EmailIntegrationRepository) GetAllActiveIntegrations(ctx context.Context) ([]*models.EmailIntegration, error) {
	query := `
		SELECT id, user_id, email_address, imap_host, imap_port, use_ssl,
		       password, last_sync_at, created_at, updated_at
		FROM email_integrations 
		ORDER BY last_sync_at NULLS FIRST, created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all email integrations: %w", err)
	}
	defer rows.Close()

	var integrations []*models.EmailIntegration
	for rows.Next() {
		var integration models.EmailIntegration
		var lastSyncAt sql.NullTime

		err := rows.Scan(
			&integration.ID,
			&integration.UserID,
			&integration.EmailAddress,
			&integration.IMAPHost,
			&integration.IMAPPort,
			&integration.UseSSL,
			&integration.Password,
			&lastSyncAt,
			&integration.CreatedAt,
			&integration.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email integration: %w", err)
		}

		if lastSyncAt.Valid {
			integration.LastSyncAt = &lastSyncAt.Time
		}

		integrations = append(integrations, &integration)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return integrations, nil
}
