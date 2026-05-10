package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// DevPrintf prints a formatted debug message.
func DevPrintf(format string, args ...interface{}) {
	slog.Debug(fmt.Sprintf(format, args...))
}

// Info logs an informational message.
func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

// Error logs an error message.
func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

// Debug logs a debug message.
func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

// Warn logs a warning message.
func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

// CleanHandler is a minimal slog.Handler that writes clean messages to stderr.
type CleanHandler struct {
	level slog.Level
}

func (h *CleanHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *CleanHandler) Handle(_ context.Context, r slog.Record) error {
	fmt.Fprintln(os.Stderr, strings.TrimRight(r.Message, "\n"))
	return nil
}

func (h *CleanHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *CleanHandler) WithGroup(_ string) slog.Handler      { return h }

// NewCleanLogger returns a slog.Logger configured for console output.
// When verbose is true, it logs at debug level; otherwise error level only.
func NewCleanLogger(verbose bool) *slog.Logger {
	level := slog.LevelError
	if verbose {
		level = slog.LevelDebug
	}
	l := slog.New(&CleanHandler{level: level})
	slog.SetDefault(l)
	return l
}
