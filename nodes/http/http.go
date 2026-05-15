package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ekken/internal/features/workflow/node"
	"ekken/internal/logger"
)

type HTTPNode struct {
	Action node.NodeAction
	Output any
}

func init() {
	node.GlobalRegistry.Register(node.NodeRegistration{
		NodeSpec: node.NodeSpec{
			NodeMetadata: node.NodeMetadata{
				Type:        "http",
				Tags:        []string{"Network"},
				Label:       "HTTP",
				Icon:        "https://www.svgrepo.com/show/221325/http.svg",
				Description: "Call HTTP endpoints with custom methods, headers, and body.",
			},

			DefaultAction: "http_request",
			Actions: []node.NodeAction{
				{
					Key:         "http_request",
					Label:       "HTTP Request",
					Description: "Send an HTTP request to an endpoint",
					HasResponse: true,
					ResponseType: &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
					Fields: []node.NodeField{
						{Key: "url", Type: "string", Required: true, Label: "Endpoint URL"},
						{
							Key:     "method",
							Type:    "string",
							Default: "GET",
							Label:   "HTTP method to use",
							Options: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
						},
						{Key: "headers", Type: "string", Label: "Custom headers (Key: Value per line)"},
						{Key: "body", Type: "string", Label: "Request body template"},
						{Key: "timeout", Type: "number", Default: 60, Label: "Timeout in seconds"},
						{Key: "retry_count", Type: "number", Label: "Retry attempts on failure"},
					},
				},
			},
			Outputs: []node.HandleEdge{
				{Key: "success", Label: "Success", Tone: "success"},
				{Key: "error", Label: "Error", Tone: "error"},
			},
		},
		ExecutorFactory: func(action node.NodeAction) node.NodeExecutor {
			return &HTTPNode{Action: action}
		},
	})
}

func (n *HTTPNode) Execute(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	action := n.Action.Key
	if action == "" {
		action = "http_request"
	}

	logger.Info(fmt.Sprintf("[HTTP] Starting Execute Action: %s", action))

	switch action {
	case "http_request":
		return n.runRequest(ctx)
	default:
		return node.NodeExecutionResult{}, fmt.Errorf("unknown action: %s", action)
	}
}

func (n *HTTPNode) runRequest(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	url, _ := node.FieldValue(n.Action, "url").(string)
	if url == "" {
		return node.NodeExecutionResult{}, fmt.Errorf("URL is required")
	}

	method, _ := node.FieldValue(n.Action, "method").(string)
	if method == "" {
		method = "GET"
	}
	headersRaw, _ := node.FieldValue(n.Action, "headers").(string)
	bodyTemplate, _ := node.FieldValue(n.Action, "body").(string)
	timeoutSec, _ := node.FieldValue(n.Action, "timeout").(float64)
	if timeoutSec <= 0 {
		timeoutSec = 60
	}

	// Parsing templates
	url = node.ParseTemplate(url, ctx.Variables)
	method = strings.ToUpper(node.ParseTemplate(method, ctx.Variables))
	bodyStr := node.ParseTemplate(bodyTemplate, ctx.Variables)
	headersRaw = node.ParseTemplate(headersRaw, ctx.Variables)

	var body io.Reader
	if bodyStr != "" {
		body = bytes.NewBufferString(bodyStr)
	}

	req, err := http.NewRequestWithContext(ctx.Context, method, url, body)
	if err != nil {
		return node.NodeExecutionResult{}, fmt.Errorf("failed to create request: %v", err)
	}

	// Set default content-type if not provided
	hasContentType := false

	// Parse custom headers
	lines := strings.SplitSeq(headersRaw, "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			req.Header.Set(key, val)
			if strings.ToLower(key) == "content-type" {
				hasContentType = true
			}
		}
	}

	if !hasContentType && (method == "POST" || method == "PUT" || method == "PATCH") {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{
		Timeout: time.Duration(timeoutSec) * time.Second,
	}

	logger.Info(fmt.Sprintf("[HTTP] %s %s...", method, url))
	resp, err := client.Do(req)
	if err != nil {
		return node.NodeExecutionResult{}, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return node.NodeExecutionResult{}, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode >= 400 {
		return node.NodeExecutionResult{}, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	bodyRaw := string(respBody)

	var bodyParsed any = bodyRaw
	if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		var parsed any
		if err := json.Unmarshal(respBody, &parsed); err == nil {
			bodyParsed = parsed
		}
	}

	result := map[string]any{
		"status_code": resp.StatusCode,
		"body":        bodyParsed,
		"headers":     resp.Header,
	}

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: result,
		Type:     &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
	}, nil
}
