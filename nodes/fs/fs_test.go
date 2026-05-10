package fs

import (
	"ekken/internal/features/workflow/node"
	"os"
	"path/filepath"
	"testing"
)

func TestFSNode_ExecuteWrite(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]any
		variables   map[string]interface{}
		wantHandle  string
		wantErr     bool
		errContains string
		verify      func(t *testing.T, path string)
	}{
		{
			name: "write simple content",
			config: map[string]any{
				"action":  "write",
				"path":    "testdata/write_simple.txt",
				"content": "hello world",
			},
			variables:  map[string]interface{}{},
			wantHandle: "success",
			verify: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if string(content) != "hello world" {
					t.Errorf("content = %q, want %q", string(content), "hello world")
				}
			},
		},
		{
			name: "write with template variable",
			config: map[string]any{
				"action":  "write",
				"path":    "testdata/write_template.txt",
				"content": "{{message}}",
			},
			variables:  map[string]interface{}{"message": "templated content"},
			wantHandle: "success",
			verify: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if string(content) != "templated content" {
					t.Errorf("content = %q, want %q", string(content), "templated content")
				}
			},
		},
		{
			name: "write with explicit variable",
			config: map[string]any{
				"action":  "write",
				"path":    "testdata/write_output.txt",
				"content": "{{my_data}}",
			},
			variables:  map[string]interface{}{"my_data": "output data"},
			wantHandle: "success",
			verify: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if string(content) != "output data" {
					t.Errorf("content = %q, want %q", string(content), "output data")
				}
			},
		},
		{
			name: "write creates nested directories",
			config: map[string]any{
				"action":  "write",
				"path":    "testdata/nested/deep/file.txt",
				"content": "nested content",
			},
			variables:  map[string]interface{}{},
			wantHandle: "success",
			verify: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if string(content) != "nested content" {
					t.Errorf("content = %q, want %q", string(content), "nested content")
				}
			},
		},
		{
			name: "write overwrites existing file",
			config: map[string]any{
				"action":  "write",
				"path":    "testdata/overwrite.txt",
				"content": "new content",
			},
			variables:  map[string]interface{}{},
			wantHandle: "success",
			verify: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if string(content) != "new content" {
					t.Errorf("content = %q, want %q", string(content), "new content")
				}
			},
		},
		{
			name: "write empty path",
			config: map[string]any{
				"action":  "write",
				"path":    "",
				"content": "content",
			},
			variables:   map[string]interface{}{},
			wantHandle:  "error",
			wantErr:     true,
			errContains: "path is required",
		},
		{
			name: "write with path template",
			config: map[string]any{
				"action":  "write",
				"path":    "testdata/{{filename}}.txt",
				"content": "dynamic path",
			},
			variables:  map[string]interface{}{"filename": "dynamic"},
			wantHandle: "success",
			verify: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if string(content) != "dynamic path" {
					t.Errorf("content = %q, want %q", string(content), "dynamic path")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &FSNode{Config: tt.config}
			ctx := &node.NodeContext{
				Stop:      make(chan struct{}),
				Variables: tt.variables,
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
				if result.Handle != tt.wantHandle {
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

			if tt.verify != nil {
				path := node.ParseTemplate(tt.config["path"].(string), tt.variables)
				tt.verify(t, path)
			}
		})
	}

	// Cleanup
	os.RemoveAll("testdata")
}

func TestFSNode_ExecuteAppend(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]any
		variables   map[string]interface{}
		setup       func(t *testing.T, path string)
		wantHandle  string
		wantErr     bool
		errContains string
		verify      func(t *testing.T, path string)
	}{
		{
			name: "append to existing file",
			config: map[string]any{
				"action":  "append",
				"path":    "testdata/append_existing.txt",
				"content": " appended",
			},
			variables: map[string]interface{}{},
			setup: func(t *testing.T, path string) {
				os.MkdirAll(filepath.Dir(path), 0755)
				os.WriteFile(path, []byte("initial"), 0644)
			},
			wantHandle: "success",
			verify: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if string(content) != "initial appended" {
					t.Errorf("content = %q, want %q", string(content), "initial appended")
				}
			},
		},
		{
			name: "append to non-existing file",
			config: map[string]any{
				"action":  "append",
				"path":    "testdata/append_new.txt",
				"content": "first line",
			},
			variables:  map[string]interface{}{},
			wantHandle: "success",
			verify: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if string(content) != "first line" {
					t.Errorf("content = %q, want %q", string(content), "first line")
				}
			},
		},
		{
			name: "append multiple times",
			config: map[string]any{
				"action":  "append",
				"path":    "testdata/append_multiple.txt",
				"content": "line\n",
			},
			variables: map[string]interface{}{},
			setup: func(t *testing.T, path string) {
				os.MkdirAll(filepath.Dir(path), 0755)
				os.WriteFile(path, []byte("line\n"), 0644)
			},
			wantHandle: "success",
			verify: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if string(content) != "line\nline\n" {
					t.Errorf("content = %q, want %q", string(content), "line\nline\n")
				}
			},
		},
		{
			name: "append with template",
			config: map[string]any{
				"action":  "append",
				"path":    "testdata/append_template.txt",
				"content": "{{data}}",
			},
			variables: map[string]interface{}{"data": "templated append"},
			setup: func(t *testing.T, path string) {
				os.MkdirAll(filepath.Dir(path), 0755)
				os.WriteFile(path, []byte("start "), 0644)
			},
			wantHandle: "success",
			verify: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if string(content) != "start templated append" {
					t.Errorf("content = %q, want %q", string(content), "start templated append")
				}
			},
		},
		{
			name: "append creates nested directories",
			config: map[string]any{
				"action":  "append",
				"path":    "testdata/append/nested/file.txt",
				"content": "content",
			},
			variables:  map[string]interface{}{},
			wantHandle: "success",
			verify: func(t *testing.T, path string) {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if string(content) != "content" {
					t.Errorf("content = %q, want %q", string(content), "content")
				}
			},
		},
		{
			name: "append empty path",
			config: map[string]any{
				"action":  "append",
				"path":    "",
				"content": "content",
			},
			variables:   map[string]interface{}{},
			wantHandle:  "error",
			wantErr:     true,
			errContains: "path is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := node.ParseTemplate(tt.config["path"].(string), tt.variables)
			
			if tt.setup != nil {
				tt.setup(t, path)
			}

			n := &FSNode{Config: tt.config}
			ctx := &node.NodeContext{
				Stop:      make(chan struct{}),
				Variables: tt.variables,
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
				if result.Handle != tt.wantHandle {
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

			if tt.verify != nil {
				tt.verify(t, path)
			}
		})
	}

	// Cleanup
	os.RemoveAll("testdata")
}

