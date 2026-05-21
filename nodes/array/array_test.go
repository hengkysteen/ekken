package array

import (
	"strings"
	"testing"

	"ekken/internal/features/workflow/node"
)

func TestArrayNodeExecute(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]any
		variables   map[string]any
		want        any
		wantType    string
		wantErr     bool
		errContains string
	}{
		{
			name:      "last from exact variable",
			config:    map[string]any{"type": "last", "input": "{{items}}"},
			variables: map[string]any{"items": []any{"first", "last"}},
			want:      "last",
			wantType:  "application/json",
		},
		{
			name:      "first",
			config:    map[string]any{"type": "first", "input": "{{items}}"},
			variables: map[string]any{"items": []any{"first", "last"}},
			want:      "first",
			wantType:  "application/json",
		},
		{
			name:      "get negative index",
			config:    map[string]any{"type": "get", "input": "{{items}}", "index": -1.0},
			variables: map[string]any{"items": []any{"first", "last"}},
			want:      "last",
			wantType:  "application/json",
		},
		{
			name:      "length",
			config:    map[string]any{"type": "length", "input": "{{items}}"},
			variables: map[string]any{"items": []any{"a", "b", "c"}},
			want:      3,
			wantType:  "application/json",
		},
		{
			name:      "join strings",
			config:    map[string]any{"type": "join", "input": "{{items}}", "separator": ""},
			variables: map[string]any{"items": []any{"", "1", " +", " 1 = **", "2**"}},
			want:      "1 + 1 = **2**",
			wantType:  "text/plain",
		},
		{
			name:      "parse json array input",
			config:    map[string]any{"type": "last", "input": "[{\"id\":1},{\"id\":2}]"},
			variables: map[string]any{},
			want:      map[string]any{"id": float64(2)},
			wantType:  "application/json",
		},
		{
			name:        "non array input",
			config:      map[string]any{"type": "last", "input": "{{value}}"},
			variables:   map[string]any{"value": "not array"},
			wantErr:     true,
			errContains: "input must be an array",
		},
		{
			name:        "empty array",
			config:      map[string]any{"type": "last", "input": "{{items}}"},
			variables:   map[string]any{"items": []any{}},
			wantErr:     true,
			errContains: "array is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &ArrayNode{Action: node.ActionFromMap(tt.config)}
			result, err := n.Execute(&node.NodeContext{
				Stop:      make(chan struct{}),
				Variables: tt.variables,
			})

			if tt.wantErr {
				if err == nil {
					t.Fatalf("Execute() expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("error = %v, want containing %q", err, tt.errContains)
				}
				if result.Handle != "error" {
					t.Fatalf("handle = %q, want error", result.Handle)
				}
				return
			}
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if result.Handle != "success" {
				t.Fatalf("handle = %q, want success", result.Handle)
			}
			if result.Type == nil || result.Type.Mime != tt.wantType {
				t.Fatalf("type = %#v, want %s", result.Type, tt.wantType)
			}
			if !deepEqual(result.Response, tt.want) {
				t.Fatalf("response = %#v, want %#v", result.Response, tt.want)
			}
		})
	}
}

func TestArraySpecHasResponseVars(t *testing.T) {
	spec, ok := node.GlobalRegistry.GetSpec("array")
	if !ok {
		t.Fatal("array spec not registered")
	}
	if spec.DefaultAction != "last" {
		t.Fatalf("default action = %q, want last", spec.DefaultAction)
	}
	for _, action := range spec.Actions {
		if !action.HasResponse {
			t.Fatalf("action %s HasResponse = false", action.Type)
		}
		if action.ResponseVar != "array."+action.Type+"_" {
			t.Fatalf("action %s ResponseVar = %q", action.Type, action.ResponseVar)
		}
		if len(action.AutoLayout) == 0 {
			t.Fatalf("action %s AutoLayout is empty", action.Type)
		}
	}
}

func deepEqual(a, b any) bool {
	switch av := a.(type) {
	case []any:
		bv, ok := b.([]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !deepEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, v := range av {
			if !deepEqual(v, bv[k]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
