package workflow

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"ekken/internal/features/workflow/node"
)

// WorkflowServicer defines the interface for workflow management.
type WorkflowServicer interface {
	List() ([]WorkflowFile, error)
	Get(id string) (Workflow, []byte, error)
	Create(wf Workflow) (Workflow, string, error)
	Update(id string, wf Workflow) (Workflow, string, error)
	Delete(id string) error
	DeleteAll() error
	Import(raw []byte) (Workflow, string, error)
	Export(id string) ([]byte, error)
	Validate(wf Workflow) node.ValidationResult
	ValidateForRun(wf Workflow) node.ValidationResult
}

// WorkflowStorer defines the interface for persisting workflow data.
type WorkflowStorer interface {
	List() ([]WorkflowFile, error)
	Get(id string) (Workflow, []byte, error)
	Exists(id string) bool
	Save(id string, wf Workflow) (string, error)
	Delete(id string) error
	DeleteAll() error
}

// WorkflowService implements the core business logic for workflows.
type WorkflowService struct {
	store WorkflowStorer
}

func NewWorkflowService(store WorkflowStorer) *WorkflowService {
	return &WorkflowService{store: store}
}

// --- Basic CRUD ---

func (s *WorkflowService) List() ([]WorkflowFile, error) {
	return s.store.List()
}

func (s *WorkflowService) Get(id string) (Workflow, []byte, error) {
	return s.store.Get(id)
}

func (s *WorkflowService) Create(wf Workflow) (Workflow, string, error) {
	// Generate 10-char random ID for better UX and aesthetics
	if wf.ID == "" {
		wf.ID = generateRandomID(10)
	}
	if wf.CreatedAt.IsZero() {
		wf.CreatedAt = time.Now()
	}
	if wf.UpdatedAt.IsZero() {
		wf.UpdatedAt = time.Now()
	}
	result := s.Validate(wf)
	if !result.Valid {
		return wf, "", errors.New(strings.Join(result.Errors, "\n"))
	}
	if s.store.Exists(wf.ID) {
		return wf, "", fmt.Errorf("workflow already exists")
	}
	path, err := s.store.Save(wf.ID, wf)
	return wf, path, err
}

func (s *WorkflowService) Update(id string, wf Workflow) (Workflow, string, error) {
	if !s.store.Exists(id) {
		return wf, "", ErrWorkflowNotFound
	}

	oldWf, _, err := s.store.Get(id)
	if err == nil {
		if wf.ID == "" {
			wf.ID = oldWf.ID
		}
		if wf.CreatedAt.IsZero() {
			wf.CreatedAt = oldWf.CreatedAt
		}
		if wf.CreatedBy == "" {
			wf.CreatedBy = oldWf.CreatedBy
		}
	}

	// Ensure ID is present
	if wf.ID == "" {
		wf.ID = generateRandomID(10)
	}
	wf.UpdatedAt = time.Now()

	result := s.Validate(wf)
	if !result.Valid {
		return wf, "", errors.New(strings.Join(result.Errors, "\n"))
	}

	path, err := s.store.Save(wf.ID, wf)
	if err != nil {
		return wf, "", err
	}

	return wf, path, nil
}

func (s *WorkflowService) Delete(id string) error {
	return s.store.Delete(id)
}

func (s *WorkflowService) DeleteAll() error {
	return s.store.DeleteAll()
}

// --- Serialization ---

func (s *WorkflowService) Import(raw []byte) (Workflow, string, error) {
	var wf Workflow
	if err := json.Unmarshal(raw, &wf); err != nil {
		return Workflow{}, "", err
	}
	return s.Create(wf)
}

func (s *WorkflowService) Export(id string) ([]byte, error) {
	_, raw, err := s.store.Get(id)
	return raw, err
}

// --- Validation ---

func (s *WorkflowService) Validate(wf Workflow) node.ValidationResult {
	return Validate(wf)
}

func (s *WorkflowService) ValidateForRun(wf Workflow) node.ValidationResult {
	result := s.Validate(wf)
	if !result.Valid {
		return result
	}

	if len(wf.Nodes) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "workflow must have at least one node to run")
	}

	return result
}

// --- Helpers ---

// generateRandomID creates a random alphanumeric string of a given length.
func generateRandomID(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	rand.Read(result)
	for i := 0; i < length; i++ {
		result[i] = chars[result[i]%byte(len(chars))]
	}
	return string(result)
}
