package workflow

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type devWriter struct{}

func (d devWriter) Write(p []byte) (int, error) {
	if gin.Mode() != gin.ReleaseMode {
		return os.Stdout.Write(p)
	}
	return len(p), nil
}

func (r *Runner) formatConsoleLog(name, level, message string) string {
	return fmt.Sprintf("[%s] [%s]: %s\n", name, strings.ToUpper(level), message)
}

func (r *Runner) logInfo(name, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	r.writeToConsole(r.formatConsoleLog(name, "INFO", msg))
	if r.observer != nil {
		r.observer.OnLog(name, "info", msg, "")
	}
}

func (r *Runner) logError(name, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	r.writeToConsole(r.formatConsoleLog(name, "ERROR", msg))
	if r.observer != nil {
		r.observer.OnLog(name, "error", msg, "")
	}
}

func (r *Runner) logDebug(name, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	r.writeToConsole(r.formatConsoleLog(name, "DEBUG", msg))
	if r.observer != nil {
		r.observer.OnLog(name, "debug", msg, "")
	}
}

func (r *Runner) logRaw(name, level, message, raw string) {
	r.writeToConsole(r.formatConsoleLog(name, level, message+" (with raw data)"))
	if r.observer != nil {
		r.observer.OnLog(name, level, message, raw)
	}
}

func (r *Runner) writeToConsole(message string) {
	out := devWriter{}
	_, _ = io.WriteString(out, message)
}
