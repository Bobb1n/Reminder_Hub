package zaplogger

import (
	"context"
	"log"
	"yfp/internal/logger"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Creating an adapter to make a Logger interface implementation
type zapAdapter struct {
	z *zap.Logger
}

// Создаём zapLogger, аналогичную функцию в целом можно написать
// и для любого другого логгера и также её настроить. В этом
// и заключается преимущество использования адаптера и интерфейсов
func NewLoggerAdapter(env string) (*zapAdapter, func()) {
	var loggerCfg zap.Config
	if env == "development" {
		loggerCfg = zap.Config{
			Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
			Development:      true,
			Encoding:         "console",
			EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}
	} else {
		loggerCfg = zap.NewProductionConfig()
		loggerCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	loggerCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	loggerCfg.DisableStacktrace = true

	logger, err := loggerCfg.Build(zap.AddCaller())
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	lo := zapAdapter{z: logger}
	return &lo, func() {
		_ = logger.Sync()
	}
}

// enrichZapFields - Выносит общую логику для Debug, Info и Error фукций
func enrichZapFields(ctx context.Context, fields []zap.Field) []zap.Field {
	id, ok := logger.RequestID(ctx)
	if !ok {
		id = "id request was not specified"
	}
	return append(fields, zap.String("RequestID", id))
}

func (l *zapAdapter) toZapFields(args ...any) []zap.Field {
	// если аргументов нечётное количество — логируем и обрезаем последний
	if len(args)%2 != 0 {
		//l.z.Warn("toZapFields called with odd number of arguments", zap.Int("len", len(args)))
		args = args[:len(args)-1]
	}

	zapFields := make([]zap.Field, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			//l.z.Warn("non-string key in toZapFields", zap.Any("key", args[i]))
			continue
		}

		value := args[i+1]
		switch v := value.(type) {
		case string:
			zapFields = append(zapFields, zap.String(key, v))
		case int:
			zapFields = append(zapFields, zap.Int(key, v))
		case bool:
			zapFields = append(zapFields, zap.Bool(key, v))
		default:
			zapFields = append(zapFields, zap.Any(key, v))
		}
	}

	return zapFields
}

func (l *zapAdapter) Info(ctx context.Context, msg string, fields ...any) {
	zapFields := l.toZapFields(fields...)
	l.z.Info(msg, enrichZapFields(ctx, zapFields)...)
}

func (l *zapAdapter) Debug(ctx context.Context, msg string, fields ...any) {
	zapFields := l.toZapFields(fields...)
	l.z.Debug(msg, enrichZapFields(ctx, zapFields)...)
}
func (l *zapAdapter) Error(ctx context.Context, msg string, fields ...any) {
	zapFields := l.toZapFields(fields...)
	l.z.Error(msg, enrichZapFields(ctx, zapFields)...)
}

func (l *zapAdapter) Warn(ctx context.Context, msg string, fields ...any) {
	zapFields := l.toZapFields(fields...)
	l.z.Warn(msg, enrichZapFields(ctx, zapFields)...)
}

func (l *zapAdapter) Panic(ctx context.Context, msg string, fields ...any) {
	zapFields := l.toZapFields(fields...)
	l.z.Panic(msg, enrichZapFields(ctx, zapFields)...)
}

func (l *zapAdapter) Fatal(ctx context.Context, msg string, fields ...any) {
	zapFields := l.toZapFields(fields...)
	l.z.Fatal(msg, enrichZapFields(ctx, zapFields)...)
}

func (l *zapAdapter) Sync() error {
	return l.z.Sync()
}

func (l *zapAdapter) GetLevel() string {
	return l.z.Level().String()
}
