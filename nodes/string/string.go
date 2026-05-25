package string

import (
	"ekken/internal/features/workflow/node"
	"fmt"
	"regexp"
	"strings"
)

type StringNode struct {
	Action node.Action
}

func init() {
	node.GlobalRegistry.Register(node.NodeRegistration{
		Spec: node.Spec{
			Meta: node.Meta{
				Type:        "string",
				Label:       "String",
				Icon:        "https://www.svgrepo.com/show/451357/string.svg",
				Tags:        []string{"System"},
				Description: "String manipulation operations.",
			},
			DefaultAction: "concat",
			Actions: []node.Action{
				{
					Type:         "split",
					Label:        "Split",
					Description:  "Split string into array by delimiter",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
					Fields: []node.NodeField{
						{Key: "input", Type: "string", Required: true, Label: "Input"},
						{Key: "delimiter", Type: "string", Required: true, Label: "Delimiter", Default: ","},
					},
					AutoLayout: [][]node.AutoLayout{
						{{Key: "input", Component: "textarea", Flex: 24, Options: map[string]any{"helper": "String to split"}}},
						{{Key: "delimiter", Component: "input", Flex: 24, Options: map[string]any{"placeholder": ","}}},
					},
				},
				{
					Type:         "replace",
					Label:        "Replace",
					Description:  "Replace substring with another string",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
					Fields: []node.NodeField{
						{Key: "input", Type: "string", Required: true, Label: "Input"},
						{Key: "old", Type: "string", Required: true, Label: "Find"},
						{Key: "new", Type: "string", Required: true, Label: "Replace with"},
						{Key: "count", Type: "number", Default: -1, Label: "Max replacements"},
					},
					AutoLayout: [][]node.AutoLayout{
						{{Key: "input", Component: "textarea", Flex: 24}},
						{{Key: "old", Component: "input", Flex: 12}, {Key: "new", Component: "input", Flex: 12}},
						{{Key: "count", Component: "number", Flex: 24, Options: map[string]any{"helper": "-1 for all occurrences"}}},
					},
				},
				{
					Type:         "trim",
					Label:        "Trim",
					Description:  "Remove leading and trailing whitespace",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
					Fields: []node.NodeField{
						{Key: "input", Type: "string", Required: true, Label: "Input"},
					},
					AutoLayout: [][]node.AutoLayout{
						{{Key: "input", Component: "textarea", Flex: 24}},
					},
				},
				{
					Type:         "concat",
					Label:        "Concat",
					Description:  "Concatenate multiple strings",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
					Fields: []node.NodeField{
						{Key: "strings", Type: "string", Required: true, Label: "Strings"},
						{Key: "separator", Type: "string", Default: "", Label: "Separator"},
					},
					AutoLayout: [][]node.AutoLayout{
						{{Key: "strings", Component: "textarea", Flex: 24, Options: map[string]any{"helper": "One string per line"}}},
						{{Key: "separator", Component: "input", Flex: 24, Options: map[string]any{"placeholder": "Optional separator"}}},
					},
				},
				{
					Type:         "substring",
					Label:        "Substring",
					Description:  "Extract substring by start and end index",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
					Fields: []node.NodeField{
						{Key: "input", Type: "string", Required: true, Label: "Input"},
						{Key: "start", Type: "number", Default: 0, Label: "Start index"},
						{Key: "end", Type: "number", Default: -1, Label: "End index"},
					},
					AutoLayout: [][]node.AutoLayout{
						{{Key: "input", Component: "textarea", Flex: 24}},
						{{Key: "start", Component: "number", Flex: 12}, {Key: "end", Component: "number", Flex: 12, Options: map[string]any{"helper": "-1 for end of string"}}},
					},
				},
				{
					Type:         "to_upper",
					Label:        "To Upper",
					Description:  "Convert string to uppercase",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
					Fields: []node.NodeField{
						{Key: "input", Type: "string", Required: true, Label: "Input"},
					},
					AutoLayout: [][]node.AutoLayout{
						{{Key: "input", Component: "textarea", Flex: 24}},
					},
				},
				{
					Type:         "to_lower",
					Label:        "To Lower",
					Description:  "Convert string to lowercase",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
					Fields: []node.NodeField{
						{Key: "input", Type: "string", Required: true, Label: "Input"},
					},
					AutoLayout: [][]node.AutoLayout{
						{{Key: "input", Component: "textarea", Flex: 24}},
					},
				},
				{
					Type:         "regex_match",
					Label:        "Regex Match",
					Description:  "Match string against regex pattern",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
					Fields: []node.NodeField{
						{Key: "input", Type: "string", Required: true, Label: "Input"},
						{Key: "pattern", Type: "string", Required: true, Label: "Regex pattern"},
					},
					AutoLayout: [][]node.AutoLayout{
						{{Key: "input", Component: "textarea", Flex: 24}},
						{{Key: "pattern", Component: "input", Flex: 24, Options: map[string]any{"placeholder": "e.g. \\d+"}}},
					},
				},
				{
					Type:         "regex_replace",
					Label:        "Regex Replace",
					Description:  "Replace string using regex pattern",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
					Fields: []node.NodeField{
						{Key: "input", Type: "string", Required: true, Label: "Input"},
						{Key: "pattern", Type: "string", Required: true, Label: "Regex pattern"},
						{Key: "replacement", Type: "string", Required: true, Label: "Replacement"},
					},
					AutoLayout: [][]node.AutoLayout{
						{{Key: "input", Component: "textarea", Flex: 24}},
						{{Key: "pattern", Component: "input", Flex: 12}, {Key: "replacement", Component: "input", Flex: 12}},
					},
				},
				{
					Type:         "length",
					Label:        "Length",
					Description:  "Return string length",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
					Fields: []node.NodeField{
						{Key: "input", Type: "string", Required: true, Label: "Input"},
					},
					AutoLayout: [][]node.AutoLayout{
						{{Key: "input", Component: "textarea", Flex: 24}},
					},
				},
			},
			OutputHandles: []string{"success", "error"},
		},
		ExecutorFactory: func(action node.Action) node.NodeExecutor {
			return &StringNode{Action: action}
		},
	})
}

