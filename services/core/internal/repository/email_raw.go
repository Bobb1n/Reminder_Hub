package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/core/internal/models"
	"github.com/google/uuid"
)

type EmailRawRepository struct {
	db *sql.DB
}

func NewEmailRawRepository(db *sql.DB) *EmailRawRepository {
	return &EmailRawRepository{db: db}
}

func (r *EmailRawRepository) Create(ctx context.Context, email *models.EmailRaw) error {
	query := `
		INSERT INTO emails_raw (
			id, user_id, message_id, from_address, subject, body_text, 
			date_received, created_at, processed
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		email.ID,
		email.UserID,
		email.MessageID,
		email.FromAddress,
		email.Subject,
		email.BodyText,
		email.DateReceived,
		email.CreatedAt,
		email.Processed,
	)

	if err != nil {
		return fmt.Errorf("failed to create raw email: %w", err)
	}

	return nil
}

func (r *EmailRawRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*models.EmailRaw, error) {
	query := `
		SELECT id, user_id, message_id, from_address, subject, body_text,
		       date_received, created_at, processed
		FROM emails_raw 
		WHERE user_id = $1
		ORDER BY date_received DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query raw emails: %w", err)
	}
	defer rows.Close()

	var emails []*models.EmailRaw
	for rows.Next() {
		var email models.EmailRaw

		err := rows.Scan(
			&email.ID,
			&email.UserID,
			&email.MessageID,
			&email.FromAddress,
			&email.Subject,
			&email.BodyText,
			&email.DateReceived,
			&email.CreatedAt,
			&email.Processed,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan raw email: %w", err)
		}

		emails = append(emails, &email)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return emails, nil
}

func (r *EmailRawRepository) FindUnprocessed(ctx context.Context, limit int) ([]*models.EmailRaw, error) {
	query := `
		SELECT id, user_id, message_id, from_address, subject, body_text,
		       date_received, created_at, processed
		FROM emails_raw 
		WHERE processed = false
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query unprocessed emails: %w", err)
	}
	defer rows.Close()

	var emails []*models.EmailRaw
	for rows.Next() {
		var email models.EmailRaw

		err := rows.Scan(
			&email.ID,
			&email.UserID,
			&email.MessageID,
			&email.FromAddress,
			&email.Subject,
			&email.BodyText,
			&email.DateReceived,
			&email.CreatedAt,
			&email.Processed,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan raw email: %w", err)
		}

		emails = append(emails, &email)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return emails, nil
}

func (r *EmailRawRepository) MarkAsProcessed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE emails_raw SET processed = true WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark email as processed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("email not found")
	}

	return nil
}

func (r *EmailRawRepository) ExistsByMessageID(ctx context.Context, messageID string) (bool, error) {
	query := `SELECT COUNT(*) FROM emails_raw WHERE message_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, messageID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}

func (r *EmailRawRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM emails_raw WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user emails: %w", err)
	}

	return nil
}
