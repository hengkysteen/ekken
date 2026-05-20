package json

import (
	"ekken/internal/features/workflow/node"
	"testing"
)

func TestJsonNode_Execute(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]any
		input       any
		wantHandle  string
		wantResult  any
		wantErr     bool
		errContains string
	}{
		{
			name:       "simple key access",
			config:     map[string]any{"input": "{{my_input}}", "path": "name"},
			input:      map[string]any{"name": "John", "age": 30},
			wantHandle: "success",
			wantResult: "John",
		},
		{
			name:       "nested key access",
			config:     map[string]any{"input": "{{my_input}}", "path": "user.name"},
			input:      map[string]any{"user": map[string]any{"name": "Jane", "age": 25}},
			wantHandle: "success",
			wantResult: "Jane",
		},
		{
			name:       "array index access",
			config:     map[string]any{"input": "{{my_input}}", "path": "items[0]"},
			input:      map[string]any{"items": []any{"first", "second", "third"}},
			wantHandle: "success",
			wantResult: "first",
		},
		{
			name:       "nested array access",
			config:     map[string]any{"input": "{{my_input}}", "path": "data.items[1]"},
			input:      map[string]any{"data": map[string]any{"items": []any{"a", "b", "c"}}},
			wantHandle: "success",
			wantResult: "b",
		},
		{
			name:   "complex path - OpenAI response",
			config: map[string]any{"input": "{{my_input}}", "path": "choices[0].message.content"},
			input: map[string]any{
				"choices": []any{
					map[string]any{
						"message": map[string]any{
							"content": "Hello, world!",
							"role":    "assistant",
						},
					},
				},
			},
			wantHandle: "success",
			wantResult: "Hello, world!",
		},
		{
			name:       "multiple array indices",
			config:     map[string]any{"input": "{{my_input}}", "path": "matrix[0]"},
			input:      map[string]any{"matrix": []any{[]any{1, 2, 3}, []any{4, 5, 6}}},
			wantHandle: "success",
			wantResult: []any{1, 2, 3},
		},
		{
			name:        "empty path",
			config:      map[string]any{"input": "{{my_input}}", "path": ""},
			input:       map[string]any{"key": "value"},
			wantErr:     true,
			errContains: "path is required",
		},
		{
			name:        "nil input",
			config:      map[string]any{"input": "{{my_input}}", "path": "key"},
			input:       nil,
			wantErr:     true,
			errContains: "cannot access",
		},
		{
			name:        "key not found",
			config:      map[string]any{"input": "{{my_input}}", "path": "missing"},
			input:       map[string]any{"key": "value"},
			wantHandle:  "error",
			wantErr:     true,
			errContains: "key 'missing' not found",
		},
		{
			name:        "nested key not found",
			config:      map[string]any{"input": "{{my_input}}", "path": "user.missing"},
			input:       map[string]any{"user": map[string]any{"name": "John"}},
			wantHandle:  "error",
			wantErr:     true,
			errContains: "key 'missing' not found",
		},
		{
			name:        "index out of range",
			config:      map[string]any{"input": "{{my_input}}", "path": "items[5]"},
			input:       map[string]any{"items": []any{"a", "b"}},
			wantHandle:  "error",
			wantErr:     true,
			errContains: "index 5 out of range",
		},
		{
			name:        "negative index",
			config:      map[string]any{"input": "{{my_input}}", "path": "items[-1]"},
			input:       map[string]any{"items": []any{"a", "b"}},
			wantHandle:  "error",
			wantErr:     true,
			errContains: "index -1 out of range",
		},
		{
			name:        "invalid index format",
			config:      map[string]any{"input": "{{my_input}}", "path": "items[abc]"},
			input:       map[string]any{"items": []any{"a", "b"}},
			wantHandle:  "error",
			wantErr:     true,
			errContains: "invalid index",
		},
		{
			name:        "access array on non-array",
			config:      map[string]any{"input": "{{my_input}}", "path": "name[0]"},
			input:       map[string]any{"name": "John"},
			wantHandle:  "error",
			wantErr:     true,
			errContains: "expected array",
		},
		{
			name:        "access key on non-object",
			config:      map[string]any{"input": "{{my_input}}", "path": "items.name"},
			input:       map[string]any{"items": []any{"a", "b"}},
			wantHandle:  "error",
			wantErr:     true,
			errContains: "expected object",
		},
		{
			name:        "access on nil value",
			config:      map[string]any{"input": "{{my_input}}", "path": "user.name"},
			input:       map[string]any{"user": nil},
			wantHandle:  "error",
			wantErr:     true,
			errContains: "cannot access",
		},
		{
			name:   "deep nesting",
			config: map[string]any{"input": "{{my_input}}", "path": "a.b.c.d.e"},
			input: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": map[string]any{
							"d": map[string]any{
								"e": "deep value",
							},
						},
					},
				},
			},
			wantHandle: "success",
			wantResult: "deep value",
		},
		{
			name:       "number value",
			config:     map[string]any{"input": "{{my_input}}", "path": "count"},
			input:      map[string]any{"count": 42},
			wantHandle: "success",
			wantResult: 42,
		},
		{
			name:       "boolean value",
			config:     map[string]any{"input": "{{my_input}}", "path": "active"},
			input:      map[string]any{"active": true},
			wantHandle: "success",
			wantResult: true,
		},
		{
			name:       "null value",
			config:     map[string]any{"input": "{{my_input}}", "path": "nullable"},
			input:      map[string]any{"nullable": nil},
			wantHandle: "success",
			wantResult: nil,
		},
		{
			name:       "array as final result",
			config:     map[string]any{"input": "{{my_input}}", "path": "items"},
			input:      map[string]any{"items": []any{1, 2, 3}},
			wantHandle: "success",
			wantResult: []any{1, 2, 3},
		},
		{
			name:       "object as final result",
			config:     map[string]any{"input": "{{my_input}}", "path": "user"},
			input:      map[string]any{"user": map[string]any{"name": "John", "age": 30}},
			wantHandle: "success",
			wantResult: map[string]any{"name": "John", "age": 30},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &JsonNode{Action: node.ActionFromMap(tt.config)}
			ctx := &node.NodeContext{
				Stop:      make(chan struct{}),
				Variables: map[string]interface{}{"my_input": tt.input},
			}

			result, err := n.Execute(ctx)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Execute() expected error, got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Execute() error = %v, want error containing %v", err, tt.errContains)
				}
				if tt.wantHandle != "" && result.Handle != tt.wantHandle {
					t.Errorf("Execute() handle = %v, want %v", result.Handle, tt.wantHandle)
				}
				return
			}

			if err != nil {
				t.Errorf("Execute() unexpected error = %v", err)
				return
			}

			if result.Handle != tt.wantHandle {
				t.Errorf("Execute() handle = %v, want %v", result.Handle, tt.wantHandle)
			}

			if !deepEqual(result.Response, tt.wantResult) {
				t.Errorf("Execute() response = %v, want %v", result.Response, tt.wantResult)
			}

			if result.Type == nil || result.Type.Mime != "application/json" {
				t.Errorf("Execute() type = %v, want application/json", result.Type)
			}
		})
	}
}

