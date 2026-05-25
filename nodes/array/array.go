package array

import (
	"encoding/json"
	"fmt"
	"strings"

	"ekken/internal/features/workflow/node"
)

type ArrayNode struct {
	Action node.Action
}

func init() {
	node.GlobalRegistry.Register(node.NodeRegistration{
		Spec: node.Spec{
			Meta: node.Meta{
				Type:        "array",
				Label:       "Array",
				Icon:        "https://www.svgrepo.com/show/458043/array.svg",
				Tags:        []string{"System"},
				Description: "Array utility operations.",
			},
			DefaultAction: "last",
			Actions: []node.Action{
				{
					Type:         "last",
					Label:        "Last",
					Description:  "Return the last item from an array",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
					Fields:       inputFields(),
					AutoLayout:   inputLayout(),
				},
				{
					Type:         "first",
					Label:        "First",
					Description:  "Return the first item from an array",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
					Fields:       inputFields(),
					AutoLayout:   inputLayout(),
				},
				{
					Type:         "get",
					Label:        "Get",
					Description:  "Return an item by index",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
					Fields: []node.NodeField{
						{Key: "input", Type: "string", Required: true, Label: "Input"},
						{Key: "index", Type: "number", Required: true, Label: "Index", Default: 0},
					},
					AutoLayout: [][]node.AutoLayout{
						{{Key: "input", Component: "textarea", Flex: 24, Options: map[string]any{"helper": "Array data or variable"}}},
						{{Key: "index", Component: "number", Flex: 24}},
					},
				},
				{
					Type:         "length",
					Label:        "Length",
					Description:  "Return the array length",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
					Fields:       inputFields(),
					AutoLayout:   inputLayout(),
				},
				{
					Type:         "join",
					Label:        "Join",
					Description:  "Join array items into text",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
					Fields: []node.NodeField{
						{Key: "input", Type: "string", Required: true, Label: "Input"},
						{Key: "separator", Type: "string", Required: false, Label: "Separator", Default: ""},
					},
					AutoLayout: [][]node.AutoLayout{
						{{Key: "input", Component: "textarea", Flex: 24, Options: map[string]any{"helper": "Array data or variable"}}},
						{{Key: "separator", Component: "input", Flex: 24}},
					},
				},
			},
			OutputHandles: []string{"success", "error"},
		},
		ExecutorFactory: func(action node.Action) node.NodeExecutor {
			return &ArrayNode{Action: action}
		},
	})
}

func inputFields() []node.NodeField {
	return []node.NodeField{{Key: "input", Type: "string", Required: true, Label: "Input"}}
}

func inputLayout() [][]node.AutoLayout {
	return [][]node.AutoLayout{
		{{Key: "input", Component: "textarea", Flex: 24, Options: map[string]any{"helper": "Array data or variable"}}},
	}
}

func (n *ArrayNode) Execute(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	select {
	case <-ctx.Stop:
		return node.NodeExecutionResult{}, node.ErrNodeStopped
	default:
	}

	action := n.Action.Type
	if action == "" {
		action = "last"
	}

	items, err := n.inputArray(ctx)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, err
	}

	var response any
	responseType := &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"}

	switch action {
	case "last":
		response, err = arrayLast(items)
	case "first":
		response, err = arrayFirst(items)
	case "get":
		response, err = n.arrayGet(items)
	case "length":
		response = len(items)
	case "join":
		response = n.arrayJoin(items)
		responseType = &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"}
	default:
		err = fmt.Errorf("unknown action: %s", action)
	}
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, err
	}

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: response,
		Type:     responseType,
	}, nil
}

func (n *ArrayNode) inputArray(ctx *node.NodeContext) ([]any, error) {
	inputRaw, _ := node.FieldValue(n.Action, "input").(string)
	if strings.TrimSpace(inputRaw) == "" {
		return nil, fmt.Errorf("input is required")
	}

	value, ok := exactVariableValue(inputRaw, ctx.Variables)
	if !ok {
		parsed := node.ParseTemplate(inputRaw, ctx.Variables)
		var decoded any
		if err := json.Unmarshal([]byte(parsed), &decoded); err != nil {
			return nil, fmt.Errorf("input must be an array")
		}
		value = decoded
	}

	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("input must be an array, got %T", value)
	}
	return items, nil
}

func exactVariableValue(input string, variables map[string]any) (any, bool) {
	trimmed := strings.TrimSpace(input)
	if !strings.HasPrefix(trimmed, "{{") || !strings.HasSuffix(trimmed, "}}") {
		return nil, false
	}

	name := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "{{"), "}}"))
	value, ok := variables[name]
	return value, ok
}

func arrayLast(items []any) (any, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("array is empty")
	}
	return items[len(items)-1], nil
}

func arrayFirst(items []any) (any, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("array is empty")
	}
	return items[0], nil
}

func (n *ArrayNode) arrayGet(items []any) (any, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("array is empty")
	}

	index, err := fieldInt(n.Action, "index")
	if err != nil {
		return nil, err
	}
	if index < 0 {
		index = len(items) + index
	}
	if index < 0 || index >= len(items) {
		return nil, fmt.Errorf("index %d out of range", index)
	}
	return items[index], nil
}

func (n *ArrayNode) arrayJoin(items []any) string {
	separator, _ := node.FieldValue(n.Action, "separator").(string)
	parts := make([]string, 0, len(items))
	for _, item := range items {
		switch v := item.(type) {
		case string:
			parts = append(parts, v)
		case nil:
			parts = append(parts, "")
		default:
			encoded, err := json.Marshal(v)
			if err != nil {
				parts = append(parts, fmt.Sprintf("%v", v))
			} else {
				parts = append(parts, string(encoded))
			}
		}
	}
	return strings.Join(parts, separator)
}

func fieldInt(action node.Action, key string) (int, error) {
	value := node.FieldValue(action, key)
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		var parsed int
		if _, err := fmt.Sscanf(v, "%d", &parsed); err != nil {
			return 0, fmt.Errorf("%s must be a number", key)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("%s must be a number", key)
	}
}
