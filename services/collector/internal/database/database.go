package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

type DBer interface {
	CreateTask(ctx context.Context, task *Task) error
	GetTask(ctx context.Context, taskID, userID string) (*Task, error)
	GetUserTasks(ctx context.Context, filter TaskFilter) ([]Task, error)
	UpdateTask(ctx context.Context, taskID, userID string, update UpdateTaskRequest) error
	DeleteTask(ctx context.Context, taskID, userID string) error
	CompleteTask(ctx context.Context, taskID, userID string) error
	GetTaskStats(ctx context.Context, userID string) (*TaskStats, error)
	TaskExists(ctx context.Context, emailID, userID string) (bool, error)
}

func NewDB(url string) (*DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) CreateTask(ctx context.Context, task *Task) error {
	query := `INSERT INTO tasks (id, user_id, email_id, title, description, deadline, status, priority, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())`

	_, err := db.ExecContext(ctx, query,
		task.ID, task.UserID, task.EmailID, task.Title,
		task.Description, task.Deadline, task.Status, task.Priority)
	return err
}

func (db *DB) GetTask(ctx context.Context, taskID, userID string) (*Task, error) {
	query := `SELECT id, user_id, email_id, title, description, deadline, status, priority, created_at, updated_at, completed_at
              FROM tasks 
              WHERE id = $1 AND user_id = $2`

	var task Task
	err := db.QueryRowContext(ctx, query, taskID, userID).Scan(
		&task.ID, &task.UserID, &task.EmailID, &task.Title,
		&task.Description, &task.Deadline, &task.Status, &task.Priority,
		&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt)

	if err == sql.ErrNoRows {
		return nil, ErrTaskNotFound
	}
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (db *DB) GetUserTasks(ctx context.Context, filter TaskFilter) ([]Task, error) {
	query := `SELECT id, user_id, email_id, title, description, deadline, status, priority, created_at, updated_at, completed_at
              FROM tasks 
              WHERE user_id = $1`

	args := []interface{}{filter.UserID}
	argPos := 2

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *filter.Status)
		argPos++
	}

	if filter.Priority != nil {
		query += fmt.Sprintf(" AND priority = $%d", argPos)
		args = append(args, *filter.Priority)
		argPos++
	}

	if filter.FromDeadline != nil {
		query += fmt.Sprintf(" AND deadline >= $%d", argPos)
		args = append(args, *filter.FromDeadline)
		argPos++
	}

	if filter.ToDeadline != nil {
		query += fmt.Sprintf(" AND deadline <= $%d", argPos)
		args = append(args, *filter.ToDeadline)
		argPos++
	}

	query += " ORDER BY deadline ASC NULLS LAST, created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filter.Limit)
		argPos++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, filter.Offset)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(
			&task.ID, &task.UserID, &task.EmailID, &task.Title,
			&task.Description, &task.Deadline, &task.Status, &task.Priority,
			&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (db *DB) UpdateTask(ctx context.Context, taskID, userID string, update UpdateTaskRequest) error {
	setParts := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argPos := 1

	if update.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argPos))
		args = append(args, *update.Title)
		argPos++
	}

	if update.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argPos))
		args = append(args, *update.Description)
		argPos++
	}

	if update.Deadline != nil {
		setParts = append(setParts, fmt.Sprintf("deadline = $%d", argPos))
		args = append(args, *update.Deadline)
		argPos++
	}

	if update.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argPos))
		args = append(args, *update.Status)
		argPos++

		if *update.Status == "completed" {
			setParts = append(setParts, fmt.Sprintf("completed_at = $%d", argPos))
			args = append(args, time.Now())
			argPos++
		}
	}

	if update.Priority != nil {
		setParts = append(setParts, fmt.Sprintf("priority = $%d", argPos))
		args = append(args, *update.Priority)
		argPos++
	}

	if len(args) == 0 {
		return nil
	}

	query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = $%d AND user_id = $%d",
		strings.Join(setParts, ", "), argPos, argPos+1)
	args = append(args, taskID, userID)

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func (db *DB) DeleteTask(ctx context.Context, taskID, userID string) error {
	query := `DELETE FROM tasks WHERE id = $1 AND user_id = $2`
	result, err := db.ExecContext(ctx, query, taskID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func (db *DB) CompleteTask(ctx context.Context, taskID, userID string) error {
	query := `UPDATE tasks SET status = 'completed', completed_at = NOW(), updated_at = NOW() 
              WHERE id = $1 AND user_id = $2`

	result, err := db.ExecContext(ctx, query, taskID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func (db *DB) GetTaskStats(ctx context.Context, userID string) (*TaskStats, error) {
	query := `SELECT
                COUNT(*) as total,
                COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending,
                COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed,
                COUNT(CASE WHEN deadline < NOW() AND status != 'completed' THEN 1 END) as overdue,
                COUNT(CASE WHEN DATE(deadline) = CURRENT_DATE AND status != 'completed' THEN 1 END) as today,
                COUNT(CASE WHEN deadline >= CURRENT_DATE AND deadline < CURRENT_DATE + INTERVAL '7 days' AND status != 'completed' THEN 1 END) as this_week
              FROM tasks
              WHERE user_id = $1`

	var stats TaskStats
	err := db.QueryRowContext(ctx, query, userID).Scan(
		&stats.TotalTasks, &stats.PendingTasks, &stats.CompletedTasks,
		&stats.OverdueTasks, &stats.TodayTasks, &stats.ThisWeekTasks)

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (db *DB) TaskExists(ctx context.Context, emailID, userID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM tasks WHERE email_id = $1 AND user_id = $2)`
	err := db.QueryRowContext(ctx, query, emailID, userID).Scan(&exists)
	return exists, err
}
