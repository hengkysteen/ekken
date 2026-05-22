package shell

import (
	"ekken/internal/features/workflow/node"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestShellExecute(t *testing.T) {
	tests := []struct {
		name       string
		config     map[string]any
		vars       map[string]any
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "echo command",
			config:     map[string]any{"type": "execute", "command": "echo hello"},
			wantOutput: "hello",
		},
		{
			name:    "empty command",
			config:  map[string]any{"type": "execute", "command": ""},
			wantErr: true,
		},
		{
			name:       "command with variable",
			config:     map[string]any{"type": "execute", "command": "echo {{text}}"},
			vars:       map[string]any{"text": "world"},
			wantOutput: "world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &ShellNode{Action: node.ActionFromMap(tt.config)}
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
				if result.Handle != "success" {
					t.Errorf("Handle = %s, want success", result.Handle)
				}

				output, ok := result.Response.(string)
				if !ok {
					t.Errorf("Response is not string, got %T", result.Response)
					return
				}

				if !strings.Contains(strings.TrimSpace(output), tt.wantOutput) {
					t.Errorf("Output = %q, want to contain %q", output, tt.wantOutput)
				}
			}
		})
	}
}

func TestShellBlocklist(t *testing.T) {
	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{
			name:    "safe command",
			command: "echo hello",
			wantErr: false,
		},
		{
			name:    "dangerous rm -rf /",
			command: "rm -rf /",
			wantErr: true,
		},
		{
			name:    "dangerous format",
			command: "format c:",
			wantErr: true,
		},
		{
			name:    "dangerous dd",
			command: "dd if=/dev/zero of=/dev/sda",
			wantErr: true,
		},
		{
			name:    "fork bomb",
			command: ":(){:|:&};:",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommand(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestShellWorkingDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	tests := []struct {
		name       string
		config     map[string]any
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "pwd in /tmp",
			config:     map[string]any{"type": "execute", "command": "pwd", "working_dir": "/tmp"},
			wantOutput: "/tmp",
		},
		{
			name:    "invalid working dir",
			config:  map[string]any{"type": "execute", "command": "pwd", "working_dir": "/nonexistent"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &ShellNode{Action: node.ActionFromMap(tt.config)}
			ctx := &node.NodeContext{
				Variables: map[string]any{},
				Stop:      make(chan struct{}),
			}

			result, err := n.Execute(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				output, ok := result.Response.(string)
				if !ok {
					t.Errorf("Response is not string, got %T", result.Response)
					return
				}

				if !strings.Contains(strings.TrimSpace(output), tt.wantOutput) {
					t.Errorf("Output = %q, want to contain %q", output, tt.wantOutput)
				}
			}
		})
	}
}

func TestShellTimeout(t *testing.T) {
	var sleepCmd string
	if runtime.GOOS == "windows" {
		sleepCmd = "timeout /t 5 /nobreak"
	} else {
		sleepCmd = "sleep 5"
	}

	n := &ShellNode{
		Action: node.ActionFromMap(map[string]any{
			"type":    "execute",
			"command": sleepCmd,
			"timeout": 1.0,
		}),
	}

	ctx := &node.NodeContext{
		Variables: map[string]any{},
		Stop:      make(chan struct{}),
	}

	start := time.Now()
	_, err := n.Execute(ctx)
	duration := time.Since(start)

	if err == nil {
		t.Error("Expected timeout error")
	}

	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("Expected timeout error, got: %v", err)
	}

	if duration > 3*time.Second {
		t.Errorf("Timeout took too long: %v", duration)
	}
}

func TestShellStopSignal(t *testing.T) {
	n := &ShellNode{
		Action: node.ActionFromMap(map[string]any{
			"type":    "execute",
			"command": "echo test",
		}),
	}

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

func TestShellCrossPlatform(t *testing.T) {
	shell, args := getShellCommand()

	if runtime.GOOS == "windows" {
		if shell != "powershell.exe" && shell != "cmd.exe" {
			t.Errorf("Expected powershell.exe or cmd.exe on Windows, got %s", shell)
		}
		if shell == "powershell.exe" && args[0] != "-Command" {
			t.Errorf("Expected -Command for PowerShell, got %v", args)
		}
		if shell == "cmd.exe" && args[0] != "/C" {
			t.Errorf("Expected /C for cmd, got %v", args)
		}
	} else {
		if !strings.HasSuffix(shell, "sh") && !strings.HasSuffix(shell, "bash") && !strings.HasSuffix(shell, "zsh") {
			t.Errorf("Expected shell ending with sh/bash/zsh, got %s", shell)
		}
		if args[0] != "-c" {
			t.Errorf("Expected -c for Unix shells, got %v", args)
		}
	}
}

func TestShellMultiline(t *testing.T) {
	var command string
	if runtime.GOOS == "windows" {
		command = "echo line1\necho line2"
	} else {
		command = "echo line1\necho line2"
	}

	n := &ShellNode{
		Action: node.ActionFromMap(map[string]any{
			"type":    "execute",
			"command": command,
		}),
	}

	ctx := &node.NodeContext{
		Variables: map[string]any{},
		Stop:      make(chan struct{}),
	}

	result, err := n.Execute(ctx)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
		return
	}

	output, ok := result.Response.(string)
	if !ok {
		t.Errorf("Response is not string, got %T", result.Response)
		return
	}

	if !strings.Contains(output, "line1") || !strings.Contains(output, "line2") {
		t.Errorf("Output = %q, want to contain line1 and line2", output)
	}
}