func (n *StringNode) Execute(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	select {
	case <-ctx.Stop:
		return node.NodeExecutionResult{}, node.ErrNodeStopped
	default:
	}

	switch n.Action.Type {
	case "split":
		return n.executeSplit(ctx)
	case "replace":
		return n.executeReplace(ctx)
	case "trim":
		return n.executeTrim(ctx)
	case "concat":
		return n.executeConcat(ctx)
	case "substring":
		return n.executeSubstring(ctx)
	case "to_upper":
		return n.executeToUpper(ctx)
	case "to_lower":
		return n.executeToLower(ctx)
	case "regex_match":
		return n.executeRegexMatch(ctx)
	case "regex_replace":
		return n.executeRegexReplace(ctx)
	case "length":
		return n.executeLength(ctx)
	default:
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("unknown action: %s", n.Action.Type)
	}
}

func (n *StringNode) executeSplit(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	inputRaw, _ := node.FieldValue(n.Action, "input").(string)
	delimiter, _ := node.FieldValue(n.Action, "delimiter").(string)

	input := node.ParseTemplate(inputRaw, ctx.Variables)
	delimiter = node.ParseTemplate(delimiter, ctx.Variables)

	if delimiter == "" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("delimiter is required")
	}

	parts := strings.Split(input, delimiter)
	result := make([]any, len(parts))
	for i, part := range parts {
		result[i] = part
	}

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: result,
		Type:     &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
	}, nil
}

