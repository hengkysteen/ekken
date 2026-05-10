package fs

import (
	"ekken/internal/features/workflow/node"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FSNode struct {
	Config map[string]any
}

func init() {
	node.GlobalRegistry.Register(node.NodeRegistration{
		NodeSpec: node.NodeSpec{
			NodeMetadata: node.NodeMetadata{
				Type:  "fs",
				Label: "File System",
				Icon:  "https://www.svgrepo.com/show/506485/file-o.svg",
				Tags:  []string{"System"},
			},
			Description:   "File System operations.",
			DefaultAction: "write",
			Actions: []node.NodeAction{
				{
					Key:         "write",
					Label:       "Write",
					Description: "Write content to a file.",
					HasResponse: true,
					Fields: []node.NodeField{
						{
							Key:      "path",
							Type:     "string",
							Required: true,
							Label:    "Target file",
						},
						{
							Key:   "content",
							Type:  "string",
							Label: "Content to write",
						},
					},
					Form: [][]node.Form{
						{
							{
								Key:       "path",
								Component: "input",
								Flex:      24,
								FormOptions: map[string]any{
									"native_file_picker":           true,
									"native_file_picker_directory": true,
								},
							},
						},
						{{Key: "content", Component: "textarea", Flex: 24}},
					},
				},
				{
					Key:         "append",
					Label:       "Append",
					Description: "Append content to the end of a file.",
					HasResponse: true,
					Fields: []node.NodeField{
						{
							Key:      "path",
							Type:     "string",
							Required: true,
							Label:    "Target file",
						},
						{
							Key:   "content",
							Type:  "string",
							Label: "Content to append",
						},
					},
					Form: [][]node.Form{
						{
							{
								Key:       "path",
								Component: "input",
								Flex:      24,
								FormOptions: map[string]any{
									"native_file_picker":           true,
									"native_file_picker_directory": true,
								},
							},
						},
						{{Key: "content", Component: "textarea", Flex: 24}},
					},
				},
				{
					Key:         "delete",
					Label:       "Delete",
					Description: "Delete a file or directory.",
					HasResponse: true,
					Fields: []node.NodeField{
						{
							Key:      "path",
							Type:     "string",
							Required: true,
							Label:    "Target file/folder",
						},
					},
					Form: [][]node.Form{
						{
							{
								Key:       "path",
								Component: "input",
								Flex:      24,
								FormOptions: map[string]any{
									"native_file_picker":           true,
									"native_file_picker_multiple":  true,
									"native_file_picker_directory": true,
								},
							},
						},
					},
				},
			},
			Outputs: []node.NodeOutputDef{
				{Key: "success", Label: "Success", Tone: "success"},
				{Key: "error", Label: "Error", Tone: "error"},
			},
		},
		ExecutorFactory: func(config map[string]any, _ []node.Node) node.NodeExecutor {
			return &FSNode{Config: config}
		},
	})
}

func (n *FSNode) Execute(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	select {
	case <-ctx.Stop:
		return node.NodeExecutionResult{}, node.ErrNodeStopped
	default:
	}

	action, _ := n.Config["action"].(string)

	switch action {
	case "append":
		return n.executeAppend(ctx)
	case "delete":
		return n.executeDelete(ctx)
	default:
		return n.executeWrite(ctx)
	}
}

func (n *FSNode) executeWrite(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	pathRaw, _ := n.Config["path"].(string)
	contentRaw, _ := n.Config["content"].(string)

	if pathRaw == "" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("file path is required")
	}

	path := node.ParseTemplate(pathRaw, ctx.Variables)
	content := node.ParseTemplate(contentRaw, ctx.Variables)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to write file: %v", err)
	}

	detail, _ := getPathDetail(path)
	return node.NodeExecutionResult{
		Handle:   "success",
		Response: fmt.Sprintf("Successfully written: %s", detail),
		Type:     &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
	}, nil
}

func (n *FSNode) executeAppend(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	pathRaw, _ := n.Config["path"].(string)
	contentRaw, _ := n.Config["content"].(string)

	if pathRaw == "" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("file path is required")
	}

	path := node.ParseTemplate(pathRaw, ctx.Variables)
	content := node.ParseTemplate(contentRaw, ctx.Variables)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to create directory: %v", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to open file for append: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to append file: %v", err)
	}

	detail, _ := getPathDetail(path)
	return node.NodeExecutionResult{
		Handle:   "success",
		Response: fmt.Sprintf("Successfully appended to: %s", detail),
		Type:     &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
	}, nil
}

func (n *FSNode) executeDelete(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	pathRaw, _ := n.Config["path"].(string)

	if pathRaw == "" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("file path is required")
	}

	pathFull := node.ParseTemplate(pathRaw, ctx.Variables)
	// Handle multiple paths (separated by newline or comma from UI)
	var paths []string
	if strings.Contains(pathFull, "\n") {
		paths = strings.Split(pathFull, "\n")
	} else {
		paths = strings.Split(pathFull, ",")
	}

	var deletedCount int
	var details []string

	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}

		// Detect info before it's gone
		detail, exists := getPathDetail(path)

		if err := os.RemoveAll(path); err != nil {
			return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to delete %s: %v", path, err)
		}

		if exists {
			deletedCount++
			details = append(details, detail)
		}
	}

	var responseMsg string
	if deletedCount == 0 {
		if len(paths) == 1 {
			detail, _ := getPathDetail(strings.TrimSpace(paths[0]))
			responseMsg = fmt.Sprintf("%s not found", detail)
		} else {
			responseMsg = "No items were found to delete"
		}
	} else if deletedCount == 1 {
		responseMsg = fmt.Sprintf("Deleted %s", details[0])
	} else {
		responseMsg = fmt.Sprintf("Successfully deleted %d items", deletedCount)
	}

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: responseMsg,
		Type:     &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
	}, nil
}

func getPathDetail(path string) (string, bool) {
	info, err := os.Stat(path)
	name := filepath.Base(path)
	if err != nil {
		// If it doesn't exist or error, try to guess from extension
		if filepath.Ext(name) != "" {
			return fmt.Sprintf("File '%s'", name), false
		}
		return fmt.Sprintf("File/Folder '%s'", name), false
	}

	if info.IsDir() {
		return fmt.Sprintf("Folder '%s'", name), true
	}

	ext := filepath.Ext(name)
	if ext != "" {
		return fmt.Sprintf("File '%s' ", name), true
	}
	return fmt.Sprintf("File '%s'", name), true
}
