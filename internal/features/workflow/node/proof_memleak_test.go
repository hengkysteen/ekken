package node

import (
	"fmt"
	"testing"
)

// TestDependencyTracker_MemoryLeak verifies two things:
//  1. Without ClearWorkflow, dependency data accumulates.
//  2. With ClearWorkflow, dependency data is cleaned after each workflow.
func TestDependencyTracker_MemoryLeak(t *testing.T) {
	t.Run("leak occurs without cleanup", func(t *testing.T) {
		tracker := NewDependencyTracker()

		for i := range 10 {
			wfID := fmt.Sprintf("workflow-%d", i)
			for j := range 5 {
				tracker.RecordExecuted(wfID, fmt.Sprintf("node_%d", j), "run")
			}
			// Intentionally skip ClearWorkflow to simulate the old lifecycle.
		}

		tracker.mu.RLock()
		liveEntries := len(tracker.executedDeps)
		tracker.mu.RUnlock()

		if liveEntries == 0 {
			t.Error("expected a leak without cleanup, but the map is empty; test setup is wrong")
		} else {
			t.Logf("confirmed: %d workflow IDs remain in memory without ClearWorkflow", liveEntries)
		}
	})

	t.Run("no leak after ClearWorkflow is called", func(t *testing.T) {
		tracker := NewDependencyTracker()

		totalWorkflows := 100

		for i := range totalWorkflows {
			wfID := fmt.Sprintf("workflow-%d", i)
			for j := range 10 {
				tracker.RecordExecuted(wfID, fmt.Sprintf("node_type_%d", j), "run")
			}
			// Simulate the correct lifecycle: clean up after the workflow finishes.
			tracker.ClearWorkflow(wfID)
		}

		tracker.mu.RLock()
		liveEntries := len(tracker.executedDeps)
		totalRecords := 0
		for _, records := range tracker.executedDeps {
			totalRecords += len(records)
		}
		tracker.mu.RUnlock()

		t.Logf("Workflow IDs in memory after %d runs: %d", totalWorkflows, liveEntries)
		t.Logf("Remaining DepRecord total: %d", totalRecords)

		if liveEntries != 0 {
			t.Errorf("leak remains: %d workflow IDs and %d records were not cleaned", liveEntries, totalRecords)
		} else {
			t.Log("Fix verified: no data remains after all workflows finish")
		}
	})
}
