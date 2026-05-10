package json

import (
	"ekken/internal/features/workflow/node"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type JsonNode struct {
	Config map[string]any
}

func init() {
	node.GlobalRegistry.Register(node.NodeRegistration{
		NodeSpec: node.NodeSpec{
			NodeMetadata: node.NodeMetadata{
				Type:  "json",
				Label: "JSON",
				Icon:  "https://www.svgrepo.com/show/458247/json.svg",
				Tags:  []string{"System"},
			},
			Description:   "Extract a value from JSON using a dot-notation path.",
			DefaultAction: "extract",
			Actions: []node.NodeAction{
				{
					Key:         "extract",
					Label:       "Extract",
					Description: "Extract a value by path",
					HasResponse: true,
					Fields: []node.NodeField{
						{
							Key:      "input",
							Type:     "string",
							Required: true,
							Label:    "Input",
						},
						{
							Key:      "path",
							Type:     "string",
							Required: true,
							Label:    "JSON path",
						},
					},
					Form: [][]node.Form{
						{{Key: "input", Component: "textarea", Flex: 24, FormOptions: map[string]any{"helper": "JSON data or variable to extract from"}}},
						{{Key: "path", Component: "input", Flex: 24, FormOptions: map[string]any{"helper": "Dot-notation path to extract"}}},
					},
				},
			},
			Outputs: []node.NodeOutputDef{
				{Key: "success", Label: "Success"},
				{Key: "error", Label: "Error", Tone: "error"},
			},
		},
		ExecutorFactory: func(config map[string]any, childNodes []node.Node) node.NodeExecutor {
			return &JsonNode{Config: config}
		},
	})
}

func (n *JsonNode) Execute(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	select {
	case <-ctx.Stop:
		return node.NodeExecutionResult{}, node.ErrNodeStopped
	default:
	}

	path, _ := n.Config["path"].(string)
	if path == "" {
		return node.NodeExecutionResult{}, fmt.Errorf("path is required")
	}

	inputRaw, _ := n.Config["input"].(string)
	if inputRaw == "" {
		return node.NodeExecutionResult{}, fmt.Errorf("input is required")
	}

	var data interface{}
	isExactVar := false

	// Check if input is exactly a variable reference like "{{my_var}}"
	if strings.HasPrefix(strings.TrimSpace(inputRaw), "{{") && strings.HasSuffix(strings.TrimSpace(inputRaw), "}}") {
		varName := strings.TrimSpace(inputRaw)
		varName = strings.TrimPrefix(varName, "{{")
		varName = strings.TrimSuffix(varName, "}}")
		varName = strings.TrimSpace(varName)

		isExactVar = true
		if val, ok := ctx.Variables[varName]; ok {
			data = val
		}
	}

	if !isExactVar && data == nil {
		// Fallback to parsed string
		parsedStr := node.ParseTemplate(inputRaw, ctx.Variables)

		// Try to parse string as json
		var unmarshaled interface{}
		if err := json.Unmarshal([]byte(parsedStr), &unmarshaled); err == nil {
			data = unmarshaled
		} else {
			// If it's not valid JSON, we just pass the string (traverse will probably fail if a path is specified)
			data = parsedStr
		}
	}

	result, err := traverse(data, path)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, err
	}

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: result,
		Type:     &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
	}, nil
}

// traverse walks a map/slice using dot-notation path (e.g. "body.choices[0].message.content")
func traverse(data any, path string) (any, error) {
	if path == "" {
		return data, nil
	}

	segments := splitPath(path)
	current := data

	for _, seg := range segments {
		if current == nil {
			return nil, fmt.Errorf("cannot access '%s' on nil", seg)
		}

		v := reflect.ValueOf(current)

		// Handle array access like choices[0]
		open := strings.Index(seg, "[")
		close := strings.Index(seg, "]")

		if open != -1 && close > open {
			key := seg[:open]
			idx, err := strconv.Atoi(seg[open+1 : close])
			if err != nil {
				return nil, fmt.Errorf("invalid index in '%s'", seg)
			}

			if key != "" {
				// Access map key first
				if v.Kind() != reflect.Map {
					return nil, fmt.Errorf("expected object at '%s', got %v", key, v.Kind())
				}
				val := v.MapIndex(reflect.ValueOf(key))
				if !val.IsValid() {
					return nil, fmt.Errorf("key '%s' not found", key)
				}
				current = val.Interface()
				if current == nil {
					return nil, fmt.Errorf("key '%s' is nil", key)
				}
				v = reflect.ValueOf(current)
			}

			// Access array index
			if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
				return nil, fmt.Errorf("expected array at '%s', got %v", key, v.Kind())
			}
			if idx < 0 || idx >= v.Len() {
				return nil, fmt.Errorf("index %d out of range at '%s'", idx, key)
			}
			current = v.Index(idx).Interface()
		} else {
			// Normal map access
			if v.Kind() != reflect.Map {
				return nil, fmt.Errorf("expected object at '%s', got %v", seg, v.Kind())
			}
			val := v.MapIndex(reflect.ValueOf(seg))
			if !val.IsValid() {
				return nil, fmt.Errorf("key '%s' not found", seg)
			}
			current = val.Interface()
		}
	}
	return current, nil
}

// splitPath splits "body.choices[0].message.content" into ["body", "choices[0]", "message", "content"]
func splitPath(path string) []string {
	return strings.Split(path, ".")
}
