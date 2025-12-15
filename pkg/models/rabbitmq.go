package models

import "time"

type RawEmails struct {
	RawEmail []RawEmail `json:"emails"`
}

type RawEmail struct {
	EmailID   string    `json:"email_id"`
	UserID    string    `json:"user_id"`
	MessageID string    `json:"message_id"`
	From      string    `json:"from_address"`
	Subject   string    `json:"subject"`
	Text      string    `json:"body_text"`
	Date      time.Time `json:"date_received"`
	TimeStamp time.Time `json:"sync_timestamp"`
}

type ParsedEmails struct {
	UserID      string    `json:"user_id"`
	EmailID     string    `json:"email_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Deadline    time.Time `json:"deadline"`
	From        string    `json:"from_address"`
}

