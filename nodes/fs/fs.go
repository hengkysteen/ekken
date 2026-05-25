package fs

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"ekken/internal/features/workflow/node"
)

type FSNode struct {
	Action node.Action
}

func init() {
	node.GlobalRegistry.Register(node.NodeRegistration{
		Spec: node.Spec{
			Meta: node.Meta{
				Type:        "fs",
				Label:       "File System",
				Tags:        []string{"System"},
				Icon:        "https://www.svgrepo.com/show/506485/file-o.svg",
				Description: "File System operations.",
			},

			DefaultAction: "write",
			Actions: []node.Action{
				{
					Type:         "write",
					Label:        "Write",
					Description:  "Write content to a file.",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
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
						{
							Key:     "encoding",
							Type:    "string",
							Default: "text",
							Label:   "Encoding",
							Options: []map[string]string{
								{"label": "Text", "value": "text"},
								{"label": "Base64", "value": "base64"},
								{"label": "Hex", "value": "hex"},
								{"label": "Data URL", "value": "data_url"},
							},
						},
					},
					AutoLayout: [][]node.AutoLayout{
						{
							{
								Key:       "path",
								Component: "input",
								Flex:      16,
								Options: map[string]any{
									"native_file_picker":           true,
									"native_file_picker_directory": true,
								},
							},
							{Key: "encoding", Component: "select", Flex: 8},
						},
						{{Key: "content", Component: "textarea", Flex: 24}},
					},
				},
				{
					Type:         "append",
					Label:        "Append",
					Description:  "Append content to the end of a file.",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
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
						{
							Key:     "encoding",
							Type:    "string",
							Default: "text",
							Label:   "Encoding",
							Options: []map[string]string{
								{"label": "Text", "value": "text"},
								{"label": "Base64", "value": "base64"},
								{"label": "Hex", "value": "hex"},
								{"label": "Data URL", "value": "data_url"},
							},
						},
					},
					AutoLayout: [][]node.AutoLayout{
						{
							{
								Key:       "path",
								Component: "input",
								Flex:      16,
								Options: map[string]any{
									"native_file_picker":           true,
									"native_file_picker_directory": true,
								},
							},
							{Key: "encoding", Component: "select", Flex: 8},
						},
						{{Key: "content", Component: "textarea", Flex: 24}},
					},
				},
				{
					Type:         "delete",
					Label:        "Delete",
					Description:  "Delete a file or directory.",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
					Fields: []node.NodeField{
						{
							Key:      "path",
							Type:     "string",
							Required: true,
							Label:    "Target file/folder",
						},
					},
					AutoLayout: [][]node.AutoLayout{
						{
							{
								Key:       "path",
								Component: "input",
								Flex:      24,
								Options: map[string]any{
									"native_file_picker":           true,
									"native_file_picker_multiple":  true,
									"native_file_picker_directory": true,
								},
							},
						},
					},
				},
			},
			OutputHandles: []string{"success", "error"},
		},
		ExecutorFactory: func(action node.Action) node.NodeExecutor {
			return &FSNode{Action: action}
		},
	})
}

func (n *FSNode) Execute(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	select {
	case <-ctx.Stop:
		return node.NodeExecutionResult{}, node.ErrNodeStopped
	default:
	}

	switch n.Action.Type {
	case "append":
		return n.executeAppend(ctx)
	case "delete":
		return n.executeDelete(ctx)
	default:
		return n.executeWrite(ctx)
	}
}

func (n *FSNode) executeWrite(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	pathRaw, _ := node.FieldValue(n.Action, "path").(string)
	contentRaw, _ := node.FieldValue(n.Action, "content").(string)
	encoding, _ := node.FieldValue(n.Action, "encoding").(string)

	if pathRaw == "" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("file path is required")
	}

	path := node.ParseTemplate(pathRaw, ctx.Variables)
	content := node.ParseTemplate(contentRaw, ctx.Variables)
	data, err := decodeContent(content, encoding)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
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
	pathRaw, _ := node.FieldValue(n.Action, "path").(string)
	contentRaw, _ := node.FieldValue(n.Action, "content").(string)
	encoding, _ := node.FieldValue(n.Action, "encoding").(string)

	if pathRaw == "" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("file path is required")
	}

	path := node.ParseTemplate(pathRaw, ctx.Variables)
	content := node.ParseTemplate(contentRaw, ctx.Variables)
	data, err := decodeContent(content, encoding)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to create directory: %v", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to open file for append: %v", err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to append file: %v", err)
	}

	detail, _ := getPathDetail(path)
	return node.NodeExecutionResult{
		Handle:   "success",
		Response: fmt.Sprintf("Successfully appended to: %s", detail),
		Type:     &node.NodeResponseType{Mime: "text/plain", Charset: "utf-8"},
	}, nil
}

func decodeContent(content, encoding string) ([]byte, error) {
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "", "text", "plain":
		return []byte(content), nil
	case "base64":
		if strings.HasPrefix(strings.TrimSpace(content), "data:") {
			return decodeDataURL(content)
		}
		return decodeBase64(content)
	case "hex":
		decoded, err := hex.DecodeString(compactEncodedContent(content))
		if err != nil {
			return nil, fmt.Errorf("failed to decode hex content: %v", err)
		}
		return decoded, nil
	case "data_url":
		return decodeDataURL(content)
	default:
		return nil, fmt.Errorf("unsupported encoding: %s", encoding)
	}
}

func decodeBase64(content string) ([]byte, error) {
	value := compactEncodedContent(content)
	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}
	for _, enc := range encodings {
		if decoded, err := enc.DecodeString(value); err == nil {
			return decoded, nil
		}
	}
	return nil, fmt.Errorf("failed to decode base64 content")
}

func decodeDataURL(content string) ([]byte, error) {
	value := strings.TrimSpace(content)
	if !strings.HasPrefix(value, "data:") {
		return nil, fmt.Errorf("data_url content must start with data:")
	}

	parts := strings.SplitN(value, ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid data_url content")
	}

	if strings.Contains(strings.ToLower(parts[0]), ";base64") {
		return decodeBase64(parts[1])
	}

	decoded, err := url.PathUnescape(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode data_url content: %v", err)
	}
	return []byte(decoded), nil
}

func compactEncodedContent(content string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
			return -1
		}
		return r
	}, strings.TrimSpace(content))
}

func (n *FSNode) executeDelete(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	pathRaw, _ := node.FieldValue(n.Action, "path").(string)

	if pathRaw == "" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("file path is required")
	}

	pathFull := node.ParseTemplate(pathRaw, ctx.Variables)
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
	switch deletedCount {
	case 0:
		if len(paths) == 1 {
			detail, _ := getPathDetail(strings.TrimSpace(paths[0]))
			responseMsg = fmt.Sprintf("%s not found", detail)
		} else {
			responseMsg = "No items were found to delete"
		}
	case 1:
		responseMsg = fmt.Sprintf("Deleted %s", details[0])
	default:
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
