package shell

import (
	"context"
	"ekken/internal/features/workflow/node"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type ShellNode struct {
	Action node.Action
}

var dangerousCommands = []string{
	"rm -rf /",
	"rm -rf /*",
	"rmdir /s",
	"format",
	"mkfs",
	"dd if=",
	"> /dev/sda",
	"fdisk",
	":(){:|:&};:",
	"fork bomb",
}

func init() {
	node.GlobalRegistry.Register(node.NodeRegistration{
		Spec: node.Spec{
			Meta: node.Meta{
				Type:        "shell",
				Label:       "Shell",
				Icon:        "https://www.svgrepo.com/show/441290/terminal.svg",
				Tags:        []string{"System", "Advanced"},
				Description: "Execute shell commands. Requires developer knowledge.",
			},
			DefaultAction: "execute",
			Actions: []node.Action{
				{
					Type:         "execute",
					Label:        "Execute",
					Description:  "Execute shell command",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
					Fields: []node.NodeField{
						{Key: "command", Type: "string", Required: true, Label: "Command"},
						{Key: "working_dir", Type: "string", Label: "Working Directory"},
					},
					AutoLayout: [][]node.AutoLayout{
						{{Key: "command", Component: "textarea", Flex: 24, Options: map[string]any{"helper": "Shell command to execute", "rows": 4}}},
						{{Key: "working_dir", Component: "input", Flex: 24, Options: map[string]any{"placeholder": "Optional working directory", "native_file_picker_directory": true}}},
					},
				},
			},
			GlobalFields: []node.NodeField{
				{Key: "timeout", Type: "number", Default: 30, Label: "Timeout (sec)"},
			},
			Outputs: []node.HandleEdge{
				{Key: "success", Label: "Success", Tone: "success"},
				{Key: "error", Label: "Error", Tone: "error"},
			},
		},
		ExecutorFactory: func(action node.Action) node.NodeExecutor {
			return &ShellNode{Action: action}
		},
	})
}

func (n *ShellNode) Execute(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	select {
	case <-ctx.Stop:
		return node.NodeExecutionResult{}, node.ErrNodeStopped
	default:
	}

	commandRaw, _ := node.FieldValue(n.Action, "command").(string)
	workingDirRaw, _ := node.FieldValue(n.Action, "working_dir").(string)

	if strings.TrimSpace(commandRaw) == "" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("command is required")
	}

	command := node.ParseTemplate(commandRaw, ctx.Variables)
	workingDir := node.ParseTemplate(workingDirRaw, ctx.Variables)

	// Validate against blocklist
	if err := validateCommand(command); err != nil {
		return node.NodeExecutionResult{Handle: "error"}, err
	}

	// Get timeout
	timeout := 30
	if t, ok := node.FieldValue(n.Action, "timeout").(float64); ok {
		timeout = int(t)
	}

	// Execute command
	output, err := executeCommand(ctx, command, workingDir, timeout)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, err
	}

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: output,
		Type:     &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
	}, nil
}

func validateCommand(command string) error {
	commandLower := strings.ToLower(command)

	for _, dangerous := range dangerousCommands {
		if strings.Contains(commandLower, strings.ToLower(dangerous)) {
			return fmt.Errorf("blocked dangerous command: %s", dangerous)
		}
	}

	return nil
}

func executeCommand(ctx *node.NodeContext, command, workingDir string, timeoutSec int) (string, error) {
	shell, shellArgs := getShellCommand()

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	// Build command
	args := append(shellArgs, command)
	cmd := exec.CommandContext(execCtx, shell, args...)

	// Set working directory if provided
	if workingDir != "" {
		if _, err := os.Stat(workingDir); os.IsNotExist(err) {
			return "", fmt.Errorf("working directory does not exist: %s", workingDir)
		}
		cmd.Dir = workingDir
	}

	// Handle stop signal
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Stop:
			cancel()
		case <-done:
		}
	}()

	// Execute command
	output, err := cmd.CombinedOutput()
	close(done)

	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command timeout after %d seconds", timeoutSec)
		}
		return string(output), fmt.Errorf("command failed: %v\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

func getShellCommand() (string, []string) {
	if runtime.GOOS == "windows" {
		// Check if PowerShell is available
		if _, err := exec.LookPath("powershell.exe"); err == nil {
			return "powershell.exe", []string{"-Command"}
		}
		// Fallback to cmd
		return "cmd.exe", []string{"/C"}
	}

	// macOS and Linux
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	return shell, []string{"-c"}
}