func TestJsonNode_ExtractStream(t *testing.T) {
	tests := []struct {
		name       string
		input      any
		path       string
		wantResult any
	}{
		{
			name: "sse data",
			input: "data: {\"delta\":\"Hello\"}\n\n" +
				"data: {\"delta\":\" world\"}\n\n" +
				"data: [DONE]\n\n",
			path:       "delta",
			wantResult: []any{"Hello", " world"},
		},
		{
			name:       "jsonl",
			input:      "{\"id\":1,\"name\":\"Alice\"}\n{\"id\":2,\"name\":\"Bob\"}\n",
			path:       "name",
			wantResult: []any{"Alice", "Bob"},
		},
		{
			name:       "array",
			input:      []any{map[string]any{"id": 1}, map[string]any{"id": 2}},
			path:       "id",
			wantResult: []any{1, 2},
		},
		{
			name: "http response body sse with body data path",
			input: map[string]any{
				"body": "data: {\"choices\":[{\"delta\":{\"content\":\"1\"}}]}\n\n" +
					"data: {\"choices\":[{\"delta\":{\"content\":\" +\"}}]}\n\n" +
					"data: [DONE]\n\n",
			},
			path: "body.data.choices[0]",
			wantResult: []any{
				map[string]any{"delta": map[string]any{"content": "1"}},
				map[string]any{"delta": map[string]any{"content": " +"}},
			},
		},
		{
			name: "empty path returns stream items",
			input: "data: {\"id\":1}\n\n" +
				"data: {\"id\":2}\n\n",
			wantResult: []any{map[string]any{"id": float64(1)}, map[string]any{"id": float64(2)}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &JsonNode{Action: node.ActionFromMap(map[string]any{
				"action": "extract_stream",
				"input":  "{{my_input}}",
				"path":   tt.path,
			})}
			ctx := &node.NodeContext{
				Stop:      make(chan struct{}),
				Variables: map[string]interface{}{"my_input": tt.input},
			}

			result, err := n.Execute(ctx)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if result.Handle != "success" {
				t.Fatalf("handle = %q, want success", result.Handle)
			}
			if !deepEqual(result.Response, tt.wantResult) {
				t.Fatalf("response = %#v, want %#v", result.Response, tt.wantResult)
			}
		})
	}
}

func TestTraverse(t *testing.T) {
	tests := []struct {
		name    string
		data    any
		path    string
		want    any
		wantErr bool
	}{
		{
			name: "empty path returns data",
			data: map[string]any{"key": "value"},
			path: "",
			want: map[string]any{"key": "value"},
		},
		{
			name: "single key",
			data: map[string]any{"name": "Alice"},
			path: "name",
			want: "Alice",
		},
		{
			name: "nested keys",
			data: map[string]any{"user": map[string]any{"profile": map[string]any{"email": "test@example.com"}}},
			path: "user.profile.email",
			want: "test@example.com",
		},
		{
			name: "array index",
			data: map[string]any{"colors": []any{"red", "green", "blue"}},
			path: "colors[1]",
			want: "green",
		},
		{
			name: "last array element",
			data: map[string]any{"nums": []any{10, 20, 30}},
			path: "nums[2]",
			want: 30,
		},
		{
			name:    "nil data with path",
			data:    nil,
			path:    "key",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := traverse(tt.data, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("traverse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !deepEqual(got, tt.want) {
				t.Errorf("traverse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		path string
		want []string
	}{
		{"name", []string{"name"}},
		{"user.name", []string{"user", "name"}},
		{"data.items[0]", []string{"data", "items[0]"}},
		{"choices[0].message.content", []string{"choices[0]", "message", "content"}},
		{"a.b.c.d", []string{"a", "b", "c", "d"}},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := splitPath(tt.path)
			if len(got) != len(tt.want) {
				t.Errorf("splitPath() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitPath()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func deepEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch va := a.(type) {
	case []any:
		vb, ok := b.([]any)
		if !ok || len(va) != len(vb) {
			return false
		}
		for i := range va {
			if !deepEqual(va[i], vb[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		vb, ok := b.(map[string]any)
		if !ok || len(va) != len(vb) {
			return false
		}
		for k, v := range va {
			if !deepEqual(v, vb[k]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
