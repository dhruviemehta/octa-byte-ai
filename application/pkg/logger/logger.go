package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New() *zap.SugaredLogger {
	config := zap.NewProductionConfig()

	// Set log level from environment
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "debug":
		config.Level.SetLevel(zapcore.DebugLevel)
	case "info":
		config.Level.SetLevel(zapcore.InfoLevel)
	case "warn":
		config.Level.SetLevel(zapcore.WarnLevel)
	case "error":
		config.Level.SetLevel(zapcore.ErrorLevel)
	default:
		config.Level.SetLevel(zapcore.InfoLevel)
	}

	// Configure encoder for structured JSON logging
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "caller"

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return logger.Sugar()
}
