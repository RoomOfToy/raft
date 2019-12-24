package raft

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
)

// var Logger = newLogger(LOGGER_LEVEL, "")

func newLogger(level string, filePath string) *zap.Logger {
	cfg := zap.NewProductionConfig()
	if filePath == "" {
		filePath = "stdout"
	}
	cfg.OutputPaths = []string{filePath}
	var l zapcore.Level
	switch level {
	case "debug":
		l = zap.DebugLevel
	case "info":
		l = zap.InfoLevel
	case "warn":
		l = zap.WarnLevel
	case "error":
		l = zap.ErrorLevel
	case "fatal":
		l = zap.FatalLevel
	case "panic":
		l = zap.PanicLevel
	default:
		l = zap.DebugLevel
		log.Println("Log level '" + level + "' not recognized. Default set to Debug")
	}
	cfg.Level = zap.NewAtomicLevelAt(l)
	logger, err := cfg.Build()
	if err != nil {
		log.Panicf("Error when build logger: %s", err.Error())
	}
	return logger
}
