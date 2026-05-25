package chromedpnode

import (
	"context"
	"encoding/base64"

	"ekken/internal/features/workflow/node"

	"github.com/chromedp/chromedp"
)

func (n *BrowserNode) screenshot(tabCtx context.Context) (node.NodeExecutionResult, error) {
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

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: base64.StdEncoding.EncodeToString(buf),
		Type:     &node.NodeResponseType{Mime: "image/png", Encoding: "base64"},
	}, nil
}
