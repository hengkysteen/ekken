package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

)

type StoreDatabase interface {
	ListWorkflows() ([]WorkflowSummary, error)
	GetWorkflow(id string) (string, error)
	SaveWorkflow(id, name, data, createdBy string) error
	DeleteWorkflow(id string) error
	DeleteAllWorkflows() error
}

var ErrWorkflowNotFound = errors.New("workflow not found")

type WorkflowStore struct {
	db      StoreDatabase
	dataDir string
}

func NewWorkflowStore(database StoreDatabase, dataDir string) *WorkflowStore {
	return &WorkflowStore{db: database, dataDir: dataDir}
}

func (s *WorkflowStore) List() ([]WorkflowFile, error) {
	summaries, err := s.db.ListWorkflows()
	if err != nil {
		return nil, err
	}
	files := make([]WorkflowFile, 0, len(summaries))
	for _, item := range summaries {
		status := item.Status
		if status == "" {
			status = "idle"
		}

		var wf Workflow
		nodeCount := 0
		if err := json.Unmarshal([]byte(item.Data), &wf); err == nil {
			nodeCount = len(wf.Nodes)
		}

		files = append(files, WorkflowFile{
			ID:        item.ID,
			Name:      item.Name,
			CreatedBy: item.CreatedBy,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			NodeCount: nodeCount,
			Status:    status,
			LastRunAt: item.LastRunAt,
		})
	}
	return files, nil
}

func (s *WorkflowStore) Get(id string) (Workflow, []byte, error) {
	raw, err := s.db.GetWorkflow(id)
	if err != nil {
		if err.Error() == fmt.Sprintf("workflow not found: %s", id) {
			return Workflow{}, nil, ErrWorkflowNotFound
		}
		return Workflow{}, nil, err
	}

	var wf Workflow
	if err := json.Unmarshal([]byte(raw), &wf); err != nil {
		return Workflow{}, nil, fmt.Errorf("unmarshal workflow: %w", err)
	}

	return wf, []byte(raw), nil
}

func (s *WorkflowStore) Save(id string, wf Workflow) (string, error) {
	data, err := json.Marshal(wf)
	if err != nil {
		return "", fmt.Errorf("marshal workflow: %w", err)
	}

	if err := s.db.SaveWorkflow(id, wf.Name, string(data), wf.CreatedBy); err != nil {
		return "", err
	}
	return id, nil
}

func (s *WorkflowStore) Delete(id string) error {
	if err := s.db.DeleteWorkflow(id); err != nil {
		if err.Error() == fmt.Sprintf("workflow not found: %s", id) {
			return ErrWorkflowNotFound
		}
		return err
	}

	// Clean up workflow log file
	logPath := filepath.Join(s.dataDir, "logs", "workflow", id+".txt")
	if err := os.Remove(logPath); err != nil && !os.IsNotExist(err) {
		// Log the error but don't fail the delete operation
		fmt.Printf("Warning: failed to delete workflow log file %s: %v\n", logPath, err)
	}

	return nil
}

func (s *WorkflowStore) DeleteAll() error {
	return s.db.DeleteAllWorkflows()
}

func (s *WorkflowStore) Exists(id string) bool {
	_, err := s.db.GetWorkflow(id)
	return err == nil
}

func WorkflowJSON(wf Workflow) ([]byte, error) {
	return json.MarshalIndent(wf, "", "  ")
}
