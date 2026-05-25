package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type MockRuntimeDB struct {
	UpdateStatusFunc  func(id, status string, iteration int) error
	UpdateLastRunFunc func(id string, lastRunAt time.Time) error
	GetStatusFunc     func(id string) (WorkflowStatusInfo, error)
}

func (m *MockRuntimeDB) UpdateStatus(id, status string, iteration int) error {
	return m.UpdateStatusFunc(id, status, iteration)
}
func (m *MockRuntimeDB) UpdateLastRun(id string, lastRunAt time.Time) error {
	return m.UpdateLastRunFunc(id, lastRunAt)
}
func (m *MockRuntimeDB) GetStatus(id string) (WorkflowStatusInfo, error) {
	return m.GetStatusFunc(id)
}

type MockEventStreamer struct {
	SendFunc       func(id string, msg SSEMessage)
	CreateFunc     func(id string)
	FinishFunc     func(id string)
	SendGlobalFunc func(msg SSEMessage)
}

func (m *MockEventStreamer) Send(id string, msg SSEMessage) { m.SendFunc(id, msg) }
func (m *MockEventStreamer) Create(id string)               { m.CreateFunc(id) }
func (m *MockEventStreamer) Finish(id string)               { m.FinishFunc(id) }
func (m *MockEventStreamer) SendGlobal(msg SSEMessage) {
	if m.SendGlobalFunc != nil {
		m.SendGlobalFunc(msg)
	}
}

func TestRuntimeService_Status(t *testing.T) {
	mockDB := &MockRuntimeDB{
		GetStatusFunc: func(id string) (WorkflowStatusInfo, error) {
			return WorkflowStatusInfo{Status: "done", Iteration: 5}, nil
		},
	}
	mockSSE := &MockEventStreamer{}
	// MockWorkflowStore is already defined in workflows_test.go (same package)
	// But I need a MockWorkflowService (which is WorkflowServicer)

	service := NewRuntimeService(nil, mockDB, mockSSE, "/tmp")
	status := service.Status("wf-1")

	if status.Status != "done" {
		t.Errorf("expected done, got %s", status.Status)
	}
	if status.Iteration != 5 {
		t.Errorf("expected 5, got %d", status.Iteration)
	}
}

func TestRuntimeService_DeleteLogs(t *testing.T) {
	tempDir := t.TempDir()
	mockDB := &MockRuntimeDB{}
	mockSSE := &MockEventStreamer{
		SendFunc: func(id string, msg SSEMessage) {},
	}

	service := NewRuntimeService(nil, mockDB, mockSSE, tempDir)
	wfID := "wf-test-delete"

	// 1. Create a dummy log file
	logFolder := filepath.Join(tempDir, "logs", "workflow")
	os.MkdirAll(logFolder, 0755)
	logFile := filepath.Join(logFolder, wfID+".txt")
	err := os.WriteFile(logFile, []byte("test log content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test log file: %v", err)
	}

	// 2. Add to in-memory cache
	service.OnLog(wfID, "info", "test message", "")

	// 3. Verify file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatal("log file should exist before deletion")
	}

	// 4. Call DeleteLogs
	err = service.DeleteLogs(wfID)
	if err != nil {
		t.Errorf("DeleteLogs failed: %v", err)
	}

	// 5. Verify file is gone
	if _, err := os.Stat(logFile); err == nil {
		t.Error("log file should have been deleted")
	}

	// 6. Verify cache is cleared (check memory cache directly if possible, or via Logs() method)
	logs := service.Logs(wfID)
	if len(logs) > 0 {
		t.Error("cache should be empty after deletion")
	}
}

func TestRuntimeService_LogsReturnsLast500FromFile(t *testing.T) {
	tempDir := t.TempDir()
	service := NewRuntimeService(nil, &MockRuntimeDB{}, &MockEventStreamer{}, tempDir)
	wfID := "wf-log-window"

	var b strings.Builder
	for i := 0; i < 600; i++ {
		entry := LogEntry{
			Time:    time.Date(2026, 5, 24, 10, 0, i%60, 0, time.UTC),
			Level:   "info",
			Message: fmt.Sprintf("entry-%03d", i),
		}
		data, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("failed to marshal log entry: %v", err)
		}
		b.Write(data)
		b.WriteByte('\n')
	}
	writeWorkflowLogFile(t, tempDir, wfID, b.String())

	logs := service.Logs(wfID)
	if len(logs) != 500 {
		t.Fatalf("len(logs) = %d, want 500", len(logs))
	}
	if logs[0].Message != "entry-100" {
		t.Fatalf("first returned log = %q, want entry-100", logs[0].Message)
	}
	if logs[len(logs)-1].Message != "entry-599" {
		t.Fatalf("last returned log = %q, want entry-599", logs[len(logs)-1].Message)
	}
}

func TestRuntimeService_LogsReadsLineLargerThanScannerLimit(t *testing.T) {
	tempDir := t.TempDir()
	service := NewRuntimeService(nil, &MockRuntimeDB{}, &MockEventStreamer{}, tempDir)
	wfID := "wf-large-log-line"
	largeRaw := strings.Repeat("x", 70*1024)
	entry := LogEntry{
		Time:    time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC),
		Level:   "debug",
		Message: "large raw",
		Raw:     largeRaw,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("failed to marshal log entry: %v", err)
	}
	writeWorkflowLogFile(t, tempDir, wfID, string(data)+"\n")

	logs := service.Logs(wfID)
	if len(logs) != 1 {
		t.Fatalf("len(logs) = %d, want 1", len(logs))
	}
	if logs[0].Raw != largeRaw {
		t.Fatalf("raw length = %d, want %d", len(logs[0].Raw), len(largeRaw))
	}
}

func TestRuntimeService_LogsParsesJSONLAndOldTextFormatAndSkipsEmptyLines(t *testing.T) {
	tempDir := t.TempDir()
	service := NewRuntimeService(nil, &MockRuntimeDB{}, &MockEventStreamer{}, tempDir)
	wfID := "wf-mixed-log-format"
	jsonEntry := LogEntry{
		Time:    time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC),
		Level:   "info",
		Message: "json line",
	}
	data, err := json.Marshal(jsonEntry)
	if err != nil {
		t.Fatalf("failed to marshal log entry: %v", err)
	}
	content := "\n" +
		string(data) + "\n" +
		"[2026-04-26 13:09:47.763] [INFO] old line | RAW: old raw\n" +
		"   \n"
	writeWorkflowLogFile(t, tempDir, wfID, content)

	logs := service.Logs(wfID)
	if len(logs) != 2 {
		t.Fatalf("len(logs) = %d, want 2", len(logs))
	}
	if logs[0].Message != "json line" {
		t.Fatalf("first message = %q, want json line", logs[0].Message)
	}
	if logs[1].Level != "info" || logs[1].Message != "old line" || logs[1].Raw != "old raw" {
		t.Fatalf("old format log parsed as %+v", logs[1])
	}
}

func writeWorkflowLogFile(t *testing.T, dataDir, wfID, content string) {
	t.Helper()
	logFolder := filepath.Join(dataDir, "logs", "workflow")
	if err := os.MkdirAll(logFolder, 0755); err != nil {
		t.Fatalf("failed to create log folder: %v", err)
	}
	logFile := filepath.Join(logFolder, wfID+".txt")
	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write log file: %v", err)
	}
}
