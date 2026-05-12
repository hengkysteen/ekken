package ifelse

import (
	"fmt"
	"strconv"
	"strings"

	"ekken/internal/features/workflow/node"
)

type IfElseNode struct {
	Action node.NodeAction
}

func init() {
	node.GlobalRegistry.Register(node.NodeRegistration{
		NodeSpec: node.NodeSpec{
			NodeMetadata: node.NodeMetadata{
				Type:        "ifelse",
				Tags:        []string{"Conditions"},
				Label:       "If Else",
				Icon:        "https://api.iconify.design/mdi/call-split.svg",
				Description: "Evaluate a condition with two outcomes (True/False).",
			},

			DefaultAction: "if_else",
			Actions: []node.NodeAction{
				{
					Key:   "if_else",
					Label: "IF Else",
					Fields: []node.NodeField{
						{Key: "operand_1", Type: "string", Label: "Operand 1"},
						{Key: "operator", Type: "string", Label: "Operator", Default: "==", Options: []map[string]string{
							{"label": "== (Equals)", "value": "=="},
							{"label": "!= (Not Equals)", "value": "!="},
							{"label": "> (Greater Than)", "value": ">"},
							{"label": "< (Less Than)", "value": "<"},
							{"label": ">= (Greater or Equals)", "value": ">="},
							{"label": "<= (Less or Equals)", "value": "<="},
							{"label": "Contains", "value": "contains"},
							{"label": "Starts With", "value": "starts_with"},
							{"label": "Ends With", "value": "ends_with"},
							{"label": "Is Empty", "value": "is_empty"},
							{"label": "Is Not Empty", "value": "is_not_empty"},
						}},
						{Key: "operand_2", Type: "string", Label: "Operand 2"},
					},
					AutoLayout: [][]node.AutoLayout{
						{
							{Key: "operand_1", Component: "input", Flex: 8, Options: map[string]any{"helper": "Value to compare", "placeholder": "e.g. {{variable}}"}},
							{Key: "operator", Component: "select", Flex: 8, Options: map[string]any{"helper": "Select comparison method"}},
							{Key: "operand_2", Component: "input", Flex: 8, Options: map[string]any{"helper": "Value to compare against", "placeholder": "e.g. value"}},
						},
					},
				},
			},
			Outputs: []node.HandleEdge{
				{Key: "true", Label: "True", Tone: "success"},
				{Key: "false", Label: "False", Tone: "warning"},
			},
		},
		ExecutorFactory: func(action node.NodeAction) node.NodeExecutor {
			return &IfElseNode{Action: action}
		},
	})
}

func (n *IfElseNode) Execute(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	op1Raw, _ := node.FieldValue(n.Action, "operand_1").(string)
	operator, _ := node.FieldValue(n.Action, "operator").(string)
	op2Raw, _ := node.FieldValue(n.Action, "operand_2").(string)

	if operator == "" {
		operator = "=="
	}

	op1 := node.ParseTemplate(op1Raw, ctx.Variables)
	op2 := node.ParseTemplate(op2Raw, ctx.Variables)

	result := false

	switch operator {
	case "==":
		result = op1 == op2
	case "!=":
		result = op1 != op2
	case "contains":
		result = strings.Contains(op1, op2)
	case "starts_with":
		result = strings.HasPrefix(op1, op2)
	case "ends_with":
		result = strings.HasSuffix(op1, op2)
	case "is_empty":
		result = op1 == ""
	case "is_not_empty":
		result = op1 != ""
	case ">", "<", ">=", "<=":
		num1, err1 := strconv.ParseFloat(op1, 64)
		num2, err2 := strconv.ParseFloat(op2, 64)
		if err1 == nil && err2 == nil {
			switch operator {
			case ">":
				result = num1 > num2
			case "<":
				result = num1 < num2
			case ">=":
				result = num1 >= num2
			case "<=":
				result = num1 <= num2
			}
		} else {
			switch operator {
			case ">":
				result = op1 > op2
			case "<":
				result = op1 < op2
			case ">=":
				result = op1 >= op2
			case "<=":
				result = op1 <= op2
			}
		}
	default:
		return node.NodeExecutionResult{}, fmt.Errorf("unknown operator: %s", operator)
	}

	handle := "false"
	if result {
		handle = "true"
	}

	return node.NodeExecutionResult{
		Handle:   handle,
		Response: result,
	}, nil
}
