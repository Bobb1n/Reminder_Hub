package database

import "errors"

var (
	ErrTaskNotFound        = errors.New("task not found")
	ErrDuplicateTask       = errors.New("duplicate task")
	ErrInvalidTaskStatus   = errors.New("invalid task status")
	ErrInvalidTaskPriority = errors.New("invalid task priority")
)