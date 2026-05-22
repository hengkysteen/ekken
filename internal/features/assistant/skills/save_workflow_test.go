package skills

import (
	"fmt"
	"testing"
)

func TestSaveWorkflow(t *testing.T) {
	// Test ini memanggil API asli yang sedang berjalan
	createSkill := &CreateWorkflow{}
	saveSkill := &SaveWorkflow{}

	tempID := "tmp_test_save_flow"

	// 1. Create the draft first
	createArgs := map[string]any{
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

	createResult, err := createSkill.Execute(createArgs)
	if err != nil {
		t.Fatalf("failed to execute create_workflow: %v", err)
	}
	fmt.Printf("Create Draft Result: %s\n", createResult)

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
