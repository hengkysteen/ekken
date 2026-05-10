package workflow

import (
	"os"
	"path/filepath"
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
