package database

import "errors"

var (
	ErrIntegrationNotFound  = errors.New("integration not found")
	ErrDuplicateIntegration = errors.New("duplicate integration")
	ErrInvalidEmailAddress  = errors.New("invalid email address")
)
