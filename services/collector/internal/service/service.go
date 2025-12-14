package service

import (
	"context"
	"encoding/json"
	"time"

	"collector/internal/database"
	"collector/internal/util"
)

type TaskService struct {
	db database.DBer
}

func NewTaskService(db database.DBer) *TaskService {
	return &TaskService{db: db}
}

func (s *TaskService) HandleEmailMessage(ctx context.Context, body []byte) error {
	var emailData struct {
		UserID      string    `json:"user_id"`
		EmailID     string    `json:"email_id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Deadline    time.Time `json:"deadline"`
	}

	if err := json.Unmarshal(body, &emailData); err != nil {
		return err
	}

	exists, err := s.db.TaskExists(ctx, emailData.EmailID, emailData.UserID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	task := &database.Task{
		ID:          util.GenerateUUID(),
		UserID:      emailData.UserID,
		EmailID:     emailData.EmailID,
		Title:       emailData.Title,
		Description: emailData.Description,
		Deadline:    &emailData.Deadline,
		Status:      "pending",
		Priority:    s.determinePriority(emailData.Deadline),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return s.db.CreateTask(ctx, task)
}

func (s *TaskService) determinePriority(deadline time.Time) string {
	now := time.Now()
	daysUntil := int(deadline.Sub(now).Hours() / 24)

	switch {
	case daysUntil <= 1:
		return "urgent"
	case daysUntil <= 3:
		return "high"
	case daysUntil <= 7:
		return "medium"
	default:
		return "low"
	}
}

func (s *TaskService) GetTask(ctx context.Context, taskID, userID string) (*database.Task, error) {
	return s.db.GetTask(ctx, taskID, userID)
}

func (s *TaskService) GetUserTasks(ctx context.Context, userID string, filter database.TaskFilter) ([]database.Task, error) {
	filter.UserID = userID
	return s.db.GetUserTasks(ctx, filter)
}

func (s *TaskService) UpdateTask(ctx context.Context, taskID, userID string, update database.UpdateTaskRequest) error {
	return s.db.UpdateTask(ctx, taskID, userID, update)
}

func (s *TaskService) DeleteTask(ctx context.Context, taskID, userID string) error {
	return s.db.DeleteTask(ctx, taskID, userID)
}

func (s *TaskService) CompleteTask(ctx context.Context, taskID, userID string) error {
	return s.db.CompleteTask(ctx, taskID, userID)
}

func (s *TaskService) GetStats(ctx context.Context, userID string) (*database.TaskStats, error) {
	return s.db.GetTaskStats(ctx, userID)
}
