package logger

import (
	"context"
)

type Logger interface {
	Debug(ctx context.Context, msg string, fields ...any)
	Info(ctx context.Context, msg string, fields ...any)
	Error(ctx context.Context, msg string, fields ...any)
	Warn(ctx context.Context, msg string, fields ...any)
	Panic(ctx context.Context, msg string, fields ...any)
	Fatal(ctx context.Context, msg string, fields ...any)
}

type CurrentLogger struct {
	logger Logger
}

func NewCurrentLogger(adapter Logger) *CurrentLogger {
	return &CurrentLogger{
		logger: adapter,
	}
}

func (cl *CurrentLogger) Debug(ctx context.Context, msg string, fields ...any) {
	cl.logger.Debug(ctx, msg, fields...)
}

func (cl *CurrentLogger) Info(ctx context.Context, msg string, fields ...any) {
	cl.logger.Info(ctx, msg, fields...)
}
func (cl *CurrentLogger) Error(ctx context.Context, msg string, fields ...any) {
	cl.logger.Error(ctx, msg, fields...)
}
func (cl *CurrentLogger) Warn(ctx context.Context, msg string, fields ...any) {
	cl.logger.Warn(ctx, msg, fields...)
}
func (cl *CurrentLogger) Panic(ctx context.Context, msg string, fields ...any) {
	cl.logger.Panic(ctx, msg, fields...)
}
func (cl *CurrentLogger) Fatal(ctx context.Context, msg string, fields ...any) {
	cl.logger.Fatal(ctx, msg, fields...)
}

type loggerRequestID string

const (
	loggerRequestIDKey loggerRequestID = "x-request_id"
)

// ReuestID получает из контекста id запроса
func RequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(loggerRequestIDKey).(string)
	return id, ok
}

// WithRequestID Добавляет в контекст ключ запроса и сам id запроса
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, loggerRequestIDKey, requestID)
}

// Ключ для получения логгера из контекста:
type ctxKey struct{}

// Функция, которая запихивает логгер в контекст
func WithLogger(ctx context.Context, logger *CurrentLogger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

// Функция, которая достаёт логгер из контекста
func LoggerFromContext(ctx context.Context) *CurrentLogger {
	logger, _ := ctx.Value(ctxKey{}).(*CurrentLogger)
	return logger
}

