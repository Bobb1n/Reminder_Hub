package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/core/internal/models"
	"github.com/google/uuid"
)

type ReminderRepository struct {
	db *sql.DB
}

func NewReminderRepository(db *sql.DB) *ReminderRepository {
	return &ReminderRepository{db: db}
}

func (r *ReminderRepository) Create(ctx context.Context, reminder *models.Reminder) error {
	query := `
		INSERT INTO reminders (
			id, user_id, source_type, email_raw_id, title, description,
			deadline, is_completed, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		reminder.ID,
		reminder.UserID,
		reminder.SourceType,
		reminder.EmailRawID,
		reminder.Title,
		reminder.Description,
		reminder.Deadline,
		reminder.IsCompleted,
		reminder.CreatedAt,
		reminder.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create reminder: %w", err)
	}

	return nil
}

func (r *ReminderRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Reminder, error) {
	query := `
		SELECT id, user_id, source_type, email_raw_id, title, description,
		       deadline, is_completed, created_at, updated_at
		FROM reminders 
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query reminders: %w", err)
	}
	defer rows.Close()

	var reminders []*models.Reminder
	for rows.Next() {
		var reminder models.Reminder
		var emailRawID sql.NullString
		var description sql.NullString
		var deadline sql.NullTime

		err := rows.Scan(
			&reminder.ID,
			&reminder.UserID,
			&reminder.SourceType,
			&emailRawID,
			&reminder.Title,
			&description,
			&deadline,
			&reminder.IsCompleted,
			&reminder.CreatedAt,
			&reminder.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reminder: %w", err)
		}

		if emailRawID.Valid {
			parsedID, err := uuid.Parse(emailRawID.String)
			if err == nil {
				reminder.EmailRawID = &parsedID
			}
		}

		if description.Valid {
			reminder.Description = &description.String
		}

		if deadline.Valid {
			reminder.Deadline = &deadline.Time
		}

		reminders = append(reminders, &reminder)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return reminders, nil
}

func (r *ReminderRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Reminder, error) {
	query := `
		SELECT id, user_id, source_type, email_raw_id, title, description,
		       deadline, is_completed, created_at, updated_at
		FROM reminders 
		WHERE id = $1
	`

	var reminder models.Reminder
	var emailRawID sql.NullString
	var description sql.NullString
	var deadline sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&reminder.ID,
		&reminder.UserID,
		&reminder.SourceType,
		&emailRawID,
		&reminder.Title,
		&description,
		&deadline,
		&reminder.IsCompleted,
		&reminder.CreatedAt,
		&reminder.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query reminder: %w", err)
	}

	if emailRawID.Valid {
		parsedID, err := uuid.Parse(emailRawID.String)
		if err == nil {
			reminder.EmailRawID = &parsedID
		}
	}

	if description.Valid {
		reminder.Description = &description.String
	}

	if deadline.Valid {
		reminder.Deadline = &deadline.Time
	}

	return &reminder, nil
}

func (r *ReminderRepository) Update(ctx context.Context, reminder *models.Reminder) error {
	query := `
		UPDATE reminders 
		SET title = $1, description = $2, deadline = $3, is_completed = $4,
		    updated_at = $5
		WHERE id = $6
	`

	_, err := r.db.ExecContext(ctx, query,
		reminder.Title,
		reminder.Description,
		reminder.Deadline,
		reminder.IsCompleted,
		time.Now(),
		reminder.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update reminder: %w", err)
	}

	return nil
}

func (r *ReminderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM reminders WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete reminder: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("reminder not found")
	}

	return nil
}

func (r *ReminderRepository) FindByEmailRawID(ctx context.Context, emailRawID uuid.UUID) (*models.Reminder, error) {
	query := `
		SELECT id, user_id, source_type, email_raw_id, title, description,
		       deadline, is_completed, created_at, updated_at
		FROM reminders 
		WHERE email_raw_id = $1
	`

	var reminder models.Reminder
	var description sql.NullString
	var deadline sql.NullTime

	err := r.db.QueryRowContext(ctx, query, emailRawID).Scan(
		&reminder.ID,
		&reminder.UserID,
		&reminder.SourceType,
		&reminder.EmailRawID,
		&reminder.Title,
		&description,
		&deadline,
		&reminder.IsCompleted,
		&reminder.CreatedAt,
		&reminder.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query reminder: %w", err)
	}

	if description.Valid {
		reminder.Description = &description.String
	}

	if deadline.Valid {
		reminder.Deadline = &deadline.Time
	}

	return &reminder, nil
}
