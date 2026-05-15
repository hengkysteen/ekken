package node

import (
	"fmt"
	"reflect"
	"strings"
)

// ValidateNodes recursively validates a list of nodes against the provided index of specifications.
func ValidateNodes(ns []Node, path string, index map[string]NodeSpec, issues *[]string) {
	// Track which combinations of Node Type + Action exist in the workflow prior to the current node
	nodesInWorkflow := make(map[string]bool)

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

		// Validate action-specific fields
		action := node.Action.Key
		if action == "" {
			action = def.DefaultAction
			ns[i].Action.Key = action
		}

		var fieldsToValidate []NodeField
		fieldsToValidate = append(fieldsToValidate, def.GlobalFields...)

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

		fieldIndex := make(map[string]NodeField, len(fieldsToValidate))
		for _, field := range fieldsToValidate {
			fieldIndex[field.Key] = field
		}

		for _, field := range fieldsToValidate {
			if !field.Required {
				continue
			}
			value := FieldValue(node.Action, field.Key)
			if !IsEmptyValue(value) {
				continue
			}
			// Try to auto-populate from default if missing
			if field.Default != nil && !IsEmptyValue(field.Default) {
				setFieldDefault(&ns[i].Action, field.Key, field.Default)
				continue
			}
			*issues = append(*issues, fmt.Sprintf("Node \"%s\" action \"%s\": field \"%s\" is required", nodeIdentifier, action, field.Key))
		}

		// Validate instance fields against catalog schema. The payload only needs key/value;
		// schema metadata from the client is ignored unless it explicitly conflicts.
		for _, userField := range node.Action.Fields {
			catalogField, ok := fieldIndex[userField.Key]
			if !ok {
				*issues = append(*issues, fmt.Sprintf("Node \"%s\" field \"%s\": unknown field", nodeIdentifier, userField.Key))
				continue
			}
			if userField.Type != "" && userField.Type != catalogField.Type {
				*issues = append(*issues, fmt.Sprintf("Node \"%s\" field \"%s\": invalid type. Expected \"%s\", got \"%s\"", nodeIdentifier, userField.Key, catalogField.Type, userField.Type))
			}
			if !IsEmptyValue(userField.Value) && !isValidValueType(userField.Value, catalogField.Type) {
				*issues = append(*issues, fmt.Sprintf("Node \"%s\" field \"%s\": invalid value type. Expected \"%s\"", nodeIdentifier, userField.Key, catalogField.Type))
			}
			if !IsEmptyValue(userField.Value) && !isValidOption(userField.Value, catalogField.Options) {
				*issues = append(*issues, fmt.Sprintf("Node \"%s\" field \"%s\": invalid option value", nodeIdentifier, userField.Key))
			}
		}

		// Validate dependsOn: both node type and action must match
		allDeps := append(def.DependsOn, node.DependsOn...)
		for _, dep := range allDeps {
			depKey := fmt.Sprintf("%s:%s", dep.Node, dep.Action)
			if !nodesInWorkflow[depKey] {
				*issues = append(*issues, fmt.Sprintf("Node \"%s\" depends on Node \"%s\" (\"%s\") to be executed before it", nodeIdentifier, dep.Node, dep.Action))
			}
		}

		// Register this node as available for subsequent nodes
		nodesInWorkflow[fmt.Sprintf("%s:%s", node.Type, action)] = true
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

func setFieldDefault(action *NodeAction, key string, value any) {
	for j := range action.Fields {
		if action.Fields[j].Key == key {
			action.Fields[j].Value = value
			return
		}
	}
	action.Fields = append(action.Fields, NodeField{Key: key, Value: value})
}

func isValidValueType(value any, typ string) bool {
	if typ == "" {
		return true
	}
	switch typ {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			return true
		default:
			return false
		}
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "json":
		switch value.(type) {
		case map[string]any, []any:
			return true
		default:
			return false
		}
	default:
		return true
	}
}

func isValidOption(value any, options any) bool {
	if options == nil {
		return true
	}
	rv := reflect.ValueOf(options)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return true
	}
	valueString := fmt.Sprint(value)
	for i := 0; i < rv.Len(); i++ {
		option := rv.Index(i).Interface()
		if fmt.Sprint(option) == valueString {
			return true
		}
		if m, ok := option.(map[string]string); ok {
			if m["value"] == valueString {
				return true
			}
		}
		if m, ok := option.(map[string]any); ok {
			if fmt.Sprint(m["value"]) == valueString {
				return true
			}
		}
	}
	return false
}
