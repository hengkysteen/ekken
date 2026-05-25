package workflow

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"ekken/internal/features/workflow/node"
	"ekken/internal/logger"
)

const maxWorkflowLogs = 500

type RuntimeServicer interface {
	Running() []WorkflowRunStatus
	RunByID(id string) error
	RunWorkflow(wf Workflow) error
	Stop(id string) error
	Status(id string) WorkflowRunStatus
	Logs(id string) []LogEntry
	DeleteLogs(id string) error
}

type RuntimeDatabase interface {
	UpdateStatus(id, status string, iteration int) error
	UpdateLastRun(id string, lastRunAt time.Time) error
	GetStatus(id string) (WorkflowStatusInfo, error)
}

type EventStreamer interface {
	Send(id string, msg SSEMessage)
	Create(id string)
	Finish(id string)
	SendGlobal(msg SSEMessage)
}

type activeRun struct {
	cancel context.CancelFunc
	name   string
}

// workflowLogFile is a per-workflow buffered writer that stays open
// for the entire workflow run, avoiding repeated open/close syscalls.
type workflowLogFile struct {
	mu     sync.Mutex
	f      *os.File
	buf    *bufio.Writer
	closed bool
}

func (w *workflowLogFile) writeLine(data []byte) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return
	}
	_, _ = w.buf.Write(data)
	_ = w.buf.WriteByte('\n')
}

func (w *workflowLogFile) flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.closed {
		_ = w.buf.Flush()
	}
}

func (w *workflowLogFile) close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return
	}
	w.closed = true
	_ = w.buf.Flush()
	_ = w.f.Close()
}

type RuntimeService struct {
	workflows WorkflowServicer
	database  RuntimeDatabase
	sse       EventStreamer
	runner    *Runner
	dataDir   string
	mu        sync.RWMutex
	errLogMu  sync.Mutex                  // guards global error log file only
	logFiles  map[string]*workflowLogFile // per-workflow buffered writers
	runs      map[string]*activeRun
	logs      map[string][]LogEntry // In-memory cache for live runs
}

func NewRuntimeService(workflows WorkflowServicer, database RuntimeDatabase, sse EventStreamer, dataDir string) *RuntimeService {
	s := &RuntimeService{
		workflows: workflows,
		database:  database,
		sse:       sse,
		dataDir:   dataDir,
		runs:      make(map[string]*activeRun),
		logs:      make(map[string][]LogEntry),
		logFiles:  make(map[string]*workflowLogFile),
	}

	// Workflow runner initialization with s as the observer
	s.runner = New(s, node.GlobalRegistry)

	// Ensure log directories exist
	os.MkdirAll(filepath.Join(dataDir, "logs", "workflow"), 0755)
	os.MkdirAll(filepath.Join(dataDir, "logs", "error"), 0755)

	return s
}

func (s *RuntimeService) OnStatusUpdate(id, status string, iteration int) {
	if err := s.database.UpdateStatus(id, status, iteration); err != nil {
		logger.Error("Failed to update workflow status in database", "id", id, "error", err)
	}

	// Try to get name from memory first
	s.mu.RLock()
	run, hasRun := s.runs[id]
	s.mu.RUnlock()

	name := id
	if hasRun {
		name = run.name
	}

	// Send to single workflow stream
	s.sse.Send(id, SSEMessage{Type: "status_update", Data: map[string]interface{}{
		"status":    status,
		"iteration": iteration,
		"id":        id,
		"name":      name,
	}})
	// Broadcast to global stream
	s.sse.SendGlobal(SSEMessage{Type: "status_update", Data: map[string]interface{}{
		"id":     id,
		"name":   name,
		"status": status,
	}})
}

func (s *RuntimeService) OnLog(id, level, message, raw string) {
	entry := LogEntry{
		Time:    time.Now(),
		Level:   level,
		Message: message,
		Raw:     raw,
	}

	// 1. Update In-Memory Cache (for very fast live access)
	s.mu.Lock()
	s.logs[id] = append(s.logs[id], entry)
	if len(s.logs[id]) > maxWorkflowLogs {
		s.logs[id] = s.logs[id][len(s.logs[id])-maxWorkflowLogs:]
	}
	s.mu.Unlock()

	// 2. Persist to File
	s.writeLogToFile(id, entry)

	// 3. Broadcast via SSE
	s.sse.Send(id, SSEMessage{Type: "log_entry", Data: entry})
}

