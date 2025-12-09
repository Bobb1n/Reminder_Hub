package database

import (
	"time"
)

type Task struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	EmailID     string     `json:"email_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type ParsedEmail struct {
	UserID      string    `json:"user_id"`
	EmailID     string    `json:"email_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Deadline    time.Time `json:"deadline"`
}

type CreateTaskRequest struct {
	UserID      string     `json:"user_id" validate:"required,uuid"`
	EmailID     string     `json:"email_id" validate:"required,uuid"`
	Title       string     `json:"title" validate:"required,min=1,max=500"`
	Description string     `json:"description"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	Priority    string     `json:"priority" validate:"required,oneof=low medium high urgent"`
}

type UpdateTaskRequest struct {
	Title       *string    `json:"title,omitempty" validate:"omitempty,min=1,max=500"`
	Description *string    `json:"description,omitempty"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	Status      *string    `json:"status,omitempty" validate:"omitempty,oneof=pending in_progress completed cancelled archived"`
	Priority    *string    `json:"priority,omitempty" validate:"omitempty,oneof=low medium high urgent"`
}

type TaskFilter struct {
	UserID       string
	Status       *string
	Priority     *string
	FromDeadline *time.Time
	ToDeadline   *time.Time
	Limit        int
	Offset       int
}

type TaskStats struct {
	TotalTasks     int `json:"total_tasks"`
	PendingTasks   int `json:"pending_tasks"`
	CompletedTasks int `json:"completed_tasks"`
	OverdueTasks   int `json:"overdue_tasks"`
	TodayTasks     int `json:"today_tasks"`
	ThisWeekTasks  int `json:"this_week_tasks"`
}
