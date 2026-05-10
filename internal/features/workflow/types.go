package workflow

import (
	"time"

	"ekken/internal/features/workflow/node"
)

// Workflow represents a complete automation graph consisting of nodes and edges.
type Workflow struct {
	ID        string                   `json:"id"`                  // Unique identifier for the workflow
	Name      string                   `json:"name"`                // Human-readable name
	CreatedBy string                   `json:"created_by"`          // User who created the workflow
	CreatedAt time.Time                `json:"created_at"`          // Timestamp of creation
	UpdatedAt time.Time                `json:"updated_at"`          // Timestamp of last update
	Nodes     []node.Node              `json:"nodes"`               // List of nodes in the workflow
	Edges     []node.Edge              `json:"edges,omitempty"`     // List of connections between nodes
	Positions map[string]node.Position `json:"positions,omitempty"` // UI positions of nodes on the canvas
}

// WorkflowFile represents a summary of a workflow for listing.
type WorkflowFile struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedBy string     `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	NodeCount int        `json:"nodeCount"`
	Status    string     `json:"status"`
	LastRunAt *time.Time `json:"last_run_at,omitempty"`
}

// LogEntry represents a single log line from a workflow run.
type LogEntry struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
	Raw     string    `json:"raw,omitempty"`
}

// WorkflowRunStatus represents the real-time execution state of a workflow.
type WorkflowRunStatus struct {
	ID         string     `json:"id"`
	Name       string     `json:"name,omitempty"`
	Status     string     `json:"status"`
	Iteration  int        `json:"iteration"`
	LastNode   string     `json:"last_node"`
	LastUpdate time.Time  `json:"last_update"`
	LastRunAt  *time.Time `json:"last_run_at,omitempty"`
}

type SSEMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

var ()
