package models

import (
	"time"

	"github.com/google/uuid"
)

type EmailRaw struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	MessageID    string    `json:"message_id" db:"message_id"`
	FromAddress  string    `json:"from_address" db:"from_address"`
	Subject      string    `json:"subject" db:"subject"`
	BodyText     string    `json:"body_text" db:"body_text"`
	DateReceived time.Time `json:"date_received" db:"date_received"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	Processed    bool      `json:"processed" db:"processed"`
}

type RabbitMQEmailMessage struct {
	UserID    uuid.UUID `json:"user_id"`
	MessageID string    `json:"message_id"`
	From      string    `json:"from"`
	Subject   string    `json:"subject"`
	Date      time.Time `json:"date"`
	Text      string    `json:"text"`
}
