package skills

import (
	"fmt"
	"testing"
)

func TestGetNodeActions(t *testing.T) {
	skill := &NodesActions{}

	args := map[string]any{
		"actions": []any{"shell"},
	}

	result, err := skill.Execute(args)
	if err != nil {
		t.Fatalf("Gagal mengambil data asli: %v. Pastikan server ekken sudah jalan.", err)
	}

	fmt.Println("\n--- HASIL OUTPUT DETAIL NODE (DATA ASLI) ---")
	fmt.Println(result)
	fmt.Printf("--- TOTAL LENGTH: %d CHARACTERS ---\n", len(result))
	fmt.Println("-------------------------------------------")
}
