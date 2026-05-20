package passistant

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const dummyPluginSrc = `
package main

import (
	"fmt"
	"time"
)

func main() {
	// Print message chunk
	fmt.Println("{\"type\":\"message\",\"message\":{\"role\":\"assistant\",\"content\":\"hello\"}}")
	time.Sleep(10 * time.Millisecond)
	// Print text chunk
	fmt.Println("{\"type\":\"chunk\",\"content\":\" world\",\"thinking\":\"thinking...\"}")
	time.Sleep(10 * time.Millisecond)
	// Print done chunk
	fmt.Println("{\"type\":\"done\"}")
}
`

func TestProcessRunner_Chat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ekken-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write dummy plugin source
	srcPath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(srcPath, []byte(dummyPluginSrc), 0644); err != nil {
		t.Fatalf("failed to write dummy source: %v", err)
	}

	// Compile dummy plugin
	binPath := filepath.Join(tmpDir, "dummy")
	if os.PathSeparator == '\\' {
		binPath += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", binPath, srcPath)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to compile dummy plugin: %v", err)
	}

	runnerSpec := RunnerSpec{
		Command: binPath,
	}

	providerSpec := ProviderSpec{
		ID:   "test-dummy",
		Name: "Test Dummy",
	}

	runner := NewProcessRunner(runnerSpec, providerSpec, tmpDir, 5*time.Second)

	req := ChatRequest{
		ConversationID: "conv-1",
		Model:          "test-model",
	}

	var chunks []string
	var thinkChunks []string
	onChunk := func(content, thinking string) {
		if content != "" {
			chunks = append(chunks, content)
		}
		if thinking != "" {
			thinkChunks = append(thinkChunks, thinking)
		}
	}

	ctx := context.Background()
	msg, err := runner.Chat(ctx, req, onChunk)
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}

	if msg.Role != "assistant" {
		t.Errorf("expected role 'assistant', got %q", msg.Role)
	}

	// Check accumulated output
	// Message + Chunk content = "hello" + " world" = "hello world"
	expectedContent := "hello world"
	if msg.Content != expectedContent {
		t.Errorf("expected content %q, got %q", expectedContent, msg.Content)
	}

	if msg.Thinking != "thinking..." {
		t.Errorf("expected thinking %q, got %q", "thinking...", msg.Thinking)
	}

	// Verify streaming callback
	if len(chunks) != 1 || chunks[0] != " world" {
		t.Errorf("unexpected streaming content chunks: %v", chunks)
	}

	if len(thinkChunks) != 1 || thinkChunks[0] != "thinking..." {
		t.Errorf("unexpected streaming thinking chunks: %v", thinkChunks)
	}
}

func TestProcessRunner_Timeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ekken-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// A plugin that hangs forever
	hangSrc := `
package main

import "time"

func main() {
	time.Sleep(10 * time.Second)
}
`
	srcPath := filepath.Join(tmpDir, "main.go")
	_ = os.WriteFile(srcPath, []byte(hangSrc), 0644)

	binPath := filepath.Join(tmpDir, "dummy_hang")
	if os.PathSeparator == '\\' {
		binPath += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", binPath, srcPath)
	_ = cmd.Run()

	runnerSpec := RunnerSpec{
		Command: binPath,
	}

	providerSpec := ProviderSpec{
		ID:   "test-dummy-hang",
		Name: "Test Dummy Hang",
	}

	// Set execution timeout to 500ms
	runner := NewProcessRunner(runnerSpec, providerSpec, tmpDir, 500*time.Millisecond)

	req := ChatRequest{
		ConversationID: "conv-2",
		Model:          "test-model",
	}

	ctx := context.Background()
	_, err = runner.Chat(ctx, req, nil)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}
