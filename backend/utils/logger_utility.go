package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

// InitLogger initializes the structured logger
func InitLogger() error {
	config := zap.NewProductionConfig()

	// Customize the encoding
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.StacktraceKey = "" // Disable stacktrace in logs

	// Set log level
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	var err error
	Logger, err = config.Build()
	if err != nil {
		return err
	}

	return nil
}

// Sync flushes any buffered log entries
func SyncLogger() {
	if Logger != nil {
		Logger.Sync()
	}
}