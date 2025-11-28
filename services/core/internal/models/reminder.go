package models

import (
    "time"
    
    "github.com/google/uuid"
)

type Reminder struct {
    ID          uuid.UUID  `json:"id" db:"id"`
    UserID      uuid.UUID  `json:"user_id" db:"user_id"`
    SourceType  string     `json:"source_type" db:"source_type"` // "manual" или "email"
    EmailRawID  *uuid.UUID `json:"email_raw_id,omitempty" db:"email_raw_id"`
    Title       string     `json:"title" db:"title"`
    Description *string    `json:"description,omitempty" db:"description"`
    Deadline    *time.Time `json:"deadline,omitempty" db:"deadline"`
    IsCompleted bool       `json:"is_completed" db:"is_completed"`
    CreatedAt   time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateReminderRequest struct {
    UserID      uuid.UUID  `json:"user_id" binding:"required"`
    SourceType  string     `json:"source_type" binding:"required"`
    EmailRawID  *uuid.UUID `json:"email_raw_id"`
    Title       string     `json:"title" binding:"required"`
    Description *string    `json:"description"`
    Deadline    *time.Time `json:"deadline"`
}