func (s *RuntimeService) writeLogToFile(wfID string, entry LogEntry) {
	// 1. Prepare JSON line
	jsonData, err := json.Marshal(entry)
	if err != nil {
		logger.Error("Failed to marshal log entry to JSON", "error", err)
		return
	}

	// 2. Write via the per-workflow buffered handle (no open/close per entry)
	s.mu.RLock()
	wlf, ok := s.logFiles[wfID]
	s.mu.RUnlock()
	if ok {
		wlf.writeLine(jsonData)
	}

	// 3. Global error log — open/close is acceptable here: errors are infrequent
	if entry.Level == "error" {
		date := time.Now().Format("2006-01-02")
		errLogPath := filepath.Join(s.dataDir, "logs", "error", date+".txt")
		s.errLogMu.Lock()
		f, ferr := os.OpenFile(errLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if ferr == nil {
			timestamp := entry.Time.Format("15:04:05.000")
			_, _ = fmt.Fprintf(f, "[%s] [%s] %s\n", timestamp, wfID, entry.Message)
			_ = f.Close()
		}
		s.errLogMu.Unlock()
	}
}

func (s *RuntimeService) RunByID(id string) error {
	wf, _, err := s.workflows.Get(id)
	if err != nil {
		return err
	}
	return s.RunWorkflow(wf)
}

func (s *RuntimeService) RunWorkflow(wf Workflow) error {
	if wf.ID == "" {
		return fmt.Errorf("workflow ID is required")
	}

	// Validate before running
	vResult := s.workflows.ValidateForRun(wf)
	if !vResult.Valid {
		return fmt.Errorf("workflow validation failed: %s", strings.Join(vResult.Errors, "; "))
	}

	s.mu.Lock()
	if _, exists := s.runs[wf.ID]; exists {
		s.mu.Unlock()
		return fmt.Errorf("workflow is already running")
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.runs[wf.ID] = &activeRun{
		cancel: cancel,
		name:   wf.Name,
	}
	s.mu.Unlock()

	s.sse.Create(wf.ID)

	// Open a buffered log file handle that will stay open for the entire run.
	wfLogPath := filepath.Join(s.dataDir, "logs", "workflow", wf.ID+".txt")
	if f, ferr := os.OpenFile(wfLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); ferr == nil {
		wlf := &workflowLogFile{f: f, buf: bufio.NewWriterSize(f, 32*1024)}
		s.mu.Lock()
		s.logFiles[wf.ID] = wlf
		s.mu.Unlock()
	}

	go func() {
		// Log start separator
		s.OnLog(wf.ID, "info", "==========================================", "")
		s.OnLog(wf.ID, "info", ">>> NEW WORKFLOW SESSION STARTED <<<", "")
		s.OnLog(wf.ID, "info", "==========================================", "")

		time.Sleep(100 * time.Millisecond)

		err := s.runner.Run(ctx, wf)

		// Update last_run_at timestamp
		now := time.Now()
		if updateErr := s.database.UpdateLastRun(wf.ID, now); updateErr != nil {
			logger.Error("Error updating last_run_at", "id", wf.ID, "error", updateErr)
		}

		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, node.ErrNodeStopped) {
				s.OnStatusUpdate(wf.ID, "stopped", 0)
			} else if errors.Is(err, node.ErrWorkflowComplete) {
				s.OnStatusUpdate(wf.ID, "done", 0)
			} else {
				s.OnStatusUpdate(wf.ID, "error", 0)
			}
		} else {
			// This else block handles if err == nil (though normally it returns ErrWorkflowComplete)
			s.OnStatusUpdate(wf.ID, "done", 0)
		}

		// Flush and close the log file handle, then cleanup run state.
		s.mu.Lock()
		wlf := s.logFiles[wf.ID]
		delete(s.logFiles, wf.ID)
		delete(s.runs, wf.ID)
		cancel()
		s.mu.Unlock()
		if wlf != nil {
			wlf.close() // flush buffer + close fd outside the lock
		}

		// Release dependency tracking data for this workflow to prevent memory accumulation.
		node.GlobalTracker.ClearWorkflow(wf.ID)

		time.AfterFunc(2*time.Second, func() {
			s.sse.Finish(wf.ID)
		})
	}()

	return nil
}

func (s *RuntimeService) Stop(id string) error {
	logger.Info("RuntimeService.Stop called", "id", id)
	s.mu.Lock()
	run, ok := s.runs[id]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("workflow is not running")
	}
	run.cancel()
	// delete(s.runs, id) // We'll let the goroutine cleanup handle this
	s.mu.Unlock()

	s.OnStatusUpdate(id, "stopped", 0)
	return nil
}

