package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveWorkflow(t *testing.T) {
	// Test ini memanggil API asli yang sedang berjalan
	draftSkill := &DraftWorkflow{}
	saveSkill := &SaveWorkflow{}

	tempID := "tmp_test_save_flow"

	// 1. Create the draft first
	draftArgs := map[string]any{
		"id":   tempID,
		"name": "Test Save Workflow Integration",
		"nodes": []any{
			map[string]any{
				"id":     "n1",
				"action": "timer.manual",
			},
			map[string]any{
				"id":     "n2",
				"action": "google_chrome.launch",
				"port":   9222,
			},
		},
		"edges": []any{
			map[string]any{
				"source":       "n1",
				"sourceHandle": "success",
				"target":       "n2",
			},
		},
	}

	draftResult, err := draftSkill.Execute(draftArgs)
	if err != nil {
		t.Fatalf("failed to execute draft_workflow: %v", err)
	}
	fmt.Printf("Create Draft Result: %s\n", draftResult)

	// 2. Now save it permanently
	saveArgs := map[string]any{
		"id": tempID,
	}

	saveResult, err := saveSkill.Execute(saveArgs)
	if err != nil {
		t.Fatalf("failed to execute save_workflow: %v", err)
	}
	fmt.Printf("Save Permanent Result: %s\n", saveResult)
}

func TestRemoveTempWorkflowFiles(t *testing.T) {
	tempDir := t.TempDir()
	tempID := "tmp_cleanup"

	jsonPath := filepath.Join(tempDir, tempID+".json")
	yamlPath := filepath.Join(tempDir, tempID+".yaml")
	if err := os.WriteFile(jsonPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write temp json: %v", err)
	}
	if err := os.WriteFile(yamlPath, []byte("name: test"), 0644); err != nil {
		t.Fatalf("failed to write temp yaml: %v", err)
	}

	if err := removeTempWorkflowFiles(tempDir, tempID); err != nil {
		t.Fatalf("failed to remove temp workflow files: %v", err)
	}

	if _, err := os.Stat(jsonPath); !os.IsNotExist(err) {
		t.Fatalf("expected json temp file to be removed, got err: %v", err)
	}
	if _, err := os.Stat(yamlPath); !os.IsNotExist(err) {
		t.Fatalf("expected yaml temp file to be removed, got err: %v", err)
	}
}
