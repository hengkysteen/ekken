package http

import (
	"bytes"
	"encoding/base64"
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
	Action node.Action
}

type curlRequest struct {
	Method  string
	URL     string
	Headers http.Header
	Body    string
}

func init() {
	node.GlobalRegistry.Register(node.NodeRegistration{
		Spec: node.Spec{
			Meta: node.Meta{
				Type:        "http",
				Tags:        []string{"Network"},
				Label:       "HTTP",
				Icon:        "https://www.svgrepo.com/show/221325/http.svg",
				Description: "Call HTTP endpoints from a curl command.",
			},
			DefaultAction: "curl",
			Actions: []node.Action{
				{
					Type:         "curl",
					Label:        "Curl",
					Description:  "Send an HTTP request using curl syntax",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
					Fields: []node.NodeField{
						{
							Key:      "curl",
							Type:     "string",
							Required: true,
							Label:    "Curl command",
						},
					},
					AutoLayout: [][]node.AutoLayout{
						{
							{
								Key:       "curl",
								Component: "textarea",
								Flex:      24,
							},
						},
					},
				},
			},
			GlobalFields: []node.NodeField{
				{Key: "timeout", Type: "number", Default: 60, Label: "Timeout (sec)"},
				{Key: "retry_count", Type: "number", Default: 0, Label: "Retry Count"},
			},
			Outputs: []node.HandleEdge{
				{Key: "success", Label: "Success", Tone: "success"},
				{Key: "error", Label: "Error", Tone: "error"},
			},
		},
		ExecutorFactory: func(action node.Action) node.NodeExecutor {
			return &HTTPNode{Action: action}
		},
	})
}

func (n *HTTPNode) Execute(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	select {
	case <-ctx.Stop:
		return node.NodeExecutionResult{}, node.ErrNodeStopped
	default:
	}

	action := n.Action.Type
	if action == "" {
		action = "curl"
	}
	if action != "curl" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("unknown action: %s", action)
	}

	return n.executeCurl(ctx)
}

func (n *HTTPNode) executeCurl(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	curlRaw, _ := node.FieldValue(n.Action, "curl").(string)
	if strings.TrimSpace(curlRaw) == "" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("curl command is required")
	}

	curlCommand := node.ParseTemplate(curlRaw, ctx.Variables)
	parsed, err := parseCurl(curlCommand)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, err
	}

	timeoutSec := 60.0
	if v, ok := node.FieldValue(n.Action, "timeout").(float64); ok && v > 0 {
		timeoutSec = v
	}

	var body io.Reader
	if parsed.Body != "" {
		body = bytes.NewBufferString(parsed.Body)
	}

	req, err := http.NewRequestWithContext(ctx.Context, parsed.Method, parsed.URL, body)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header = parsed.Headers.Clone()
	if parsed.Body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}

	logger.Info(fmt.Sprintf("[HTTPV2] %s %s...", parsed.Method, parsed.URL))
	resp, err := client.Do(req)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode >= 400 {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	bodyRaw := string(respBody)
	var bodyParsed any = bodyRaw
	contentType := resp.Header.Get("Content-Type")
	if shouldParseJSON(contentType, bodyRaw) {
		var parsedBody any
		if err := json.Unmarshal(respBody, &parsedBody); err == nil {
			bodyParsed = parsedBody
		}
	}

	return node.NodeExecutionResult{
		Handle: "success",
		Response: map[string]any{
			"status_code":  resp.StatusCode,
			"content_type": contentType,
			"body":         bodyParsed,
			"headers":      resp.Header,
		},
		Type: &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
	}, nil
}

func shouldParseJSON(contentType, body string) bool {
	contentType = strings.ToLower(contentType)
	if strings.Contains(contentType, "application/json") {
		return true
	}

	body = strings.TrimSpace(body)
	return !strings.Contains(contentType, "text/event-stream") &&
		(strings.HasPrefix(body, "{") || strings.HasPrefix(body, "["))
}

