package workflow

import (
	
	"fmt"
	"testing"
)

type MockStoreDB struct {
	ListWorkflowsFunc      func() ([]WorkflowSummary, error)
	GetWorkflowFunc        func(id string) (string, error)
	SaveWorkflowFunc       func(id, name, data, createdBy string) error
	DeleteWorkflowFunc     func(id string) error
	DeleteAllWorkflowsFunc func() error
}

func (m *MockStoreDB) ListWorkflows() ([]WorkflowSummary, error) { return m.ListWorkflowsFunc() }
func (m *MockStoreDB) GetWorkflow(id string) (string, error) {
	return m.GetWorkflowFunc(id)
}
func (m *MockStoreDB) SaveWorkflow(id, name, data, createdBy string) error {
	return m.SaveWorkflowFunc(id, name, data, createdBy)
}
func (m *MockStoreDB) DeleteWorkflow(id string) error        { return m.DeleteWorkflowFunc(id) }
func (m *MockStoreDB) DeleteAllWorkflows() error {
	if m.DeleteAllWorkflowsFunc != nil {
		return m.DeleteAllWorkflowsFunc()
	}
	return nil
}

func TestWorkflowStore_Exists(t *testing.T) {
	mock := &MockStoreDB{
		GetWorkflowFunc: func(id string) (string, error) {
			if id == "exists" {
				return "{}", nil
			}
			return "", fmt.Errorf("workflow not found: %s", id)
		},
	}

	store := NewWorkflowStore(mock, "/tmp")

	if !store.Exists("exists") {
		t.Error("expected exists to be true")
	}
	if store.Exists("missing") {
		t.Error("expected exists to be false")
	}
}

func TestWorkflowStore_List(t *testing.T) {
	mock := &MockStoreDB{
		ListWorkflowsFunc: func() ([]WorkflowSummary, error) {
			return []WorkflowSummary{
				{ID: "wf1", Name: "Workflow 1", Data: "{\"nodes\": [{}, {}, {}]}"},
			}, nil
		},
	}

	store := NewWorkflowStore(mock, "/tmp")
	files, err := store.List()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(files) != 1 || files[0].ID != "wf1" {
		t.Error("list result mismatch")
	}
	if files[0].NodeCount != 3 {
		t.Errorf("expected node count 3, got %d", files[0].NodeCount)
	}
}
