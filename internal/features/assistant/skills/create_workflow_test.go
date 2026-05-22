package skills

import (
	"fmt"
	"testing"
)

func TestCreateWorkflow(t *testing.T) {
	// Test ini akan memanggil API asli yang sedang berjalan
	skill := &CreateWorkflow{}

	args := map[string]any{
		"name": "Test Workflow AI",
		"nodes": []any{
			map[string]any{
				"id":     "n1",
				"action": "babi.babi",
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

	result, err := skill.Execute(args)
	if err != nil {
		result = fmt.Sprintf("Error : %v", err)
	}

	fmt.Println("\n--- HASIL EKSEKUSI CREATE WORKFLOW ---")
	fmt.Println(result)
	fmt.Println("---------------------------------------")
}
