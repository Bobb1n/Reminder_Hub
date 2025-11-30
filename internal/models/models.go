package models

import "time"

type ParsedEmails struct {
	UserID      string    `json:"user_id"`
	RawEmailID  string    `json:"raw_email_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Deadline    time.Time `json:"deadline"`
	SourceType  string    `json:"source_type"`
}

type RawEmails struct {
	UserID    string `json:"user_id"`
	Subject   string `json:"subject"`
	From      string `json:"from"`
	Date      string `json:"date"`
	Text      string `json:"text"`
	MessageID string `json:"message_id"`
}
