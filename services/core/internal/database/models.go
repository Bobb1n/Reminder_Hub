package database

import (
	"time"
)

type EmailIntegration struct {
	ID           string     `json:"id" validate:"required,uuid"`
	UserID       string     `json:"user_id" validate:"required,uuid"`
	EmailAddress string     `json:"email_address" validate:"required,email"`
	ImapHost     string     `json:"imap_host" validate:"required,hostname"`
	ImapPort     int        `json:"imap_port" validate:"required,min=1,max=65535"`
	UseSSL       bool       `json:"use_ssl"`
	Password     string     `json:"-" validate:"required,min=1"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastSyncAt   *time.Time `json:"last_sync_at,omitempty"`
}

type EmailRaw struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	MessageID    string    `json:"message_id"`
	FromAddress  string    `json:"from_address"`
	Subject      string    `json:"subject"`
	BodyText     string    `json:"body_text"`
	DateReceived time.Time `json:"date_received"`
	Processed    bool      `json:"processed"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateIntegrationRequest struct {
	EmailAddress string `json:"email_address" validate:"required,email"`
	ImapHost     string `json:"imap_host" validate:"required,hostname"`
	ImapPort     int    `json:"imap_port" validate:"required,min=1,max=65535"`
	UseSSL       bool   `json:"use_ssl"`
	Password     string `json:"password" validate:"required,min=1"`
}

type CreateIntegrationResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
