package archive

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ekken/internal/features/workflow/node"
)

func TestArchiveNode_CompressAndDecompress(t *testing.T) {
	// Setup temp directory
	tmpDir, err := os.MkdirTemp("", "ekken-archive-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test folder structure
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(filepath.Join(srcDir, "sub"), 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}

	file1 := filepath.Join(srcDir, "file1.txt")
	file2 := filepath.Join(srcDir, "sub", "file2.txt")

	if err := os.WriteFile(file1, []byte("hello file1"), 0644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("hello file2"), 0644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}

	archivePath := filepath.Join(tmpDir, "test.zip")
	destDir := filepath.Join(tmpDir, "dest")

	// Execute Compress
	compressNode := &ArchiveNode{
		Action: node.Action{
			Type: "compress",
			Fields: []node.NodeField{
				{Key: "source_path", Value: srcDir},
				{Key: "archive_path", Value: archivePath},
			},
		},
	}

	ctx := &node.NodeContext{
		Context:   context.Background(),
		Variables: make(map[string]any),
		Stop:      make(chan struct{}),
	}

	res, err := compressNode.Execute(ctx)
	if err != nil {
		t.Fatalf("compress execute failed: %v", err)
	}
	if res.Handle != "success" {
		t.Fatalf("expected success, got %s", res.Handle)
	}

	// Verify zip file exists
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Fatalf("expected zip file to exist")
	}

	// Execute Decompress
	decompressNode := &ArchiveNode{
		Action: node.Action{
			Type: "decompress",
			Fields: []node.NodeField{
				{Key: "archive_path", Value: archivePath},
				{Key: "destination_path", Value: destDir},
			},
		},
	}

	res, err = decompressNode.Execute(ctx)
	if err != nil {
		t.Fatalf("decompress execute failed: %v", err)
	}
	if res.Handle != "success" {
		t.Fatalf("expected success, got %s", res.Handle)
	}

	// Verify extracted files
	b1, err := os.ReadFile(filepath.Join(destDir, "src", "file1.txt"))
	if err != nil {
		t.Fatalf("failed to read extracted file1: %v", err)
	}
	if string(b1) != "hello file1" {
		t.Fatalf("expected 'hello file1', got '%s'", string(b1))
	}

	b2, err := os.ReadFile(filepath.Join(destDir, "src", "sub", "file2.txt"))
	if err != nil {
		t.Fatalf("failed to read extracted file2: %v", err)
	}
	if string(b2) != "hello file2" {
		t.Fatalf("expected 'hello file2', got '%s'", string(b2))
	}
}

func TestArchiveNode_ZipSlipPrevention(t *testing.T) {
	// Setup temp directory
	tmpDir, err := os.MkdirTemp("", "ekken-zipslip-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, "malicious.zip")
	destDir := filepath.Join(tmpDir, "dest")

	// Create malicious zip file
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("failed to create zip: %v", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	
	// Create entry with relative path traversal pointing outside target
	h := &zip.FileHeader{
		Name:   "../outside.txt",
		Method: zip.Deflate,
	}
	w, err := zw.CreateHeader(h)
	if err != nil {
		t.Fatalf("failed to create zip header: %v", err)
	}
	if _, err := w.Write([]byte("malicious content")); err != nil {
		t.Fatalf("failed to write content: %v", err)
	}
	zw.Close()

	// Decompress
	decompressNode := &ArchiveNode{
		Action: node.Action{
			Type: "decompress",
			Fields: []node.NodeField{
				{Key: "archive_path", Value: archivePath},
				{Key: "destination_path", Value: destDir},
			},
		},
	}

	ctx := &node.NodeContext{
		Context:   context.Background(),
		Variables: make(map[string]any),
		Stop:      make(chan struct{}),
	}

	_, err = decompressNode.Execute(ctx)
	if err == nil {
		t.Fatalf("expected error from zip slip prevention, got nil")
	}
	if !strings.Contains(err.Error(), "Zip Slip") {
		t.Fatalf("expected zip slip error message, got: %v", err)
	}

	// Verify outside file was NOT created
	outsideFile := filepath.Join(tmpDir, "outside.txt")
	if _, err := os.Stat(outsideFile); err == nil {
		t.Fatalf("security violation: file was extracted outside target directory!")
	}
}
