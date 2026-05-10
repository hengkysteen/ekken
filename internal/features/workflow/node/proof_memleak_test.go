package node

import (
	"fmt"
	"testing"
)

// TestDependencyTracker_MemoryLeak membuktikan dua hal:
//  1. Tanpa ClearWorkflow → data menumpuk (leak terdokumentasi)
//  2. Dengan ClearWorkflow → data bersih setelah setiap workflow selesai (fix verified)
func TestDependencyTracker_MemoryLeak(t *testing.T) {
	t.Run("leak terjadi tanpa cleanup", func(t *testing.T) {
		tracker := NewDependencyTracker()

		for i := range 10 {
			wfID := fmt.Sprintf("workflow-%d", i)
			for j := range 5 {
				tracker.RecordExecuted(wfID, fmt.Sprintf("node_%d", j), "run")
			}
			// Sengaja TIDAK memanggil ClearWorkflow — simulasi kondisi lama
		}

		tracker.mu.RLock()
		liveEntries := len(tracker.executedDeps)
		tracker.mu.RUnlock()

		if liveEntries == 0 {
			t.Error("Harusnya ada leak tanpa cleanup, tapi map kosong — test setup salah")
		} else {
			t.Logf("Terkonfirmasi: %d workflow ID masih di memori tanpa ClearWorkflow", liveEntries)
		}
	})

	t.Run("tidak ada leak setelah ClearWorkflow dipanggil", func(t *testing.T) {
		tracker := NewDependencyTracker()

		totalWorkflows := 100

		for i := range totalWorkflows {
			wfID := fmt.Sprintf("workflow-%d", i)
			for j := range 10 {
				tracker.RecordExecuted(wfID, fmt.Sprintf("node_type_%d", j), "run")
			}
			// Simulasi lifecycle yang benar: cleanup setelah workflow selesai
			tracker.ClearWorkflow(wfID)
		}

		tracker.mu.RLock()
		liveEntries := len(tracker.executedDeps)
		totalRecords := 0
		for _, records := range tracker.executedDeps {
			totalRecords += len(records)
		}
		tracker.mu.RUnlock()

		t.Logf("Workflow ID di memori setelah %d run: %d", totalWorkflows, liveEntries)
		t.Logf("Total DepRecord tersisa: %d", totalRecords)

		if liveEntries != 0 {
			t.Errorf("Masih ada leak: %d workflow ID dan %d record tidak dibersihkan", liveEntries, totalRecords)
		} else {
			t.Log("Fix verified: tidak ada data yang tersisa setelah semua workflow selesai")
		}
	})
}
