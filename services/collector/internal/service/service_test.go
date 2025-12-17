package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"collector/internal/database"
)

type stubDB struct{
	created *database.Task
}

func (s *stubDB) TaskExists(ctx context.Context, emailID, userID string) (bool, error) {
	return false, nil
}
func (s *stubDB) CreateTask(ctx context.Context, task *database.Task) error {
	s.created = task
	return nil
}
func (s *stubDB) GetTask(ctx context.Context, taskID, userID string) (*database.Task, error) {
	return nil, nil
}
func (s *stubDB) GetUserTasks(ctx context.Context, filter database.TaskFilter) ([]database.Task, error) {
	return nil, nil
}
func (s *stubDB) UpdateTask(ctx context.Context, taskID, userID string, update database.UpdateTaskRequest) error {
	return nil
}
func (s *stubDB) DeleteTask(ctx context.Context, taskID, userID string) error { return nil }
func (s *stubDB) CompleteTask(ctx context.Context, taskID, userID string) error { return nil }
func (s *stubDB) GetTaskStats(ctx context.Context, userID string) (*database.TaskStats, error) {
	return nil, nil
}

func TestTaskService_HandleEmailMessage_CreatesTask(t *testing.T) {
	db := &stubDB{}
	svc := NewTaskService(db)

	deadline := time.Now().Add(48 * time.Hour).UTC().Truncate(time.Second)
	body, _ := json.Marshal(map[string]any{
		"user_id":    "user-1",
		"email_id":   "email-1",
		"title":      "Title",
		"description": "Desc",
		"deadline":   deadline,
	})

	if err := svc.HandleEmailMessage(context.Background(), body); err != nil {
		t.Fatalf("HandleEmailMessage error: %v", err)
	}
	if db.created == nil {
		t.Fatal("expected task to be created")
	}
	if db.created.UserID != "user-1" || db.created.EmailID != "email-1" {
		t.Fatalf("unexpected task: %+v", db.created)
	}
}
