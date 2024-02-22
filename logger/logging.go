package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

var logger *slog.Logger

func Init(verbose bool, json bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	if json {
		logger = slog.New(
			slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level}),
		)
	} else {

		logger = slog.New(
			tint.NewHandler(os.Stderr, &tint.Options{
				Level:      level,
				TimeFormat: time.Kitchen,
			}))

	}
}

func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}

func init() {
	Init(false, false)
}
