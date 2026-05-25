package mathnode

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"ekken/internal/features/workflow/node"
)

type MathNode struct {
	Action node.Action
}

func init() {
	node.GlobalRegistry.Register(node.NodeRegistration{
		Spec: node.Spec{
			Meta: node.Meta{
				Type:        "math",
				Label:       "Math",
				Icon:        "https://www.svgrepo.com/show/515242/calculator.svg",
				Tags:        []string{"System"},
				Description: "Basic arithmetic and math operations.",
			},

			DefaultAction: "calculate",
			Actions: []node.Action{
				{
					Type:         "calculate",
					Label:        "Calculate",
					Description:  "Perform basic arithmetic calculations",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
					Fields: []node.NodeField{
						{Key: "operand_1", Type: "string", Required: true, Label: "Operand 1"},
						{Key: "operator", Type: "string", Required: true, Label: "Operator", Default: "+", Options: []map[string]string{
							{"label": "+ (Add)", "value": "+"},
							{"label": "- (Subtract)", "value": "-"},
							{"label": "* (Multiply)", "value": "*"},
							{"label": "/ (Divide)", "value": "/"},
							{"label": "% (Modulo)", "value": "%"},
							{"label": "^ (Power)", "value": "^"},
						}},
						{Key: "operand_2", Type: "string", Required: true, Label: "Operand 2"},
					},
					AutoLayout: [][]node.AutoLayout{
						{
							{Key: "operand_1", Component: "input", Flex: 8, Options: map[string]any{"placeholder": "e.g. {{my_var}}"}},
							{Key: "operator", Component: "select", Flex: 8},
							{Key: "operand_2", Component: "input", Flex: 8, Options: map[string]any{"placeholder": "e.g. 10"}},
						},
					},
				},
				{
					Type:         "round",
					Label:        "Round",
					Description:  "Round a number",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
					Fields: []node.NodeField{
						{Key: "value", Type: "string", Required: true, Label: "Value"},
						{Key: "decimals", Type: "number", Default: 0, Label: "Decimal places"},
						{Key: "method", Type: "string", Default: "round", Label: "Method", Options: []map[string]string{
							{"label": "Round to nearest", "value": "round"},
							{"label": "Floor (down)", "value": "floor"},
							{"label": "Ceil (up)", "value": "ceil"},
						}},
					},
					AutoLayout: [][]node.AutoLayout{
						{
							{Key: "value", Component: "input", Flex: 12, Options: map[string]any{"placeholder": "e.g. 3.14159"}},
							{Key: "decimals", Component: "number", Flex: 6},
							{Key: "method", Component: "select", Flex: 6},
						},
					},
				},
				{
					Type:         "random",
					Label:        "Random",
					Description:  "Generate a random number within a range",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
					Fields: []node.NodeField{
						{Key: "min", Type: "number", Default: 0, Label: "Min"},
						{Key: "max", Type: "number", Default: 100, Label: "Max"},
						{Key: "type", Type: "string", Default: "integer", Label: "Type", Options: []map[string]string{
							{"label": "Integer", "value": "integer"},
							{"label": "Float", "value": "float"},
						}},
					},
					AutoLayout: [][]node.AutoLayout{
						{
							{Key: "min", Component: "number", Flex: 8},
							{Key: "max", Component: "number", Flex: 8},
							{Key: "type", Component: "select", Flex: 8},
						},
					},
				},
			},
			OutputHandles: []string{"success"},
		},

		ExecutorFactory: func(action node.Action) node.NodeExecutor {
			return &MathNode{Action: action}
		},
	})
}

func (n *MathNode) Execute(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	select {
	case <-ctx.Stop:
		return node.NodeExecutionResult{}, node.ErrNodeStopped
	default:
	}

	switch n.Action.Type {
	case "round":
		return n.executeRound(ctx)
	case "random":
		return n.executeRandom()
	default:
		return n.executeCalculate(ctx)
	}
}

func (n *MathNode) executeCalculate(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	op1Raw, _ := node.FieldValue(n.Action, "operand_1").(string)
	operator, _ := node.FieldValue(n.Action, "operator").(string)
	op2Raw, _ := node.FieldValue(n.Action, "operand_2").(string)

	if operator == "" {
		operator = "+"
	}

	op1Str := node.ParseTemplate(op1Raw, ctx.Variables)
	op2Str := node.ParseTemplate(op2Raw, ctx.Variables)

	num1, err := strconv.ParseFloat(strings.TrimSpace(op1Str), 64)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("invalid operand_1: %w", err)
	}
	num2, err := strconv.ParseFloat(strings.TrimSpace(op2Str), 64)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("invalid operand_2: %w", err)
	}

	var result float64
	switch operator {
	case "+":
		result = num1 + num2
	case "-":
		result = num1 - num2
	case "*":
		result = num1 * num2
	case "/":
		if num2 == 0 {
			return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("division by zero")
		}
		result = num1 / num2
	case "%":
		if num2 == 0 {
			return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("modulo by zero")
		}
		result = math.Mod(num1, num2)
	case "^":
		result = math.Pow(num1, num2)
	default:
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("unknown operator: %s", operator)
	}

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: result,
		Type:     &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
	}, nil
}

func (n *MathNode) executeRound(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	valRaw, _ := node.FieldValue(n.Action, "value").(string)
	decimalsF, _ := node.FieldValue(n.Action, "decimals").(float64)
	method, _ := node.FieldValue(n.Action, "method").(string)

	if method == "" {
		method = "round"
	}

	valStr := node.ParseTemplate(valRaw, ctx.Variables)
	val, err := strconv.ParseFloat(strings.TrimSpace(valStr), 64)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("invalid value: %w", err)
	}

	decimals := int(decimalsF)
	if decimals < 0 {
		decimals = 0
	}

	shift := math.Pow(10, float64(decimals))
	var result float64

	switch method {
	case "floor":
		result = math.Floor(val*shift) / shift
	case "ceil":
		result = math.Ceil(val*shift) / shift
	default: // "round"
		result = math.Round(val*shift) / shift
	}

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: result,
		Type:     &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
	}, nil
}

func (n *MathNode) executeRandom() (node.NodeExecutionResult, error) {
	minF, _ := node.FieldValue(n.Action, "min").(float64)
	maxF, _ := node.FieldValue(n.Action, "max").(float64)
	randomType, _ := node.FieldValue(n.Action, "type").(string)

	if randomType == "" {
		randomType = "integer"
	}

	if minF == 0 && maxF == 0 {
		maxF = 100
	}

	if minF > maxF {
		minF, maxF = maxF, minF
	}

	var result float64
	if randomType == "float" {
		result = minF + rand.Float64()*(maxF-minF)
	} else {
		minInt := int(math.Ceil(minF))
		maxInt := int(math.Floor(maxF))
		if minInt > maxInt {
			result = float64(minInt)
		} else {
			result = float64(minInt + rand.Intn(maxInt-minInt+1))
		}
	}

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: result,
		Type:     &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
	}, nil
}
