package chromedpnode

import (
	"context"
	"fmt"
	"time"

	"ekken/internal/features/workflow/node"
	"ekken/internal/logger"
	"ekken/nodes/chrome"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

type Session struct {
	TabCtx context.Context
	Cancel context.CancelFunc
}

type BrowserNode struct {
	Action       node.Action
	Output       interface{}
	AllocatorCtx context.Context
}

func init() {
	node.GlobalRegistry.Register(node.NodeRegistration{
		Spec: node.Spec{
			Meta: node.Meta{
				Type:        "chromedp",
				Tags:        []string{"Web Automation"},
				Label:       "Chromedp",
				Icon:        "https://www.svgrepo.com/show/522006/browser.svg",
				Description: "Navigate, click, screenshot from a Chrome tab.",
				DependsOn: []node.DependsOn{
					{Node: "google_chrome", Action: "launch"},
				},
			},

			DefaultAction: "navigate",
			Actions: []node.Action{
				{
					Type:        "navigate",
					Label:       "Navigate",
					Description: "Navigate to a URL",
					Fields: []node.NodeField{
						{Key: "url", Type: "string", Required: true, Label: "URL"},
					},
					AutoLayout: [][]node.AutoLayout{
						{
							{Key: "url", Component: "input", Flex: 24, Options: map[string]any{"placeholder": "https://example.com"}},
						},
					},
				},
				{
					Type:        "click",
					Label:       "Click",
					Description: "Click an element by selector",
					Fields: []node.NodeField{
						{Key: "selector_type", Type: "string", Default: "css", Label: "Selector Type", Options: []map[string]string{
							{"label": "CSS", "value": "css"},
							{"label": "XPath", "value": "xpath"},
						}},
						{Key: "selector", Type: "string", Required: true, Label: "Selector"},
					},
					AutoLayout: [][]node.AutoLayout{
						{
							{Key: "selector_type", Component: "select", Flex: 8},
							{Key: "selector", Component: "input", Flex: 16, Options: map[string]any{"placeholder": "button.submit-btn"}},
						},
					},
				},
				{
					Type:        "screenshot",
					Label:       "Screenshot",
					Description: "Take a screenshot of the page",
					Fields: []node.NodeField{
						{Key: "path", Type: "string", Required: true, Label: "Output Path"},
						{Key: "full_page", Type: "boolean", Default: false, Label: "Full Page"},
					},
					AutoLayout: [][]node.AutoLayout{
						{
							{Key: "path", Component: "input", Flex: 18, Options: map[string]any{"placeholder": "/path/to/screenshot.png"}},
							{Key: "full_page", Component: "switch", Flex: 6, Options: map[string]any{"helper": "Capture full scrollable page"}},
						},
					},
				},
				{
					Type:        "input",
					Label:       "Input",
					Description: "Type into an input field",
					Fields: []node.NodeField{
						{Key: "selector_type", Type: "string", Default: "css", Label: "Selector Type", Options: []map[string]string{
							{"label": "CSS", "value": "css"},
							{"label": "XPath", "value": "xpath"},
						}},
						{Key: "selector", Type: "string", Required: true, Label: "Selector"},
						{Key: "value", Type: "string", Required: true, Label: "Value to Type"},
						{Key: "press_enter", Type: "boolean", Default: false, Label: "Press Enter"},
					},
					AutoLayout: [][]node.AutoLayout{
						{
							{Key: "selector_type", Component: "select", Flex: 8},
							{Key: "selector", Component: "input", Flex: 16, Options: map[string]any{"placeholder": "input[name='search']"}},
						},
						{
							{Key: "value", Component: "input", Flex: 18, Options: map[string]any{"placeholder": "Hello world"}},
							{Key: "press_enter", Component: "switch", Flex: 6, Options: map[string]any{"helper": "After typing"}},
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
			return &BrowserNode{Action: action}
		},
	})
}

func (n *BrowserNode) Execute(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	// Proactive check: If workflow is already stopped, return immediately
	if err := ctx.Context.Err(); err != nil {
		return node.NodeExecutionResult{}, err
	}

	action := n.Action.Type
	if action == "" {
		return node.NodeExecutionResult{}, fmt.Errorf("action is required")
	}
	logger.DevPrintf("[Browser] Starting Execute: %s\n", action)

	tabCtx, err := n.getOrCreateTab(ctx)
	if err != nil {
		return node.NodeExecutionResult{}, err
	}

	timeoutSec := 60.0
	if t, ok := node.FieldValue(n.Action, "timeout").(float64); ok && t > 0 {
		timeoutSec = t
	}

	// Link engine context (ctx.Context) with tab context (tabCtx)
	// This ensures the action stops immediately if the workflow is stopped by the user.
	runCtx, runCancel := context.WithCancel(tabCtx)
	go func() {
		select {
		case <-ctx.Context.Done():
			runCancel()
		case <-runCtx.Done():
		}
	}()
	defer runCancel()

	timeoutCtx, cancel := context.WithTimeout(runCtx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	return n.runAction(ctx, timeoutCtx, action)
}

func (n *BrowserNode) getOrCreateTab(ctx *node.NodeContext) (context.Context, error) {
	rawCtx, exists := ctx.Metadata["browser_ctx"]
	var tabCtx context.Context
	if exists {
		tabCtx, _ = rawCtx.(context.Context)
	}
	rawCancel := ctx.Metadata["browser_cancel"]
	var cancelFunc context.CancelFunc
	if rawCancel != nil {
		cancelFunc, _ = rawCancel.(context.CancelFunc)
	}

	if tabCtx != nil && n.isTabHealthy(tabCtx) {
		logger.DevPrintf("[Browser] Reusing existing healthy tab\n")
		return tabCtx, nil
	}

	if cancelFunc != nil {
		logger.DevPrintf("[Browser] Tab unhealthy, cleaning up old session...\n")
		cancelFunc()
		delete(ctx.Metadata, "browser_ctx")
		delete(ctx.Metadata, "browser_cancel")
	}

	logger.DevPrintf("[Browser] Creating new tab\n")
	session, err := n.createTab()
	if err != nil {
		return nil, err
	}
	ctx.Metadata["browser_ctx"] = session.TabCtx
	ctx.Metadata["browser_cancel"] = session.Cancel
	ctx.OnCleanup = append(ctx.OnCleanup, session.Cancel)
	return session.TabCtx, nil
}

func (n *BrowserNode) isTabHealthy(tabCtx context.Context) bool {
	checkCtx, cancel := context.WithTimeout(tabCtx, 3*time.Second)
	defer cancel()
	var url string
	err := chromedp.Run(checkCtx, chromedp.Location(&url))
	return err == nil
}

func (n *BrowserNode) createTab() (*Session, error) {
	// Gunakan AllocatorCtx jika ada, jika tidak fallback ke GlobalAllocCtx agar tidak error
	allocCtx := n.AllocatorCtx
	if allocCtx == nil {
		allocCtx = chrome.GlobalAllocCtx
	}

	if err := chrome.EnsureBrowser(allocCtx, false); err != nil {
		return nil, fmt.Errorf("failed to ensure browser: %w", err)
	}

	// Double check setelah EnsureBrowser karena GlobalAllocCtx baru terisi di sana
	if allocCtx == nil {
		allocCtx = chrome.GlobalAllocCtx
	}

	if allocCtx == nil {
		return nil, fmt.Errorf("browser allocator context is not initialized")
	}

	tabCtx, cancel := chromedp.NewContext(allocCtx)
	err := chromedp.Run(tabCtx, chromedp.ActionFunc(func(ctx context.Context) error {
		targets, err := target.GetTargets().Do(ctx)
		if err != nil {
			return err
		}
		for _, t := range targets {
			if t.Type == "page" || t.Type == "tab" {
				return target.ActivateTarget(t.TargetID).Do(ctx)
			}
		}
		return page.BringToFront().Do(ctx)
	}))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to init tab: %w", err)
	}
	return &Session{
		TabCtx: tabCtx,
		Cancel: cancel,
	}, nil
}

func (n *BrowserNode) runAction(ctx *node.NodeContext, tabCtx context.Context, action string) (node.NodeExecutionResult, error) {
	switch action {
	case "navigate":
		return n.navigate(ctx, tabCtx)
	case "click":
		return n.click(ctx, tabCtx)
	case "screenshot":
		return n.screenshot(ctx, tabCtx)
	case "input":
		return n.input(ctx, tabCtx)
	default:
		return node.NodeExecutionResult{}, fmt.Errorf("unknown action: %s", action)
	}
}

// getSelector is a helper to commonize selector parsing across actions.
func (n *BrowserNode) getSelector(ctx *node.NodeContext) (string, chromedp.QueryOption, error) {
	selectorRaw, _ := node.FieldValue(n.Action, "selector").(string)
	selector := node.ParseTemplate(selectorRaw, ctx.Variables)
	if selector == "" {
		return "", nil, fmt.Errorf("selector is required")
	}

	selectorType, _ := node.FieldValue(n.Action, "selector_type").(string)
	queryOpt := chromedp.ByQuery
	if selectorType == "xpath" {
		queryOpt = chromedp.BySearch
	}
	return selector, queryOpt, nil
}
