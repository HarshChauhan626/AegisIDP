package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// Init initialises the global Zap logger.
// level should be one of: debug, info, warn, error
func Init(level string) error {
	zapLevel := zapcore.InfoLevel
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapLevel),
		Development: zapLevel == zapcore.DebugLevel,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	l, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = l
	return nil
}

// With returns a logger with additional fields.
func With(fields ...zap.Field) *zap.Logger {
	return Log.With(fields...)
}

// WithWorkflow returns a logger scoped to a workflow execution.
func WithWorkflow(workflowID string) *zap.Logger {
	return Log.With(zap.String("workflow_id", workflowID))
}
