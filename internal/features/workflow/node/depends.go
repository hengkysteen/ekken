package node

import (
	"fmt"
	"strings"
	"sync"
)

// DepRecord tracks which node+action combinations have been executed.
type DepRecord struct {
	NodeType string
	Action   string
}

type DependencyTracker struct {
	mu           sync.RWMutex
	executedDeps map[string][]DepRecord // key = workflow ID
}

func NewDependencyTracker() *DependencyTracker {
	return &DependencyTracker{
		executedDeps: make(map[string][]DepRecord),
	}
}

var GlobalTracker = NewDependencyTracker()

// RecordExecuted marks a node+action as executed for a workflow.
func (t *DependencyTracker) RecordExecuted(workflowID, nodeType, action string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.executedDeps[workflowID] = append(t.executedDeps[workflowID], DepRecord{
		NodeType: nodeType,
		Action:   action,
	})
}

// Clear resets the tracker state. Useful for tests.
func (t *DependencyTracker) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.executedDeps = make(map[string][]DepRecord)
}

// ClearWorkflow removes all recorded dependency data for a specific workflow ID.
// Must be called after a workflow run completes to prevent memory accumulation.
func (t *DependencyTracker) ClearWorkflow(workflowID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.executedDeps, workflowID)
}

// CheckDependsOn verifies that all declared dependencies have been executed for the given workflow.
func (t *DependencyTracker) CheckDependsOn(workflowID string, deps []DependsOn) error {
	t.mu.RLock()
	executed := t.executedDeps[workflowID]
	t.mu.RUnlock()

	for _, d := range deps {
		found := false
		for _, e := range executed {
			if e.NodeType != d.Node {
				continue
			}
			if d.Action == "" || e.Action == d.Action {
				found = true
				break
			}
		}
		if !found {
			if d.Action != "" {
				return fmt.Errorf("node type '%s' action '%s' must be executed before this node", d.Node, d.Action)
			}
			return fmt.Errorf("node type '%s' must be executed before this node", d.Node)
		}
	}
	return nil
}

// FormatDepError produces a human-readable dependency hint.
func FormatDepError(nodeID string, missing []string) string {
	return fmt.Sprintf("node '%s' requires: %s (add this node before it in the workflow)", nodeID, strings.Join(missing, ", "))
}
