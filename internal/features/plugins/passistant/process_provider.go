package passistant

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"ekken/internal/logger"
)

type ProcessRunner struct {
	ID           string
	Name         string
	Icon         string
	OfficialURL  string
	ConfigFields []string
	Config       map[string]string
	Command      string
	SourcePath   string
	ExecTimeout  time.Duration
}

func NewProcessRunner(runner RunnerSpec, provider ProviderSpec, sourcePath string, execTimeout time.Duration) *ProcessRunner {
	return &ProcessRunner{
		ID:           provider.ID,
		Name:         provider.Name,
		Icon:         provider.Icon,
		OfficialURL:  provider.OfficialURL,
		ConfigFields: provider.ConfigFields,
		Config:       make(map[string]string),
		Command:      runner.Command,
		SourcePath:   sourcePath,
		ExecTimeout:  execTimeout,
	}
}

func (p *ProcessRunner) Configure(config map[string]string) {
	p.Config = config
}

func (p *ProcessRunner) Chat(ctx context.Context, req ChatRequest, onChunk func(content, thinking string)) (MessageContent, error) {
	// 1. Resolve Command path
	cmdPath := p.Command
	if strings.HasPrefix(cmdPath, "./") || strings.HasPrefix(cmdPath, "../") {
		cmdPath = filepath.Join(p.SourcePath, cmdPath)
	}
	cmdPath = filepath.Clean(cmdPath)

	// Apply ExecTimeout to Context if specified
	var cancel context.CancelFunc
	if p.ExecTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, p.ExecTimeout)
		defer cancel()
	}

	// 2. Prepare Command
	cmd := exec.CommandContext(ctx, cmdPath)
	prepareCmd(cmd)

	// Setup custom process group cancellation logic on context timeout/cancel
	cmd.Cancel = func() error {
		killProcessGroup(cmd)
		return nil
	}

	// Make sure we kill any remaining processes in the group on exit
	defer killProcessGroup(cmd)

	// Get Pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return MessageContent{}, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return MessageContent{}, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderrReader, err := cmd.StderrPipe()
	if err != nil {
		return MessageContent{}, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start continuous Stderr reader goroutine
	var stderrBuf bytes.Buffer
	go func() {
		scanner := bufio.NewScanner(stderrReader)
		for scanner.Scan() {
			line := scanner.Text()
			logger.Debug(fmt.Sprintf("[%s-stderr] %s", p.ID, line))
			stderrBuf.WriteString(line + "\n")
		}
	}()

	// Start Subprocess
	if err := cmd.Start(); err != nil {
		return MessageContent{}, fmt.Errorf("failed to start process: %w", err)
	}

	// Write request to Stdin in a separate goroutine to prevent deadlocks
	go func() {
		defer stdin.Close()

		cfg := make(map[string]any)
		for k, v := range p.Config {
			cfg[k] = v
		}

		stdinReq := StdinRequest{
			Kind:       "assistant",
			ProviderID: p.ID,
			Request:    req,
			Config:     cfg,
		}

		data, err := json.Marshal(stdinReq)
		if err != nil {
			logger.Error("Failed to marshal stdin request JSON", "provider", p.ID, "error", err)
			return
		}

		if _, err := stdin.Write(data); err != nil {
			logger.Error("Failed to write request to stdin", "provider", p.ID, "error", err)
			return
		}
		if _, err := stdin.Write([]byte("\n")); err != nil {
			logger.Error("Failed to write newline to stdin", "provider", p.ID, "error", err)
			return
		}
	}()

	// Read and Parse stdout line-by-line
	scanner := bufio.NewScanner(stdout)
	var finalMessage MessageContent
	finalMessage.Role = "assistant"

	var contentBuilder strings.Builder
	var thinkingBuilder strings.Builder
loop:
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var chunk StreamResponseChunk
		if err := json.Unmarshal(line, &chunk); err != nil {
			return MessageContent{}, fmt.Errorf("invalid json response from plugin: %w", err)
		}

		switch chunk.Type {
		case "error":
			return MessageContent{}, fmt.Errorf("plugin returned error: %s", chunk.Error)
		case "message":
			if chunk.Message != nil {
				finalMessage.Role = chunk.Message.Role
				finalMessage.Content = chunk.Message.Content
				finalMessage.Thinking = chunk.Message.Thinking
			}
		case "chunk":
			contentBuilder.WriteString(chunk.Content)
			thinkingBuilder.WriteString(chunk.Thinking)
			if onChunk != nil {
				onChunk(chunk.Content, chunk.Thinking)
			}
		case "done":
			break loop
		}
	}

	if err := scanner.Err(); err != nil {
		return MessageContent{}, fmt.Errorf("error reading stdout: %w", err)
	}

	// Wait for process to finish
	waitErr := cmd.Wait()

	if ctx.Err() != nil {
		return MessageContent{}, ctx.Err()
	}

	if waitErr != nil {
		stderrStr := strings.TrimSpace(stderrBuf.String())
		if stderrStr != "" {
			return MessageContent{}, fmt.Errorf("plugin process exited with error: %w (stderr: %s)", waitErr, stderrStr)
		}
		return MessageContent{}, fmt.Errorf("plugin process exited with error: %w", waitErr)
	}

	if contentBuilder.Len() > 0 || thinkingBuilder.Len() > 0 {
		finalMessage.Content += contentBuilder.String()
		finalMessage.Thinking += thinkingBuilder.String()
	}

	return finalMessage, nil
}
