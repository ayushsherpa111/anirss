package utils

import (
	"log"
	"log/slog"
	"os"
	"path"
)

const (
	logPath = "./log/"
)

func NewLogger(fileName string) *slog.Logger {
	logPath := path.Join(logPath, fileName)
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("Failed to create log file %s. Err: %s\n", logPath, err.Error())
	}

	logHandler := slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	sLogger := slog.New(logHandler)
	return sLogger
}
