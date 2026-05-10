package node

import (
	"fmt"
	"strings"
)

// ValidateNodes recursively validates a list of nodes against the provided index of specifications.
func ValidateNodes(ns []Node, path string, index map[string]NodeSpec, issues *[]string) {
	// Track which combinations of Node Type + Action exist in the workflow
	nodesInWorkflow := make(map[string]bool)
	for _, n := range ns {
		action, ok := n.Config["action"].(string)
		if !ok || action == "" {
			if spec, exists := index[n.Type]; exists {
				action = spec.DefaultAction
			}
		}
		// Key format: "node_type:action"
		nodesInWorkflow[fmt.Sprintf("%s:%s", n.Type, action)] = true
	}

	for i, node := range ns {
		nodePath := fmt.Sprintf("%s[%d]", path, i)
		if strings.TrimSpace(node.Type) == "" {
			*issues = append(*issues, nodePath+".type is required")
			continue
		}

		// Identify node for better error messages
		nodeIdentifier := node.Label
		if nodeIdentifier == "" {
			nodeIdentifier = node.Type
		}

		def, ok := index[node.Type]
		if !ok {
			*issues = append(*issues, nodePath+".type is unknown")
			continue
		}

		config := node.Config
		if config == nil {
			config = map[string]interface{}{}
		}

		// Validate GlobalFields
		var fieldsToValidate []NodeField
		for _, field := range def.GlobalFields {
			fieldsToValidate = append(fieldsToValidate, field)
		}

		// Validate action-specific fields
		action, hasAction := config["action"].(string)
		if !hasAction || action == "" {
			action = def.DefaultAction
		}

		if action != "" {
			actionFound := false
			for _, actionDef := range def.Actions {
				if actionDef.Key == action {
					fieldsToValidate = append(fieldsToValidate, actionDef.Fields...)
					actionFound = true
					break
				}
			}
			if !actionFound {
				*issues = append(*issues, fmt.Sprintf("Node \"%s\": action \"%s\" is not valid", nodeIdentifier, action))
			}
		}

		for _, field := range fieldsToValidate {
			if !field.Required {
				continue
			}
			value, exists := config[field.Key]
			if !exists || IsEmptyValue(value) {
				// Try to auto-populate from default if missing
				if field.Default != nil && !IsEmptyValue(field.Default) {
					if node.Config == nil {
						ns[i].Config = make(map[string]interface{})
					}
					ns[i].Config[field.Key] = field.Default
					continue
				}
				*issues = append(*issues, fmt.Sprintf("Node \"%s\" action \"%s\": field \"%s\" is required", nodeIdentifier, action, field.Key))
			}
		}

		if len(node.Nodes) > 0 && !def.SupportsNodes {
			*issues = append(*issues, nodePath+".nodes is not supported by this node type")
		}
		if len(node.Nodes) == 0 && def.SupportsNodes && node.Type == "loop" {
			*issues = append(*issues, nodePath+".nodes must not be empty")
		}
		if len(node.Nodes) > 0 {
			ValidateNodes(node.Nodes, nodePath+".nodes", index, issues)
		}

		// Validate dependsOn: both node type and action must match
		allDeps := append(def.DependsOn, node.DependsOn...)
		for _, dep := range allDeps {
			depKey := fmt.Sprintf("%s:%s", dep.Node, dep.Action)
			if !nodesInWorkflow[depKey] {
				*issues = append(*issues, fmt.Sprintf("Node \"%s\" depends on Node \"%s\" (\"%s\")", nodeIdentifier, dep.Node, dep.Action))
			}
		}
	}
}

// IsEmptyValue checks if a value is nil or an empty string.
func IsEmptyValue(value interface{}) bool {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	default:
		return value == nil
	}
}
