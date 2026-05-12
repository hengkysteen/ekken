package chromedpnode

import (
	"context"
	"ekken/internal/features/workflow/node"
	"fmt"
	"os"
	"path/filepath"

	"github.com/chromedp/chromedp"
)

func (n *BrowserNode) screenshot(ctx *node.NodeContext, tabCtx context.Context) (node.NodeExecutionResult, error) {
	pathRaw, _ := node.FieldValue(n.Action, "path").(string)
	path := node.ParseTemplate(pathRaw, ctx.Variables)
	if path == "" {
		return node.NodeExecutionResult{}, fmt.Errorf("path is required")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return node.NodeExecutionResult{}, fmt.Errorf("failed to create directory for screenshot: %v", err)
	}
	fullPage, _ := node.FieldValue(n.Action, "full_page").(bool)

	var buf []byte
	var err error
	if fullPage {
		err = chromedp.Run(tabCtx, chromedp.FullScreenshot(&buf, 100))
	} else {
		err = chromedp.Run(tabCtx, chromedp.CaptureScreenshot(&buf))
	}

	if err != nil {
		return node.NodeExecutionResult{}, err
	}
	err = os.WriteFile(path, buf, 0o644)
	if err != nil {
		return node.NodeExecutionResult{}, err
	}
	return node.NodeExecutionResult{Handle: "success"}, nil
}
