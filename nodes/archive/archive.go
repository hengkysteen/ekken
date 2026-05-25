package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"ekken/internal/features/workflow/node"
)

type ArchiveNode struct {
	Action node.Action
}

func init() {
	node.GlobalRegistry.Register(node.NodeRegistration{
		Spec: node.Spec{
			Meta: node.Meta{
				Type:        "archive",
				Label:       "Archive",
				Icon:        "https://www.svgrepo.com/show/424858/zip-file-type.svg",
				Tags:        []string{"System"},
				Description: "Compress or decompress ZIP files.",
			},

			DefaultAction: "compress",
			Actions: []node.Action{
				{
					Type:         "compress",
					Label:        "Compress (Zip)",
					Description:  "Compress file or directory into a ZIP file",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
					Fields: []node.NodeField{
						{
							Key:      "source_path",
							Type:     "string",
							Required: true,
							Label:    "Source file/folder",
						},
						{
							Key:      "archive_path",
							Type:     "string",
							Required: true,
							Label:    "Target ZIP file",
						},
					},
					AutoLayout: [][]node.AutoLayout{
						{{Key: "source_path", Component: "input", Flex: 24, Options: map[string]any{"native_file_picker": true, "native_file_picker_directory": true}}},
						{{Key: "archive_path", Component: "input", Flex: 24, Options: map[string]any{"native_file_picker": true}}},
					},
				},
				{
					Type:         "decompress",
					Label:        "Decompress (Unzip)",
					Description:  "Extract a ZIP file to target directory",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
					Fields: []node.NodeField{
						{
							Key:      "archive_path",
							Type:     "string",
							Required: true,
							Label:    "ZIP file",
						},
						{
							Key:      "destination_path",
							Type:     "string",
							Required: true,
							Label:    "Destination folder",
						},
					},
					AutoLayout: [][]node.AutoLayout{
						{{Key: "archive_path", Component: "input", Flex: 24, Options: map[string]any{"native_file_picker": true}}},
						{{Key: "destination_path", Component: "input", Flex: 24, Options: map[string]any{"native_file_picker_directory": true}}},
					},
				},
			},
			OutputHandles: []string{"success", "error"},
		},

		ExecutorFactory: func(action node.Action) node.NodeExecutor {
			return &ArchiveNode{Action: action}
		},
	})
}

func (n *ArchiveNode) Execute(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	// Proactive check to see if workflow is already stopped
	select {
	case <-ctx.Stop:
		return node.NodeExecutionResult{}, node.ErrNodeStopped
	default:
	}

	switch n.Action.Type {
	case "decompress":
		return n.executeDecompress(ctx)
	default:
		return n.executeCompress(ctx)
	}
}

func (n *ArchiveNode) executeCompress(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	sourceRaw, _ := node.FieldValue(n.Action, "source_path").(string)
	archiveRaw, _ := node.FieldValue(n.Action, "archive_path").(string)

	if sourceRaw == "" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("source path is required")
	}
	if archiveRaw == "" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("archive path is required")
	}

	source := node.ParseTemplate(sourceRaw, ctx.Variables)
	archive := node.ParseTemplate(archiveRaw, ctx.Variables)

	// Ensure destination directory for the archive exists
	archiveDir := filepath.Dir(archive)
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to create directory for archive: %w", err)
	}

	// Create zip file
	zipFile, err := os.Create(archive)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to create archive file: %w", err)
	}
	defer zipFile.Close()

	archiveWriter := zip.NewWriter(zipFile)
	defer archiveWriter.Close()

	// Check if source exists
	info, err := os.Stat(source)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("source path does not exist: %w", err)
	}

	sourceAbs, err := filepath.Abs(source)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to get absolute path of source: %w", err)
	}

	baseDir := filepath.Dir(sourceAbs)

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		select {
		case <-ctx.Stop:
			return node.ErrNodeStopped
		default:
		}

		pathAbs, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(baseDir, pathAbs)
		if err != nil {
			return err
		}

		// ZIP paths must use forward slashes for cross-platform compatibility
		relPath = filepath.ToSlash(relPath)

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = relPath

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archiveWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	}

	if info.IsDir() {
		err = filepath.Walk(sourceAbs, walkFunc)
	} else {
		err = walkFunc(sourceAbs, info, nil)
	}

	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, err
	}

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: fmt.Sprintf("Successfully compressed to %s", archive),
		Type:     &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
	}, nil
}

func (n *ArchiveNode) executeDecompress(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	archiveRaw, _ := node.FieldValue(n.Action, "archive_path").(string)
	destinationRaw, _ := node.FieldValue(n.Action, "destination_path").(string)

	if archiveRaw == "" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("archive path is required")
	}
	if destinationRaw == "" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("destination path is required")
	}

	archive := node.ParseTemplate(archiveRaw, ctx.Variables)
	destination := node.ParseTemplate(destinationRaw, ctx.Variables)

	destAbs, err := filepath.Abs(destination)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("invalid destination path: %w", err)
	}

	r, err := zip.OpenReader(archive)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to open archive: %w", err)
	}
	defer r.Close()

	if err := os.MkdirAll(destAbs, 0755); err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to create destination directory: %w", err)
	}

	for _, f := range r.File {
		select {
		case <-ctx.Stop:
			return node.NodeExecutionResult{}, node.ErrNodeStopped
		default:
		}

		cleanedName := filepath.Clean(f.Name)
		osName := filepath.FromSlash(cleanedName)
		targetPath := filepath.Join(destAbs, osName)

		// Clean targetPath to resolve any relative pathing
		targetPath = filepath.Clean(targetPath)
		// Zip Slip vulnerability check: ensure destAbs is a prefix of targetPath
		if !strings.HasPrefix(targetPath, destAbs+string(filepath.Separator)) && targetPath != destAbs {
			return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("illegal file path (Zip Slip attempt): %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, f.Mode()); err != nil {
				return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to create parent directory: %w", err)
		}

		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to create file %s: %w", targetPath, err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to open zip file content for %s: %w", f.Name, err)
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to write file content for %s: %w", f.Name, err)
		}
	}

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: fmt.Sprintf("Successfully decompressed to %s", destination),
		Type:     &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
	}, nil
}
