package models

import (
	"time"

	"github.com/google/uuid"
)

type EmailIntegration struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	EmailAddress string     `json:"email_address" db:"email_address"`
	IMAPHost     string     `json:"imap_host" db:"imap_host"`
	IMAPPort     int        `json:"imap_port" db:"imap_port"`
	UseSSL       bool       `json:"use_ssl" db:"use_ssl"`
	Password     string     `json:"password" db:"password"`
	LastSyncAt   *time.Time `json:"last_sync_at" db:"last_sync_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateEmailIntegrationRequest struct {
	UserID       uuid.UUID `json:"user_id" binding:"required"`
	EmailAddress string    `json:"email_address" binding:"required,email"`
	IMAPHost     string    `json:"imap_host" binding:"required"`
	IMAPPort     int       `json:"imap_port" binding:"required"`
	UseSSL       bool      `json:"use_ssl"`
	Password     string    `json:"password" binding:"required"`
}
