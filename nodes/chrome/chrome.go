package chrome

import (
	"context"
	"ekken/internal/features/workflow/node"
	"ekken/internal/logger"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type GoogleChromeNode struct {
	Config map[string]any
	Output any
}

var (
	GlobalAllocCtx context.Context
	GlobalCancel   context.CancelFunc
	configBin      string
	configPort     int = 9222
	configProfile  string
	configHeadless bool
	configWidth    int                 = 1920
	configHeight   int                 = 1080
	activeProcs    map[int]*os.Process = make(map[int]*os.Process)
	mu             sync.Mutex
)

func init() {
	// 1. Load config from environment variables
	if bin := os.Getenv("EKKENCHROME_BIN"); bin != "" {
		configBin = bin
	}
	if portStr := os.Getenv("EKKENCHROME_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			configPort = p
		}
	}
	if profile := os.Getenv("EKKENCHROME_PROFILE"); profile != "" {
		configProfile = profile
	}

	// 2. Register node to global registry
	node.GlobalRegistry.Register(node.NodeRegistration{
		NodeSpec: node.NodeSpec{
			NodeMetadata: node.NodeMetadata{
				Type:  "google_chrome",
				Tags:  []string{"Browser"},
				Label: "Google Chrome",
				Icon:  "https://www.svgrepo.com/show/496944/chrome.svg",
			},
			Description:   "Launch or Terminate the global Google Chrome instance.",
			DefaultAction: "launch",
			Actions: []node.NodeAction{
				{
					Key:         "launch",
					Label:       "Launch",
					Description: "Launch Google Chrome with debugging port",
					Fields: []node.NodeField{
						{Key: "bin_path", Type: "string", Default: getChromePath(), Label: "Google Chrome Path"},
						{Key: "profile", Type: "string", Default: "mybot", Label: "Profile"},
						{Key: "port", Type: "number", Default: 9222, Label: "Port"},
						{Key: "width", Type: "number", Default: 1920, Label: "Window Width"},
						{Key: "height", Type: "number", Default: 1080, Label: "Window Height"},
						{Key: "headless", Type: "boolean", Default: false, Label: "Headless"},
					},
					Form: [][]node.Form{
						{
							{Key: "bin_path", Component: "input", Flex: 24, FormOptions: map[string]any{"native_file_picker": true}},
						},
						{
							{Key: "profile", Component: "input", Flex: 12},
							{Key: "port", Component: "number", Flex: 12},
						},
						{
							{Key: "width", Component: "number", Flex: 12},
							{Key: "height", Component: "number", Flex: 12},
						},
						{
							{Key: "headless", Component: "switch", FormOptions: map[string]any{"helper": "Run Chrome without a GUI window"}},
						},
					},
				},
				{
					Key:         "terminate",
					Label:       "Terminate",
					Description: "Terminate Google Chrome instance",
					Fields: []node.NodeField{
						{Key: "port", Type: "number", Default: 9222, Label: "Port"},
					},
					Form: [][]node.Form{
						{
							{Key: "port", Component: "number", FormOptions: map[string]any{"helper": "Chrome debugging port to terminate"}},
						},
					},
				},
			},
			Outputs: []node.NodeOutputDef{
				{Key: "success", Label: "Success", Tone: "success"},
				{Key: "error", Label: "Error", Tone: "error"},
			},
		},
		ExecutorFactory: func(config map[string]interface{}, _ []node.Node) node.NodeExecutor {
			return &GoogleChromeNode{Config: config}
		},
	})
}

func (n *GoogleChromeNode) Execute(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	action, _ := n.Config["action"].(string)
	action = strings.ToLower(action)
	portF, _ := n.Config["port"].(float64)
	port := int(portF)
	if port == 0 {
		port = 9222
	}
	profile, _ := n.Config["profile"].(string)
	if profile == "" {
		profile = "mybot"
	}
	resolvedProfile := node.ParseTemplate(profile, ctx.Variables)
	logger.DevPrintf("[System] Executing Chrome %s on port %d... (Profile: %s)\n", action, port, resolvedProfile)
	switch action {
	case "launch":
		return n.launch(ctx, port, resolvedProfile)
	case "terminate":
		return n.terminate(port)
	default:
		return node.NodeExecutionResult{}, fmt.Errorf("unknown action %s for google chrome node", action)
	}
}