func TestFSNode_Stop(t *testing.T) {
	n := &FSNode{
		Config: map[string]any{
			"action":  "write",
			"path":    "testdata/stop.txt",
			"content": "content",
		},
	}

	stop := make(chan struct{})
	close(stop)

	ctx := &node.NodeContext{
		Stop:      stop,
		Variables: map[string]interface{}{},
	}

	result, err := n.Execute(ctx)

	if err != node.ErrNodeStopped {
		t.Errorf("Execute() error = %v, want %v", err, node.ErrNodeStopped)
	}

	if result.Handle != "" {
		t.Errorf("Execute() handle = %v, want empty", result.Handle)
	}
}

func TestFSNode_ExecuteDelete(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]any
		variables   map[string]interface{}
		setup       func(t *testing.T, path string)
		wantHandle  string
		wantErr     bool
		errContains string
		verify      func(t *testing.T, path string)
	}{
		{
			name: "delete existing file",
			config: map[string]any{
				"action": "delete",
				"path":   "testdata/delete_file.txt",
			},
			variables: map[string]interface{}{},
			setup: func(t *testing.T, path string) {
				os.MkdirAll(filepath.Dir(path), 0755)
				os.WriteFile(path, []byte("content"), 0644)
			},
			wantHandle: "success",
			verify: func(t *testing.T, path string) {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Errorf("file still exists at %s", path)
				}
			},
		},
		{
			name: "delete existing directory",
			config: map[string]any{
				"action": "delete",
				"path":   "testdata/delete_dir",
			},
			variables: map[string]interface{}{},
			setup: func(t *testing.T, path string) {
				os.MkdirAll(path, 0755)
				os.WriteFile(filepath.Join(path, "file.txt"), []byte("content"), 0644)
			},
			wantHandle: "success",
			verify: func(t *testing.T, path string) {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Errorf("directory still exists at %s", path)
				}
			},
		},
		{
			name: "delete non-existing path",
			config: map[string]any{
				"action": "delete",
				"path":   "testdata/non_existing",
			},
			variables:  map[string]interface{}{},
			wantHandle: "success", // RemoveAll returns nil if path doesn't exist
		},
		{
			name: "delete with template",
			config: map[string]any{
				"action": "delete",
				"path":   "testdata/{{target}}",
			},
			variables: map[string]interface{}{"target": "delete_me.txt"},
			setup: func(t *testing.T, path string) {
				os.MkdirAll(filepath.Dir(path), 0755)
				os.WriteFile(path, []byte("content"), 0644)
			},
			wantHandle: "success",
			verify: func(t *testing.T, path string) {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Errorf("file still exists at %s", path)
				}
			},
		},
		{
			name: "delete empty path",
			config: map[string]any{
				"action": "delete",
				"path":   "",
			},
			variables:   map[string]interface{}{},
			wantHandle:  "error",
			wantErr:     true,
			errContains: "path is required",
		},
		{
			name: "delete multiple paths",
			config: map[string]any{
				"action": "delete",
				"path":   "testdata/f1.txt\ntestdata/f2.txt",
			},
			variables: map[string]interface{}{},
			setup: func(t *testing.T, path string) {
				os.MkdirAll("testdata", 0755)
				os.WriteFile("testdata/f1.txt", []byte("1"), 0644)
				os.WriteFile("testdata/f2.txt", []byte("2"), 0644)
			},
			wantHandle: "success",
			verify: func(t *testing.T, path string) {
				if _, err := os.Stat("testdata/f1.txt"); !os.IsNotExist(err) {
					t.Errorf("f1.txt still exists")
				}
				if _, err := os.Stat("testdata/f2.txt"); !os.IsNotExist(err) {
					t.Errorf("f2.txt still exists")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.config["path"].(string)
			if path != "" {
				path = node.ParseTemplate(path, tt.variables)
			}

			if tt.setup != nil {
				tt.setup(t, path)
			}

			n := &FSNode{Config: tt.config}
			ctx := &node.NodeContext{
				Stop:      make(chan struct{}),
				Variables: tt.variables,
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
				if result.Handle != tt.wantHandle {
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

			if tt.verify != nil {
				tt.verify(t, path)
			}
		})
	}

	// Cleanup
	os.RemoveAll("testdata")
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
