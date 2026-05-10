package chromedpnode

import (
	"context"
	"ekken/internal/features/workflow/node"
	"ekken/internal/logger"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

func (n *BrowserNode) navigate(ctx *node.NodeContext, tabCtx context.Context) (node.NodeExecutionResult, error) {
	urlRaw, _ := n.Config["url"].(string)
	url := node.ParseTemplate(urlRaw, ctx.Variables)
	if url == "" {
		return node.NodeExecutionResult{}, fmt.Errorf("url is required")
	}

	logger.DevPrintf("[Browser] Navigating to: %s\n", url)
	err := chromedp.Run(tabCtx,
		chromedp.Navigate(url),
		chromedp.Sleep(500*time.Millisecond),
	)
	if err != nil {
		return node.NodeExecutionResult{}, err
	}
	return node.NodeExecutionResult{Handle: "success"}, nil
}
