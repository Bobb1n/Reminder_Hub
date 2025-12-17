package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"collector/internal/database"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDB struct {
	mock.Mock
}

func (m *mockDB) CreateTask(ctx context.Context, task *database.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *mockDB) GetTask(ctx context.Context, taskID, userID string) (*database.Task, error) {
	args := m.Called(ctx, taskID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.Task), args.Error(1)
}

func (m *mockDB) GetUserTasks(ctx context.Context, filter database.TaskFilter) ([]database.Task, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]database.Task), args.Error(1)
}

func (m *mockDB) UpdateTask(ctx context.Context, taskID, userID string, update database.UpdateTaskRequest) error {
	args := m.Called(ctx, taskID, userID, update)
	return args.Error(0)
}

func (m *mockDB) DeleteTask(ctx context.Context, taskID, userID string) error {
	args := m.Called(ctx, taskID, userID)
	return args.Error(0)
}

func (m *mockDB) CompleteTask(ctx context.Context, taskID, userID string) error {
	args := m.Called(ctx, taskID, userID)
	return args.Error(0)
}

func (m *mockDB) GetTaskStats(ctx context.Context, userID string) (*database.TaskStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.TaskStats), args.Error(1)
}

func (m *mockDB) TaskExists(ctx context.Context, emailID, userID string) (bool, error) {
	args := m.Called(ctx, emailID, userID)
	return args.Bool(0), args.Error(1)
}

func TestTaskService_DeterminePriority_Urgent(t *testing.T) {
	service := NewTaskService(new(mockDB))
	
	deadline := time.Now().Add(12 * time.Hour) // Меньше 1 дня
	priority := service.determinePriority(deadline)
	
	assert.Equal(t, "urgent", priority)
}

func TestTaskService_DeterminePriority_High(t *testing.T) {
	service := NewTaskService(new(mockDB))
	
	deadline := time.Now().Add(2 * 24 * time.Hour) // 2 дня
	priority := service.determinePriority(deadline)
	
	assert.Equal(t, "high", priority)
}

func TestTaskService_DeterminePriority_Medium(t *testing.T) {
	service := NewTaskService(new(mockDB))
	
	deadline := time.Now().Add(5 * 24 * time.Hour) // 5 дней
	priority := service.determinePriority(deadline)
	
	assert.Equal(t, "medium", priority)
}

func TestTaskService_DeterminePriority_Low(t *testing.T) {
	service := NewTaskService(new(mockDB))
	
	deadline := time.Now().Add(10 * 24 * time.Hour) // 10 дней
	priority := service.determinePriority(deadline)
	
	assert.Equal(t, "low", priority)
}

func TestTaskService_HandleEmailMessage_TaskExists(t *testing.T) {
	mockDB := new(mockDB)
	service := NewTaskService(mockDB)
	
	emailData := map[string]interface{}{
		"user_id":     uuid.New().String(),
		"email_id":    uuid.New().String(),
		"title":       "Test Task",
		"description": "Test Description",
		"deadline":    time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	}
	
	body, _ := json.Marshal(emailData)
	
	mockDB.On("TaskExists", mock.Anything, emailData["email_id"].(string), emailData["user_id"].(string)).
		Return(true, nil)
	
	err := service.HandleEmailMessage(context.Background(), body)
	
	assert.NoError(t, err)
	mockDB.AssertNotCalled(t, "CreateTask")
	mockDB.AssertExpectations(t)
}

func TestTaskService_HandleEmailMessage_InvalidJSON(t *testing.T) {
	service := NewTaskService(new(mockDB))
	
	invalidBody := []byte("invalid json")
	
	err := service.HandleEmailMessage(context.Background(), invalidBody)
	
	assert.Error(t, err)
}

func TestTaskService_GetTask(t *testing.T) {
	mockDB := new(mockDB)
	service := NewTaskService(mockDB)
	
	taskID := uuid.New().String()
	userID := uuid.New().String()
	task := &database.Task{
		ID:     taskID,
		UserID: userID,
	}
	
	mockDB.On("GetTask", mock.Anything, taskID, userID).Return(task, nil)
	
	result, err := service.GetTask(context.Background(), taskID, userID)
	
	assert.NoError(t, err)
	assert.Equal(t, task, result)
	mockDB.AssertExpectations(t)
}

func TestTaskService_GetUserTasks(t *testing.T) {
	mockDB := new(mockDB)
	service := NewTaskService(mockDB)
	
	userID := uuid.New().String()
	filter := database.TaskFilter{
		UserID: userID,
		Limit:  10,
	}
	tasks := []database.Task{
		{ID: uuid.New().String(), UserID: userID},
	}
	
	mockDB.On("GetUserTasks", mock.Anything, filter).Return(tasks, nil)
	
	result, err := service.GetUserTasks(context.Background(), userID, filter)
	
	assert.NoError(t, err)
	assert.Equal(t, tasks, result)
	mockDB.AssertExpectations(t)
}

func TestNewTaskService(t *testing.T) {
	mockDB := new(mockDB)
	service := NewTaskService(mockDB)
	
	assert.NotNil(t, service)
	assert.Equal(t, mockDB, service.db)
}