func (n *StringNode) executeReplace(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	inputRaw, _ := node.FieldValue(n.Action, "input").(string)
	oldRaw, _ := node.FieldValue(n.Action, "old").(string)
	newRaw, _ := node.FieldValue(n.Action, "new").(string)
	count := -1
	if c, ok := node.FieldValue(n.Action, "count").(float64); ok {
		count = int(c)
	}

	input := node.ParseTemplate(inputRaw, ctx.Variables)
	old := node.ParseTemplate(oldRaw, ctx.Variables)
	new := node.ParseTemplate(newRaw, ctx.Variables)

	result := strings.Replace(input, old, new, count)

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: result,
		Type:     &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
	}, nil
}

func (n *StringNode) executeTrim(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	inputRaw, _ := node.FieldValue(n.Action, "input").(string)
	input := node.ParseTemplate(inputRaw, ctx.Variables)
	result := strings.TrimSpace(input)

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: result,
		Type:     &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
	}, nil
}

func (n *StringNode) executeConcat(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	stringsRaw, _ := node.FieldValue(n.Action, "strings").(string)
	separatorRaw, _ := node.FieldValue(n.Action, "separator").(string)

	stringsInput := node.ParseTemplate(stringsRaw, ctx.Variables)
	separator := node.ParseTemplate(separatorRaw, ctx.Variables)

	lines := strings.Split(stringsInput, "\n")
	result := strings.Join(lines, separator)

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: result,
		Type:     &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
	}, nil
}

func (n *StringNode) executeSubstring(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	inputRaw, _ := node.FieldValue(n.Action, "input").(string)
	start := 0
	if s, ok := node.FieldValue(n.Action, "start").(float64); ok {
		start = int(s)
	}
	end := -1
	if e, ok := node.FieldValue(n.Action, "end").(float64); ok {
		end = int(e)
	}

	input := node.ParseTemplate(inputRaw, ctx.Variables)

	if start < 0 {
		start = 0
	}
	if start > len(input) {
		start = len(input)
	}

	if end == -1 || end > len(input) {
		end = len(input)
	}
	if end < start {
		end = start
	}

	result := input[start:end]

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: result,
		Type:     &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
	}, nil
}

func (n *StringNode) executeToUpper(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	inputRaw, _ := node.FieldValue(n.Action, "input").(string)
	input := node.ParseTemplate(inputRaw, ctx.Variables)
	result := strings.ToUpper(input)

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: result,
		Type:     &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
	}, nil
}

func (n *StringNode) executeToLower(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	inputRaw, _ := node.FieldValue(n.Action, "input").(string)
	input := node.ParseTemplate(inputRaw, ctx.Variables)
	result := strings.ToLower(input)

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: result,
		Type:     &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
	}, nil
}

func (n *StringNode) executeRegexMatch(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	inputRaw, _ := node.FieldValue(n.Action, "input").(string)
	patternRaw, _ := node.FieldValue(n.Action, "pattern").(string)

	input := node.ParseTemplate(inputRaw, ctx.Variables)
	pattern := node.ParseTemplate(patternRaw, ctx.Variables)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("invalid regex pattern: %v", err)
	}

	matches := re.FindAllString(input, -1)
	result := make([]any, len(matches))
	for i, match := range matches {
		result[i] = match
	}

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: result,
		Type:     &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
	}, nil
}

func (n *StringNode) executeRegexReplace(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	inputRaw, _ := node.FieldValue(n.Action, "input").(string)
	patternRaw, _ := node.FieldValue(n.Action, "pattern").(string)
	replacementRaw, _ := node.FieldValue(n.Action, "replacement").(string)

	input := node.ParseTemplate(inputRaw, ctx.Variables)
	pattern := node.ParseTemplate(patternRaw, ctx.Variables)
	replacement := node.ParseTemplate(replacementRaw, ctx.Variables)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("invalid regex pattern: %v", err)
	}

	result := re.ReplaceAllString(input, replacement)

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: result,
		Type:     &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
	}, nil
}

func (n *StringNode) executeLength(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	inputRaw, _ := node.FieldValue(n.Action, "input").(string)
	input := node.ParseTemplate(inputRaw, ctx.Variables)
	result := len(input)

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: result,
		Type:     &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
	}, nil
}
