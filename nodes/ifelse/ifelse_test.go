package ifelse

import (
	"context"
	"ekken/internal/features/workflow/node"
	"testing"
)

func TestIfElseNode_Execute(t *testing.T) {
	tests := []struct {
		name           string
		operand1       string
		operator       string
		operand2       string
		variables      map[string]any
		expectedHandle string
		expectedResult bool
	}{
		// Equality operators
		{
			name:           "equals true",
			operand1:       "hello",
			operator:       "==",
			operand2:       "hello",
			expectedHandle: "true",
			expectedResult: true,
		},
		{
			name:           "equals false",
			operand1:       "hello",
			operator:       "==",
			operand2:       "world",
			expectedHandle: "false",
			expectedResult: false,
		},
		{
			name:           "not equals true",
			operand1:       "hello",
			operator:       "!=",
			operand2:       "world",
			expectedHandle: "true",
			expectedResult: true,
		},
		{
			name:           "not equals false",
			operand1:       "hello",
			operator:       "!=",
			operand2:       "hello",
			expectedHandle: "false",
			expectedResult: false,
		},

		// Numeric comparisons
		{
			name:           "greater than true",
			operand1:       "10",
			operator:       ">",
			operand2:       "5",
			expectedHandle: "true",
			expectedResult: true,
		},
		{
			name:           "greater than false",
			operand1:       "5",
			operator:       ">",
			operand2:       "10",
			expectedHandle: "false",
			expectedResult: false,
		},
		{
			name:           "less than true",
			operand1:       "5",
			operator:       "<",
			operand2:       "10",
			expectedHandle: "true",
			expectedResult: true,
		},
		{
			name:           "greater or equals true (equal)",
			operand1:       "10",
			operator:       ">=",
			operand2:       "10",
			expectedHandle: "true",
			expectedResult: true,
		},
		{
			name:           "greater or equals true (greater)",
			operand1:       "15",
			operator:       ">=",
			operand2:       "10",
			expectedHandle: "true",
			expectedResult: true,
		},
		{
			name:           "less or equals true (equal)",
			operand1:       "10",
			operator:       "<=",
			operand2:       "10",
			expectedHandle: "true",
			expectedResult: true,
		},

		// String operations
		{
			name:           "contains true",
			operand1:       "hello world",
			operator:       "contains",
			operand2:       "world",
			expectedHandle: "true",
			expectedResult: true,
		},
		{
			name:           "contains false",
			operand1:       "hello world",
			operator:       "contains",
			operand2:       "foo",
			expectedHandle: "false",
			expectedResult: false,
		},
		{
			name:           "starts with true",
			operand1:       "hello world",
			operator:       "starts_with",
			operand2:       "hello",
			expectedHandle: "true",
			expectedResult: true,
		},
		{
			name:           "starts with false",
			operand1:       "hello world",
			operator:       "starts_with",
			operand2:       "world",
			expectedHandle: "false",
			expectedResult: false,
		},
		{
			name:           "ends with true",
			operand1:       "hello world",
			operator:       "ends_with",
			operand2:       "world",
			expectedHandle: "true",
			expectedResult: true,
		},
		{
			name:           "ends with false",
			operand1:       "hello world",
			operator:       "ends_with",
			operand2:       "hello",
			expectedHandle: "false",
			expectedResult: false,
		},

		// Empty checks
		{
			name:           "is empty true",
			operand1:       "",
			operator:       "is_empty",
			operand2:       "",
			expectedHandle: "true",
			expectedResult: true,
		},
		{
			name:           "is empty false",
			operand1:       "hello",
			operator:       "is_empty",
			operand2:       "",
			expectedHandle: "false",
			expectedResult: false,
		},
		{
			name:           "is not empty true",
			operand1:       "hello",
			operator:       "is_not_empty",
			operand2:       "",
			expectedHandle: "true",
			expectedResult: true,
		},
		{
			name:           "is not empty false",
			operand1:       "",
			operator:       "is_not_empty",
			operand2:       "",
			expectedHandle: "false",
			expectedResult: false,
		},

		// Variable substitution
		{
			name:           "variable substitution",
			operand1:       "{{name}}",
			operator:       "==",
			operand2:       "John",
			variables:      map[string]any{"name": "John"},
			expectedHandle: "true",
			expectedResult: true,
		},
		{
			name:           "numeric variable comparison",
			operand1:       "{{age}}",
			operator:       ">",
			operand2:       "18",
			variables:      map[string]any{"age": "25"},
			expectedHandle: "true",
			expectedResult: true,
		},

		// Default operator
		{
			name:           "default operator (empty)",
			operand1:       "hello",
			operator:       "",
			operand2:       "hello",
			expectedHandle: "true",
			expectedResult: true,
		},

		// String comparison fallback for non-numeric
		{
			name:           "string comparison with > operator",
			operand1:       "b",
			operator:       ">",
			operand2:       "a",
			expectedHandle: "true",
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &IfElseNode{
				Config: map[string]interface{}{
					"operand_1": tt.operand1,
					"operator":  tt.operator,
					"operand_2": tt.operand2,
				},
			}

			vars := tt.variables
			if vars == nil {
				vars = make(map[string]any)
			}

			ctx := &node.NodeContext{
				Context:   context.Background(),
				Variables: vars,
			}

			result, err := n.Execute(ctx)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Handle != tt.expectedHandle {
				t.Errorf("expected handle '%s', got '%s'", tt.expectedHandle, result.Handle)
			}

			if result.Response != tt.expectedResult {
				t.Errorf("expected response %v, got %v", tt.expectedResult, result.Response)
			}
		})
	}
}

func TestIfElseNode_UnknownOperator(t *testing.T) {
	n := &IfElseNode{
		Config: map[string]interface{}{
			"operand_1": "hello",
			"operator":  "invalid_op",
			"operand_2": "world",
		},
	}

	ctx := &node.NodeContext{
		Context:   context.Background(),
		Variables: make(map[string]any),
	}

	_, err := n.Execute(ctx)
	if err == nil {
		t.Error("expected error for unknown operator")
	}
	if err.Error() != "unknown operator: invalid_op" {
		t.Errorf("expected 'unknown operator: invalid_op', got '%s'", err.Error())
	}
}
