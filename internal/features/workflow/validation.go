package workflow

import (
	"strings"

	"ekken/internal/features/workflow/node"
)

// Validate checks a workflow structure against the system's core validation rules.
func Validate(wf Workflow) node.ValidationResult {
	issues := make([]string, 0)
	if strings.TrimSpace(wf.Name) == "" {
		issues = append(issues, "workflow.name is required")
	}

	index := make(map[string]node.NodeSpec)
	for _, r := range node.GlobalRegistry.AllSpecs() {
		index[r.Type] = r
	}
	node.ValidateNodes(wf.Nodes, "workflow.nodes", index, &issues)
	
	// Ensure at least one trigger node exists
	hasTrigger := false
	for _, n := range wf.Nodes {
		spec, ok := index[n.Type]
		if !ok {
			continue
		}
		for _, tag := range spec.Tags {
			if strings.EqualFold(tag, "trigger") {
				hasTrigger = true
				break
			}
		}
		if hasTrigger {
			break
		}
	}
	if !hasTrigger && len(wf.Nodes) > 0 {
		issues = append(issues, "workflow must contain at least one Trigger node")
	}

	return node.ValidationResult{Valid: len(issues) == 0, Errors: issues}
}
