package skills

import (
	"fmt"
	"testing"
)

func TestNodesList(t *testing.T) {
	// Test ini akan memanggil API asli yang sedang berjalan
	skill := &NodesList{}

	result, err := skill.Execute(nil)
	if err != nil {
		t.Fatalf("Gagal mengambil data asli: %v. Pastikan server ekken sudah jalan.", err)
	}

	fmt.Println("\n--- HASIL OUTPUT LIST NODE (DATA ASLI) ---")
	fmt.Println(result)
	fmt.Printf("--- TOTAL LENGTH: %d CHARACTERS ---\n", len(result))
	fmt.Println("------------------------------------------")
}
