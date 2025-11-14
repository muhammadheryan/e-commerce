package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

// Init initializes the global Zap logger
func Init(environment string) error {
	var config zap.Config

	if environment == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	var err error
	globalLogger, err = config.Build()
	if err != nil {
		return err
	}

	return nil
}

// Get returns the global logger
func Get() *zap.Logger {
	if globalLogger == nil {
		// Fallback to a basic logger if not initialized
		globalLogger, _ = zap.NewProduction()
	}
	return globalLogger
}

// Close flushes the logger
func Close() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// Info logs at info level
func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

// Error logs at error level
func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

// Debug logs at debug level
func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

// Warn logs at warn level
func Warn(msg string, fields ...zap.Field) {
	Get().Warn(msg, fields...)
}

// Fatal logs at fatal level and exits
func Fatal(msg string, fields ...zap.Field) {
	Get().Fatal(msg, fields...)
}
