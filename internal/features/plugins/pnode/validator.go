package pnode

import (
	"ekken/internal/features/plugins/kind"
	"ekken/internal/features/workflow/node"
	"encoding/json"
	"fmt"
)

func decodePluginSpec(plugin kind.Plugin) (PluginSpec, error) {
	if len(plugin.Spec) == 0 {
		return PluginSpec{}, fmt.Errorf("plugin.spec is required")
	}

	var wrapped struct {
		Runner RunnerSpec      `json:"runner"`
		Node   json.RawMessage `json:"node"`
	}
	if err := json.Unmarshal(plugin.Spec, &wrapped); err != nil {
		return PluginSpec{}, fmt.Errorf("failed to unmarshal node plugin spec: %w", err)
	}

	if len(wrapped.Node) > 0 {
		var nodeSpec node.Spec
		if err := json.Unmarshal(wrapped.Node, &nodeSpec); err != nil {
			return PluginSpec{}, fmt.Errorf("failed to unmarshal spec.node: %w", err)
		}
		return PluginSpec{
			Runner: wrapped.Runner,
			Node:   nodeSpec,
		}, nil
	}

	var legacyNodeSpec node.Spec
	if err := json.Unmarshal(plugin.Spec, &legacyNodeSpec); err != nil {
		return PluginSpec{}, fmt.Errorf("failed to unmarshal spec.node: %w", err)
	}
	return PluginSpec{Node: legacyNodeSpec}, nil
}

// ValidateManifest validates a node plugin envelope.
func ValidateManifest(plugin kind.Plugin) error {
	spec, err := decodePluginSpec(plugin)
	if err != nil {
		return err
	}
	if spec.Runner.Command == "" {
		return fmt.Errorf("plugin.spec.runner.command is required")
	}
	return validateNodeSpec(spec.Node)
}

func validateNodeSpec(nodeDef node.Spec) error {
	if nodeDef.Type == "" {
		return fmt.Errorf("plugin.node.type is required")
	}
	if nodeDef.Label == "" {
		return fmt.Errorf("plugin node '%s' label is required", nodeDef.Type)
	}
	if len(nodeDef.Tags) == 0 {
		return fmt.Errorf("plugin node '%s' tags are required", nodeDef.Type)
	}
	if len(nodeDef.Actions) == 0 {
		return fmt.Errorf("plugin node '%s': must have at least one action", nodeDef.Type)
	}
	for _, action := range nodeDef.Actions {
		if err := validateNodeFields(nodeDef.Type, action.Fields); err != nil {
			return err
		}
		if err := validateLayout(nodeDef.Type, action.Fields, action.AutoLayout); err != nil {
			return err
		}
	}
	if len(nodeDef.GlobalFields) > 0 {
		if err := validateNodeFields(nodeDef.Type, nodeDef.GlobalFields); err != nil {
			return err
		}
	}
	if err := validateNodeOutputs(nodeDef.Type, nodeDef.Outputs); err != nil {
		return err
	}
	return nil
}

func validateNodeFields(nodeType string, fields []node.NodeField) error {
	validTypes := map[string]bool{
		"string":  true,
		"number":  true,
		"boolean": true,
		"json":    true,
		"array":   true,
	}
	seenKeys := make(map[string]bool)
	for i, field := range fields {
		if field.Key == "" {
			return fmt.Errorf("plugin node '%s' field[%d]: key is required", nodeType, i)
		}
		if seenKeys[field.Key] {
			return fmt.Errorf("plugin node '%s': duplicate field key '%s'", nodeType, field.Key)
		}
		seenKeys[field.Key] = true
		if field.Type == "" {
			return fmt.Errorf("plugin node '%s' field '%s': type is required", nodeType, field.Key)
		}
		if !validTypes[field.Type] {
			return fmt.Errorf("plugin node '%s' field '%s': invalid type '%s' (must be one of: string, number, boolean, json, array)", nodeType, field.Key, field.Type)
		}
		if field.Label == "" {
			return fmt.Errorf("plugin node '%s' field '%s': label is required", nodeType, field.Key)
		}
	}
	return nil
}

func validateNodeOutputs(nodeType string, outputs []node.HandleEdge) error {
	if len(outputs) == 0 {
		return fmt.Errorf("plugin node '%s': must have at least one output", nodeType)
	}
	validTones := map[string]bool{
		"success": true,
		"error":   true,
		"warning": true,
		"info":    true,
		"neutral": true,
	}
	seenKeys := make(map[string]bool)
	for i, output := range outputs {
		if output.Key == "" {
			return fmt.Errorf("plugin node '%s' output[%d]: key is required", nodeType, i)
		}
		if seenKeys[output.Key] {
			return fmt.Errorf("plugin node '%s': duplicate output key '%s'", nodeType, output.Key)
		}
		seenKeys[output.Key] = true
		if output.Label == "" {
			return fmt.Errorf("plugin node '%s' output '%s': label is required", nodeType, output.Key)
		}
		if output.Tone == "" {
			return fmt.Errorf("plugin node '%s' output '%s': tone is required", nodeType, output.Key)
		}
		if !validTones[output.Tone] {
			return fmt.Errorf("plugin node '%s' output '%s': invalid tone '%s' (must be one of: success, error, warning, info, neutral)", nodeType, output.Key, output.Tone)
		}
	}
	return nil
}

var validLayoutComponents = map[string]bool{
	"input":        true,
	"number":       true,
	"number-s1":    true,
	"number-s2":    true,
	"switch":       true,
	"select":       true,
	"textarea":     true,
	"radio":        true,
	"slider":       true,
	"jsonEditor":   true,
	"color_picker": true,
	"colorPicker":  true,
	"datePicker":   true,
	"timePicker":   true,
	"text":         true,
	"file_picker":  true,
}

func validateLayout(nodeType string, fields []node.NodeField, layout [][]node.AutoLayout) error {
	if len(layout) == 0 {
		return nil
	}

	fieldKeys := make(map[string]bool)
	for _, f := range fields {
		fieldKeys[f.Key] = true
	}

	for rowIdx, row := range layout {
		if len(row) == 0 {
			return fmt.Errorf("plugin node '%s' layout row[%d]: must not be empty", nodeType, rowIdx)
		}
		for colIdx, item := range row {
			if item.Key == "" {
				return fmt.Errorf("plugin node '%s' layout[%d][%d]: key is required", nodeType, rowIdx, colIdx)
			}
			if !fieldKeys[item.Key] {
				return fmt.Errorf("plugin node '%s' layout[%d][%d]: unknown field key '%s'", nodeType, rowIdx, colIdx, item.Key)
			}
			if item.Flex < 1 {
				return fmt.Errorf("plugin node '%s' layout[%d][%d]: flex must be >= 1", nodeType, rowIdx, colIdx)
			}
			if item.Component != "" {
				if !validLayoutComponents[item.Component] {
					return fmt.Errorf("plugin node '%s' layout[%d][%d]: invalid component '%s'", nodeType, rowIdx, colIdx, item.Component)
				}
			}
		}
	}
	return nil
}
