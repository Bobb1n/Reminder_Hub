package logger

import (
	"context"
	"log"

	"reminder-hub/pkg/logger"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)


func Init(env string) *logger.CurrentLogger {
	var loggerCfg zap.Config
	if env == "production" {
		loggerCfg = zap.NewProductionConfig()
		loggerCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	} else {
		loggerCfg = zap.Config{
			Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
			Development:      true,
			Encoding:         "console",
			EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}
	}

	loggerCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	loggerCfg.DisableStacktrace = true

	zapLogger, err := loggerCfg.Build(zap.AddCaller())
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	adapter := &zapAdapterWithoutFX{z: zapLogger}
	return logger.NewCurrentLogger(adapter)
}

type zapAdapterWithoutFX struct {
	z *zap.Logger
}

func (l *zapAdapterWithoutFX) toZapFields(args ...any) []zap.Field {
	if len(args)%2 != 0 {
		args = args[:len(args)-1]
	}

	zapFields := make([]zap.Field, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
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

func (l *zapAdapterWithoutFX) enrichZapFields(ctx context.Context, fields []zap.Field) []zap.Field {
	id, ok := logger.RequestID(ctx)
	if !ok {
		id = "id request was not specified"
	}
	return append(fields, zap.String("RequestID", id))
}

func (l *zapAdapterWithoutFX) Info(ctx context.Context, msg string, fields ...any) {
	zapFields := l.toZapFields(fields...)
	l.z.Info(msg, l.enrichZapFields(ctx, zapFields)...)
}

func (l *zapAdapterWithoutFX) Debug(ctx context.Context, msg string, fields ...any) {
	zapFields := l.toZapFields(fields...)
	l.z.Debug(msg, l.enrichZapFields(ctx, zapFields)...)
}

func (l *zapAdapterWithoutFX) Error(ctx context.Context, msg string, fields ...any) {
	zapFields := l.toZapFields(fields...)
	l.z.Error(msg, l.enrichZapFields(ctx, zapFields)...)
}

func (l *zapAdapterWithoutFX) Warn(ctx context.Context, msg string, fields ...any) {
	zapFields := l.toZapFields(fields...)
	l.z.Warn(msg, l.enrichZapFields(ctx, zapFields)...)
}

func (l *zapAdapterWithoutFX) Panic(ctx context.Context, msg string, fields ...any) {
	zapFields := l.toZapFields(fields...)
	l.z.Panic(msg, l.enrichZapFields(ctx, zapFields)...)
}

func (l *zapAdapterWithoutFX) Fatal(ctx context.Context, msg string, fields ...any) {
	zapFields := l.toZapFields(fields...)
	l.z.Fatal(msg, l.enrichZapFields(ctx, zapFields)...)
}
