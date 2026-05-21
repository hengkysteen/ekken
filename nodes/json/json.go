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
	Action node.Action
}

func init() {
	node.GlobalRegistry.Register(node.NodeRegistration{
		Spec: node.Spec{
			Meta: node.Meta{
				Type:        "json",
				Label:       "JSON",
				Icon:        "https://www.svgrepo.com/show/458247/json.svg",
				Tags:        []string{"System"},
				Description: "Extract a value from JSON using a dot-notation path.",
			},

			DefaultAction: "extract",
			Actions: []node.Action{
				{
					Type:         "extract",
					Label:        "Extract",
					Description:  "Extract a value by path",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
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
					AutoLayout: [][]node.AutoLayout{
						{{Key: "input", Component: "textarea", Flex: 24, Options: map[string]any{"helper": "JSON data or variable to extract from"}}},
						{{Key: "path", Component: "input", Flex: 24, Options: map[string]any{"helper": "Dot-notation path to extract"}}},
					},
				},
				{
					Type:         "extract_stream",
					Label:        "Extract Stream",
					Description:  "Extract values from SSE, JSONL, or JSON array data",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
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
							Required: false,
							Label:    "JSON path",
						},
					},
					AutoLayout: [][]node.AutoLayout{
						{{Key: "input", Component: "textarea", Flex: 24, Options: map[string]any{"helper": "SSE, JSONL, JSON array, or variable to extract from"}}},
						{{Key: "path", Component: "input", Flex: 24, Options: map[string]any{"helper": "Path applied to each stream item. Leave empty to return all items."}}},
					},
				},
			},
			Outputs: []node.HandleEdge{
				{Key: "success", Label: "Success"},
				{Key: "error", Label: "Error", Tone: "error"},
			},
		},
		ExecutorFactory: func(action node.Action) node.NodeExecutor {
			return &JsonNode{Action: action}
		},
	})
}

func (n *JsonNode) Execute(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	select {
	case <-ctx.Stop:
		return node.NodeExecutionResult{}, node.ErrNodeStopped
	default:
	}

	action := n.Action.Type
	if action == "" {
		action = "extract"
	}
	if action != "extract" && action != "extract_stream" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("unknown action: %s", action)
	}

	path, _ := node.FieldValue(n.Action, "path").(string)
	if path == "" {
		if action == "extract" {
			return node.NodeExecutionResult{}, fmt.Errorf("path is required")
		}
	}

	inputRaw, _ := node.FieldValue(n.Action, "input").(string)
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

	var result any
	var err error
	if action == "extract_stream" {
		result, err = extractStream(data, path)
	} else {
		result, err = traverse(data, path)
	}
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

func coerceExtractableString(data any) any {
	raw, ok := data.(string)
	if !ok {
		return data
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return data
	}

	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		var parsed any
		if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
			return parsed
		}
	}

	return data
}

func extractStream(data any, path string) (any, error) {
	source, itemPath, err := streamSourceAndPath(data, path)
	if err != nil {
		return nil, err
	}
	items, err := parseStreamItems(source)
	if err != nil {
		return nil, err
	}
	if itemPath == "" {
		return items, nil
	}

	results := make([]any, 0, len(items))
	var firstErr error
	for i, item := range items {
		value, err := traverse(item, itemPath)
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("item %d: %v", i, err)
			}
			continue
		}
		results = append(results, value)
	}
	if len(results) == 0 && firstErr != nil {
		return nil, firstErr
	}
	return results, nil
}

func streamSourceAndPath(data any, path string) (any, string, error) {
	if path == "data" {
		return data, "", nil
	}
	if after, ok := strings.CutPrefix(path, "data."); ok {
		return data, after, nil
	}

	if before, after, ok := strings.Cut(path, ".data"); ok {
		if after != "" && !strings.HasPrefix(after, ".") {
			return data, path, nil
		}
		source, err := traverse(data, before)
		if err != nil {
			return nil, "", err
		}
		return source, strings.TrimPrefix(after, "."), nil
	}

	return data, path, nil
}

func parseStreamItems(data any) ([]any, error) {
	data = coerceExtractableString(data)
	switch v := data.(type) {
	case []any:
		return v, nil
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return nil, fmt.Errorf("stream input is empty")
		}
		if items, ok := parseSSEItems(trimmed); ok {
			return items, nil
		}
		if items, ok := parseJSONLItems(trimmed); ok {
			return items, nil
		}
		return nil, fmt.Errorf("unsupported stream input")
	default:
		return []any{v}, nil
	}
}

func parseSSEItems(raw string) ([]any, bool) {
	items := make([]any, 0)

	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	for _, block := range strings.Split(normalized, "\n\n") {
		var dataLines []string

		for _, line := range strings.Split(block, "\n") {
			line = strings.TrimSpace(line)

			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}

			field, value, ok := strings.Cut(line, ":")
			if !ok {
				continue
			}
			field = strings.TrimSpace(field)
			value = strings.TrimSpace(value)

			if field == "data" {
				dataLines = append(dataLines, value)
			}
		}
		if len(dataLines) == 0 {
			continue
		}

		dataRaw := strings.Join(dataLines, "\n")

		var parsed any
		if err := json.Unmarshal([]byte(dataRaw), &parsed); err != nil {
			items = append(items, dataRaw)
		} else {
			items = append(items, parsed)
		}
	}

	if len(items) == 0 {
		return nil, false
	}
	return items, true
}

func parseJSONLItems(raw string) ([]any, bool) {
	items := make([]any, 0)
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var item any
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			return nil, false
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil, false
	}
	return items, true
}

func splitPath(path string) []string {
	return strings.Split(path, ".")
}