func parseCurl(command string) (curlRequest, error) {
	args, err := splitCommand(command)
	if err != nil {
		return curlRequest{}, err
	}
	if len(args) == 0 {
		return curlRequest{}, fmt.Errorf("curl command is required")
	}
	if args[0] == "curl" {
		args = args[1:]
	}

	req := curlRequest{
		Method:  "GET",
		Headers: make(http.Header),
	}

	var dataParts []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-X" || arg == "--request":
			val, next, err := optionValue(args, i, arg)
			if err != nil {
				return curlRequest{}, err
			}
			req.Method = strings.ToUpper(val)
			i = next
		case strings.HasPrefix(arg, "-X") && arg != "-X":
			req.Method = strings.ToUpper(strings.TrimPrefix(arg, "-X"))
		case strings.HasPrefix(arg, "--request="):
			req.Method = strings.ToUpper(strings.TrimPrefix(arg, "--request="))
		case arg == "--url":
			val, next, err := optionValue(args, i, arg)
			if err != nil {
				return curlRequest{}, err
			}
			req.URL = val
			i = next
		case strings.HasPrefix(arg, "--url="):
			req.URL = strings.TrimPrefix(arg, "--url=")
		case arg == "-H" || arg == "--header":
			val, next, err := optionValue(args, i, arg)
			if err != nil {
				return curlRequest{}, err
			}
			addHeader(req.Headers, val)
			i = next
		case strings.HasPrefix(arg, "-H") && arg != "-H":
			addHeader(req.Headers, strings.TrimPrefix(arg, "-H"))
		case strings.HasPrefix(arg, "--header="):
			addHeader(req.Headers, strings.TrimPrefix(arg, "--header="))
		case isDataOption(arg):
			val, next, err := optionValue(args, i, arg)
			if err != nil {
				return curlRequest{}, err
			}
			dataParts = append(dataParts, val)
			i = next
		case hasDataOptionPrefix(arg):
			dataParts = append(dataParts, valueAfterEqualsOrShortPrefix(arg))
		case arg == "-u" || arg == "--user":
			val, next, err := optionValue(args, i, arg)
			if err != nil {
				return curlRequest{}, err
			}
			req.Headers.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(val)))
			i = next
		case strings.HasPrefix(arg, "-u") && arg != "-u":
			req.Headers.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(strings.TrimPrefix(arg, "-u"))))
		case strings.HasPrefix(arg, "--user="):
			req.Headers.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(strings.TrimPrefix(arg, "--user="))))
		case arg == "-A" || arg == "--user-agent":
			val, next, err := optionValue(args, i, arg)
			if err != nil {
				return curlRequest{}, err
			}
			req.Headers.Set("User-Agent", val)
			i = next
		case strings.HasPrefix(arg, "-A") && arg != "-A":
			req.Headers.Set("User-Agent", strings.TrimPrefix(arg, "-A"))
		case strings.HasPrefix(arg, "--user-agent="):
			req.Headers.Set("User-Agent", strings.TrimPrefix(arg, "--user-agent="))
		case arg == "-I" || arg == "--head":
			req.Method = "HEAD"
		case isIgnoredFlag(arg):
		case strings.HasPrefix(arg, "-"):
			if optionNeedsValue(arg) {
				i++
			}
		default:
			req.URL = arg
		}
	}

	if len(dataParts) > 0 {
		req.Body = strings.Join(dataParts, "&")
		if req.Method == "GET" {
			req.Method = "POST"
		}
	}
	if req.URL == "" {
		return curlRequest{}, fmt.Errorf("curl URL is required")
	}
	if req.Method == "" {
		req.Method = "GET"
	}

	return req, nil
}

func splitCommand(command string) ([]string, error) {
	command = strings.NewReplacer(
		"\\\r\n", " ",
		"\\\n", " ",
		"\\\r", " ",
	).Replace(command)

	var args []string
	var current strings.Builder
	var quote rune
	escaped := false

	for _, r := range command {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				current.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
		case r == ' ' || r == '\t' || r == '\n' || r == '\r':
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if escaped {
		current.WriteRune('\\')
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated quote in curl command")
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args, nil
}

func optionValue(args []string, i int, opt string) (string, int, error) {
	if i+1 >= len(args) {
		return "", i, fmt.Errorf("%s requires a value", opt)
	}
	return args[i+1], i + 1, nil
}

func addHeader(headers http.Header, raw string) {
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) != 2 {
		return
	}
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	if key == "" {
		return
	}
	headers.Set(key, value)
}

func isDataOption(arg string) bool {
	switch arg {
	case "-d", "--data", "--data-raw", "--data-binary", "--data-ascii", "--form", "-F":
		return true
	default:
		return false
	}
}

func hasDataOptionPrefix(arg string) bool {
	return strings.HasPrefix(arg, "-d") && arg != "-d" ||
		strings.HasPrefix(arg, "--data=") ||
		strings.HasPrefix(arg, "--data-raw=") ||
		strings.HasPrefix(arg, "--data-binary=") ||
		strings.HasPrefix(arg, "--data-ascii=") ||
		strings.HasPrefix(arg, "--form=") ||
		strings.HasPrefix(arg, "-F") && arg != "-F"
}

func valueAfterEqualsOrShortPrefix(arg string) string {
	if idx := strings.Index(arg, "="); idx >= 0 {
		return arg[idx+1:]
	}
	if strings.HasPrefix(arg, "-d") {
		return strings.TrimPrefix(arg, "-d")
	}
	if strings.HasPrefix(arg, "-F") {
		return strings.TrimPrefix(arg, "-F")
	}
	return ""
}

func optionNeedsValue(arg string) bool {
	switch arg {
	case "-b", "--cookie", "-e", "--referer", "--connect-timeout", "--max-time", "-o", "--output":
		return true
	default:
		return false
	}
}

func isIgnoredFlag(arg string) bool {
	switch arg {
	case "-L", "--location", "--compressed", "--silent", "-s", "--request-no-buffer":
		return true
	default:
		return false
	}
}