func (s *RuntimeService) Status(id string) WorkflowRunStatus {
	s.mu.RLock()
	run, running := s.runs[id]
	s.mu.RUnlock()

	if running {
		return WorkflowRunStatus{ID: id, Name: run.name, Status: "running"}
	}

	info, err := s.database.GetStatus(id)
	if err != nil {
		return WorkflowRunStatus{ID: id, Status: "idle"}
	}

	status := info.Status
	if status == "" {
		status = "idle"
	}

	result := WorkflowRunStatus{
		ID:        id,
		Status:    status,
		Iteration: info.Iteration,
	}
	if info.LastRunAt != nil {
		result.LastRunAt = info.LastRunAt
	}
	return result
}

func (s *RuntimeService) Running() []WorkflowRunStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]WorkflowRunStatus, 0, len(s.runs))
	for id, run := range s.runs {
		item := WorkflowRunStatus{ID: id, Name: run.name, Status: "running"}
		if info, err := s.database.GetStatus(id); err == nil {
			item.Iteration = info.Iteration
		}
		items = append(items, item)
	}
	return items
}

func (s *RuntimeService) Logs(id string) []LogEntry {
	// Flush buffered writes before reading so the file reflects the latest data.
	s.mu.RLock()
	wlf, hasHandle := s.logFiles[id]
	s.mu.RUnlock()
	if hasHandle {
		wlf.flush()
	}

	// Try to read from file first
	filePath := filepath.Join(s.dataDir, "logs", "workflow", id+".txt")
	file, err := os.Open(filePath)
	if err == nil {
		defer file.Close()

		fileLogs := make([]LogEntry, 0, maxWorkflowLogs)
		reader := bufio.NewReader(file)
		for {
			line, readErr := reader.ReadString('\n')
			if line != "" {
				if entry, ok := parseWorkflowLogLine(strings.TrimRight(line, "\r\n")); ok {
					fileLogs = appendLogWindow(fileLogs, entry, maxWorkflowLogs)
				}
			}

			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				logger.Error("Failed to read workflow log file", "id", id, "error", readErr)
				break
			}
		}
		return fileLogs
	}

	// Fallback to memory if file not found or error
	s.mu.RLock()
	defer s.mu.RUnlock()
	logs := s.logs[id]
	out := make([]LogEntry, len(logs))
	copy(out, logs)
	return out
}

func appendLogWindow(logs []LogEntry, entry LogEntry, limit int) []LogEntry {
	if limit <= 0 {
		return logs
	}
	if len(logs) < limit {
		return append(logs, entry)
	}
	copy(logs, logs[1:])
	logs[len(logs)-1] = entry
	return logs
}

func parseWorkflowLogLine(line string) (LogEntry, bool) {
	if strings.TrimSpace(line) == "" {
		return LogEntry{}, false
	}

	entry := LogEntry{}
	// Try to parse as JSON (new format)
	if err := json.Unmarshal([]byte(line), &entry); err == nil {
		return entry, true
	}

	// Fallback to old text format parsing for backward compatibility
	if len(line) < 25 {
		return LogEntry{}, false
	}

	// Extract timestamp: [2026-04-26 13:09:47.763]
	tStr := line[1:24]
	t, err := time.ParseInLocation("2006-01-02 15:04:05.000", tStr, time.Local)
	if err == nil {
		entry.Time = t
	}

	if len(line) <= 26 {
		return LogEntry{}, false
	}
	remaining := line[26:]

	// Extract level: [INFO]
	if strings.HasPrefix(remaining, "[") {
		lvlEnd := strings.Index(remaining, "]")
		if lvlEnd != -1 {
			entry.Level = strings.ToLower(remaining[1:lvlEnd])
			if len(remaining) > lvlEnd+2 {
				msgPart := remaining[lvlEnd+2:]
				rawIdx := strings.Index(msgPart, " | RAW: ")
				if rawIdx != -1 {
					entry.Message = msgPart[:rawIdx]
					entry.Raw = msgPart[rawIdx+8:]
				} else {
					entry.Message = msgPart
				}
			}
		}
	}
	return entry, true
}

func (s *RuntimeService) DeleteLogs(id string) error {
	// 1. Close open log handle (if any) and clear caches
	s.mu.Lock()
	wlf := s.logFiles[id]
	delete(s.logFiles, id)
	delete(s.logs, id)
	s.mu.Unlock()
	if wlf != nil {
		wlf.close()
	}

	// 2. Delete the log file
	filePath := filepath.Join(s.dataDir, "logs", "workflow", id+".txt")
	if _, err := os.Stat(filePath); err == nil {
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("failed to delete log file: %w", err)
		}
	}

	return nil
}
