package string

import (
	"ekken/internal/features/workflow/node"
	"strings"
	"testing"
)

func TestStringSplit(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]any
		vars    map[string]any
		want    []any
		wantErr bool
	}{
		{
			name:   "basic split",
			config: map[string]any{"type": "split", "input": "hello,world,test", "delimiter": ","},
			want:   []any{"hello", "world", "test"},
		},
		{
			name:   "split with space",
			config: map[string]any{"type": "split", "input": "hello world test", "delimiter": " "},
			want:   []any{"hello", "world", "test"},
		},
		{
			name:   "split with variable",
			config: map[string]any{"type": "split", "input": "{{text}}", "delimiter": ","},
			vars:   map[string]any{"text": "a,b,c"},
			want:   []any{"a", "b", "c"},
		},
		{
			name:    "empty delimiter error",
			config:  map[string]any{"type": "split", "input": "test", "delimiter": ""},
			wantErr: true,
		},
		{
			name:   "single item",
			config: map[string]any{"type": "split", "input": "hello", "delimiter": ","},
			want:   []any{"hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &StringNode{Action: node.ActionFromMap(tt.config)}
			ctx := &node.NodeContext{
				Variables: tt.vars,
				Stop:      make(chan struct{}),
			}

			result, err := n.Execute(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result.Handle != "success" {
				t.Errorf("Handle = %s, want success", result.Handle)
			}

			if !tt.wantErr {
				got, ok := result.Response.([]any)
				if !ok {
					t.Errorf("Response is not []any, got %T", result.Response)
					return
				}

				if !deepEqual(got, tt.want) {
					t.Errorf("got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestStringReplace(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]any
		vars    map[string]any
		want    string
		wantErr bool
	}{
		{
			name:   "basic replace",
			config: map[string]any{"type": "replace", "input": "hello world", "old": "world", "new": "universe", "count": -1.0},
			want:   "hello universe",
		},
		{
			name:   "replace all",
			config: map[string]any{"type": "replace", "input": "test test test", "old": "test", "new": "demo", "count": -1.0},
			want:   "demo demo demo",
		},
		{
			name:   "replace with count",
			config: map[string]any{"type": "replace", "input": "test test test", "old": "test", "new": "demo", "count": 2.0},
			want:   "demo demo test",
		},
		{
			name:   "replace with variable",
			config: map[string]any{"type": "replace", "input": "{{text}}", "old": "old", "new": "new", "count": -1.0},
			vars:   map[string]any{"text": "old value old"},
			want:   "new value new",
		},
		{
			name:   "no match",
			config: map[string]any{"type": "replace", "input": "hello world", "old": "xyz", "new": "abc", "count": -1.0},
			want:   "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &StringNode{Action: node.ActionFromMap(tt.config)}
			ctx := &node.NodeContext{
				Variables: tt.vars,
				Stop:      make(chan struct{}),
			}

			result, err := n.Execute(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, ok := result.Response.(string)
				if !ok {
					t.Errorf("Response is not string, got %T", result.Response)
					return
				}

				if got != tt.want {
					t.Errorf("got = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func TestStringTrim(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]any
		vars   map[string]any
		want   string
	}{
		{
			name:   "trim spaces",
			config: map[string]any{"type": "trim", "input": "  hello world  "},
			want:   "hello world",
		},
		{
			name:   "trim tabs and newlines",
			config: map[string]any{"type": "trim", "input": "\t\nhello\n\t"},
			want:   "hello",
		},
		{
			name:   "no trim needed",
			config: map[string]any{"type": "trim", "input": "hello"},
			want:   "hello",
		},
		{
			name:   "trim with variable",
			config: map[string]any{"type": "trim", "input": "{{text}}"},
			vars:   map[string]any{"text": "  test  "},
			want:   "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &StringNode{Action: node.ActionFromMap(tt.config)}
			ctx := &node.NodeContext{
				Variables: tt.vars,
				Stop:      make(chan struct{}),
			}

			result, err := n.Execute(ctx)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}

			got, ok := result.Response.(string)
			if !ok {
				t.Errorf("Response is not string, got %T", result.Response)
				return
			}

			if got != tt.want {
				t.Errorf("got = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStringConcat(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]any
		vars   map[string]any
		want   string
	}{
		{
			name:   "concat with comma",
			config: map[string]any{"type": "concat", "strings": "hello\nworld\ntest", "separator": ","},
			want:   "hello,world,test",
		},
		{
			name:   "concat with space",
			config: map[string]any{"type": "concat", "strings": "hello\nworld", "separator": " "},
			want:   "hello world",
		},
		{
			name:   "concat no separator",
			config: map[string]any{"type": "concat", "strings": "hello\nworld", "separator": ""},
			want:   "helloworld",
		},
		{
			name:   "concat with variable",
			config: map[string]any{"type": "concat", "strings": "{{a}}\n{{b}}", "separator": "-"},
			vars:   map[string]any{"a": "first", "b": "second"},
			want:   "first-second",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &StringNode{Action: node.ActionFromMap(tt.config)}
			ctx := &node.NodeContext{
				Variables: tt.vars,
				Stop:      make(chan struct{}),
			}

			result, err := n.Execute(ctx)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}

			got, ok := result.Response.(string)
			if !ok {
				t.Errorf("Response is not string, got %T", result.Response)
				return
			}

			if got != tt.want {
				t.Errorf("got = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStringSubstring(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]any
		vars   map[string]any
		want   string
	}{
		{
			name:   "basic substring",
			config: map[string]any{"type": "substring", "input": "hello world", "start": 0.0, "end": 5.0},
			want:   "hello",
		},
		{
			name:   "substring middle",
			config: map[string]any{"type": "substring", "input": "hello world", "start": 6.0, "end": 11.0},
			want:   "world",
		},
		{
			name:   "substring to end",
			config: map[string]any{"type": "substring", "input": "hello world", "start": 6.0, "end": -1.0},
			want:   "world",
		},
		{
			name:   "substring with variable",
			config: map[string]any{"type": "substring", "input": "{{text}}", "start": 0.0, "end": 4.0},
			vars:   map[string]any{"text": "testing"},
			want:   "test",
		},
		{
			name:   "out of bounds",
			config: map[string]any{"type": "substring", "input": "hello", "start": 10.0, "end": 20.0},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &StringNode{Action: node.ActionFromMap(tt.config)}
			ctx := &node.NodeContext{
				Variables: tt.vars,
				Stop:      make(chan struct{}),
			}

			result, err := n.Execute(ctx)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}

			got, ok := result.Response.(string)
			if !ok {
				t.Errorf("Response is not string, got %T", result.Response)
				return
			}

			if got != tt.want {
				t.Errorf("got = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStringCase(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]any
		vars   map[string]any
		want   string
	}{
		{
			name:   "to upper",
			config: map[string]any{"type": "to_upper", "input": "hello world"},
			want:   "HELLO WORLD",
		},
		{
			name:   "to lower",
			config: map[string]any{"type": "to_lower", "input": "HELLO WORLD"},
			want:   "hello world",
		},
		{
			name:   "to upper with variable",
			config: map[string]any{"type": "to_upper", "input": "{{text}}"},
			vars:   map[string]any{"text": "test"},
			want:   "TEST",
		},
		{
			name:   "to lower mixed case",
			config: map[string]any{"type": "to_lower", "input": "HeLLo WoRLd"},
			want:   "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &StringNode{Action: node.ActionFromMap(tt.config)}
			ctx := &node.NodeContext{
				Variables: tt.vars,
				Stop:      make(chan struct{}),
			}

			result, err := n.Execute(ctx)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}

			got, ok := result.Response.(string)
			if !ok {
				t.Errorf("Response is not string, got %T", result.Response)
				return
			}

			if got != tt.want {
				t.Errorf("got = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStringRegexMatch(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]any
		vars    map[string]any
		want    []any
		wantErr bool
	}{
		{
			name:   "match digits",
			config: map[string]any{"type": "regex_match", "input": "test 123 and 456", "pattern": `\d+`},
			want:   []any{"123", "456"},
		},
		{
			name:   "match words",
			config: map[string]any{"type": "regex_match", "input": "hello world test", "pattern": `\w+`},
			want:   []any{"hello", "world", "test"},
		},
		{
			name:   "no match",
			config: map[string]any{"type": "regex_match", "input": "hello world", "pattern": `\d+`},
			want:   []any{},
		},
		{
			name:    "invalid regex",
			config:  map[string]any{"type": "regex_match", "input": "test", "pattern": `[`},
			wantErr: true,
		},
		{
			name:   "match with variable",
			config: map[string]any{"type": "regex_match", "input": "{{text}}", "pattern": `\d+`},
			vars:   map[string]any{"text": "id: 999"},
			want:   []any{"999"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &StringNode{Action: node.ActionFromMap(tt.config)}
			ctx := &node.NodeContext{
				Variables: tt.vars,
				Stop:      make(chan struct{}),
			}

			result, err := n.Execute(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, ok := result.Response.([]any)
				if !ok {
					t.Errorf("Response is not []any, got %T", result.Response)
					return
				}

				if !deepEqual(got, tt.want) {
					t.Errorf("got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestStringRegexReplace(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]any
		vars    map[string]any
		want    string
		wantErr bool
	}{
		{
			name:   "replace digits",
			config: map[string]any{"type": "regex_replace", "input": "test 123 and 456", "pattern": `\d+`, "replacement": "X"},
			want:   "test X and X",
		},
		{
			name:   "replace with capture group",
			config: map[string]any{"type": "regex_replace", "input": "hello world", "pattern": `(\w+) (\w+)`, "replacement": "$2 $1"},
			want:   "world hello",
		},
		{
			name:   "no match",
			config: map[string]any{"type": "regex_replace", "input": "hello world", "pattern": `\d+`, "replacement": "X"},
			want:   "hello world",
		},
		{
			name:    "invalid regex",
			config:  map[string]any{"type": "regex_replace", "input": "test", "pattern": `[`, "replacement": "X"},
			wantErr: true,
		},
		{
			name:   "replace with variable",
			config: map[string]any{"type": "regex_replace", "input": "{{text}}", "pattern": `\d+`, "replacement": "NUM"},
			vars:   map[string]any{"text": "id: 123"},
			want:   "id: NUM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &StringNode{Action: node.ActionFromMap(tt.config)}
			ctx := &node.NodeContext{
				Variables: tt.vars,
				Stop:      make(chan struct{}),
			}

			result, err := n.Execute(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, ok := result.Response.(string)
				if !ok {
					t.Errorf("Response is not string, got %T", result.Response)
					return
				}

				if got != tt.want {
					t.Errorf("got = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func TestStringLength(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]any
		vars   map[string]any
		want   int
	}{
		{
			name:   "basic length",
			config: map[string]any{"type": "length", "input": "hello"},
			want:   5,
		},
		{
			name:   "empty string",
			config: map[string]any{"type": "length", "input": ""},
			want:   0,
		},
		{
			name:   "with spaces",
			config: map[string]any{"type": "length", "input": "hello world"},
			want:   11,
		},
		{
			name:   "with variable",
			config: map[string]any{"type": "length", "input": "{{text}}"},
			vars:   map[string]any{"text": "testing"},
			want:   7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &StringNode{Action: node.ActionFromMap(tt.config)}
			ctx := &node.NodeContext{
				Variables: tt.vars,
				Stop:      make(chan struct{}),
			}

			result, err := n.Execute(ctx)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}

			got, ok := result.Response.(int)
			if !ok {
				t.Errorf("Response is not int, got %T", result.Response)
				return
			}

			if got != tt.want {
				t.Errorf("got = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestStringStopSignal(t *testing.T) {
	n := &StringNode{Action: node.ActionFromMap(map[string]any{"type": "trim", "input": "test"})}

	stop := make(chan struct{})
	close(stop)

	ctx := &node.NodeContext{
		Variables: map[string]any{},
		Stop:      stop,
	}

	_, err := n.Execute(ctx)
	if err != node.ErrNodeStopped {
		t.Errorf("Expected ErrNodeStopped, got %v", err)
	}
}

func TestStringUnknownAction(t *testing.T) {
	n := &StringNode{Action: node.ActionFromMap(map[string]any{"type": "unknown_action", "input": "test"})}

	ctx := &node.NodeContext{
		Variables: map[string]any{},
		Stop:      make(chan struct{}),
	}

	result, err := n.Execute(ctx)
	if err == nil {
		t.Error("Expected error for unknown action")
	}
	if result.Handle != "error" {
		t.Errorf("Expected error handle, got %s", result.Handle)
	}
	if !strings.Contains(err.Error(), "unknown action") {
		t.Errorf("Expected 'unknown action' in error, got %v", err)
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
	default:
		return av == b
	}
}
