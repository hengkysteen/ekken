package chromedpnode

import (
	"context"
	"ekken/internal/features/workflow/node"
	"ekken/internal/logger"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
)

func (n *BrowserNode) click(ctx *node.NodeContext, tabCtx context.Context) (node.NodeExecutionResult, error) {
	selector, queryOpt, err := n.getSelector(ctx)
	if err != nil {
		return node.NodeExecutionResult{}, err
	}

	selectorType, _ := n.Config["selector_type"].(string)
	logger.DevPrintf("[Browser] Clicking selector (%s): %s\n", selectorType, selector)
	// Get URL before clicking
	var urlBefore string
	if err := chromedp.Run(tabCtx, chromedp.Location(&urlBefore)); err != nil {
		return node.NodeExecutionResult{}, fmt.Errorf("failed to get initial URL: %w", err)
	}
	// Preparation: Try to make the element visible & scroll
	// We don't return error immediately here to allow Strategy 1 (JS Click) a chance to try
	readyErr := chromedp.Run(tabCtx,
		chromedp.WaitVisible(selector, queryOpt),
		chromedp.ScrollIntoView(selector, queryOpt),
		chromedp.Sleep(150*time.Millisecond),
	)
	if readyErr != nil {
		logger.DevPrintf("[Browser] Warning: Element not ready/visible (%v), trying alternative strategies...\n", readyErr)
	}
	// Inject click listener
	_ = n.injectClickListener(tabCtx, selector, selectorType)
	// Strategy 1: JS click (most effective if selector is correct but element is obstructed/hidden)
	logger.DevPrintf("[Browser] Strategy 1: JS click\n")

	jsClick := buildJS(selector, selectorType, `el.click();`)
	var jsSuccess bool
	err = chromedp.Run(tabCtx, chromedp.Evaluate(jsClick, &jsSuccess))
	if err == nil && jsSuccess {
		// Wait a bit to see if there is a reaction (navigation or event)
		select {
		case <-tabCtx.Done():
			return node.NodeExecutionResult{}, tabCtx.Err()
		case <-time.After(300 * time.Millisecond):
		}
		if ok, url := n.verifyClick(tabCtx, selector, selectorType, urlBefore); ok {
			return successResult(url), nil
		}
	} else if !jsSuccess {
		// If it's not even in the DOM, then we can give up or report a selector error
		return node.NodeExecutionResult{}, fmt.Errorf("element not found in DOM: %s", selector)
	}
	// If Strategy 1 fails but element is in DOM, we need position for mouse strategies
	var box map[string]float64
	getBoxJS := buildJS(selector, selectorType, `
		var b = el.getBoundingClientRect();
		return {x: b.left + b.width/2, y: b.top + b.height/2};
	`)
	if err := chromedp.Run(tabCtx, chromedp.Evaluate(getBoxJS, &box)); err != nil {
		return node.NodeExecutionResult{}, fmt.Errorf("failed to get element position: %w", err)
	}
	cx, cy := box["x"], box["y"]

	// Strategy 2: manual MouseEvent dispatch (bypass framework synthetic events)
	logger.DevPrintf("[Browser] Strategy 2: dispatch MouseEvent\n")
	n.injectClickListener(tabCtx, selector, selectorType)
	if err := chromedp.Run(tabCtx,
		chromedp.MouseEvent(input.MousePressed, cx, cy, chromedp.Button("left")),
		chromedp.MouseEvent(input.MouseReleased, cx, cy, chromedp.Button("left")),
	); err == nil {
		if ok, url := n.verifyClick(tabCtx, selector, selectorType, urlBefore); ok {
			return successResult(url), nil
		}
	}
	// Strategy 3: native chromedp click
	logger.DevPrintf("[Browser] Strategy 3: native chromedp click\n")
	n.injectClickListener(tabCtx, selector, selectorType)
	if err := chromedp.Run(tabCtx, chromedp.Click(selector, queryOpt)); err == nil {
		if ok, url := n.verifyClick(tabCtx, selector, selectorType, urlBefore); ok {
			return successResult(url), nil
		}
	}
	// Strategy 4: full mouse simulation (slowest, last resort)
	logger.DevPrintf("[Browser] Strategy 4: full mouse simulation\n")
	n.injectClickListener(tabCtx, selector, selectorType)
	if err := n.mouseClick(tabCtx, cx, cy); err == nil {
		if ok, url := n.verifyClick(tabCtx, selector, selectorType, urlBefore); ok {
			return successResult(url), nil
		}
	}
	return node.NodeExecutionResult{}, fmt.Errorf("all click strategies failed for selector: %s", selector)
}

// mouseClick move mouse from random position -> hover -> click
func (n *BrowserNode) mouseClick(tabCtx context.Context, cx, cy float64) error {
	// Start from a random position near the element (not from 0,0)
	startX := cx + float64(rand.Intn(100)-50)
	startY := cy + float64(rand.Intn(100)-50)
	return chromedp.Run(tabCtx,
		// Move to start position first
		chromedp.MouseEvent(input.MouseMoved, startX, startY),
		chromedp.Sleep(50*time.Millisecond),
		// Move to target gradually (simulated mouse path)
		chromedp.MouseEvent(input.MouseMoved, (startX+cx)/2, (startY+cy)/2),
		chromedp.Sleep(30*time.Millisecond),
		chromedp.MouseEvent(input.MouseMoved, cx, cy),
		chromedp.Sleep(80*time.Millisecond), // hover briefly
		// Click
		chromedp.MouseEvent(input.MousePressed, cx, cy, chromedp.Button("left"), chromedp.ClickCount(1)),
		chromedp.Sleep(50*time.Millisecond),
		chromedp.MouseEvent(input.MouseReleased, cx, cy, chromedp.Button("left"), chromedp.ClickCount(1)),
		chromedp.Sleep(300*time.Millisecond),
	)
}

// injectClickListener injects __clickReceived listener to the element
func (n *BrowserNode) injectClickListener(tabCtx context.Context, selector, selectorType string) error {
	js := buildJS(selector, selectorType, `
		el.__clickReceived = false;
		el.addEventListener('click', function() { el.__clickReceived = true; }, { once: true });
	`)
	var ok bool
	return chromedp.Run(tabCtx, chromedp.Evaluate(js, &ok))
}

// verifyClick check if the click was actually received
func (n *BrowserNode) verifyClick(tabCtx context.Context, selector, selectorType, urlBefore string) (bool, string) {
	var urlAfter string
	chromedp.Run(tabCtx, chromedp.Location(&urlAfter))
	// Navigation occurred = click definitely accepted
	if urlAfter != urlBefore {
		logger.DevPrintf("[Browser] Verified: navigation to %s\n", urlAfter)
		return true, urlAfter
	}
	// Check listener
	checkJS := buildJS(selector, selectorType, `return !!el.__clickReceived;`)
	var received bool
	chromedp.Run(tabCtx, chromedp.Evaluate(checkJS, &received))
	if received {
		logger.DevPrintf("[Browser] Verified: click event accepted by element\n")
		return true, urlAfter
	}
	return false, urlAfter
}

// buildJS helper to build JS that supports CSS selector and XPath
func buildJS(selector, selectorType, body string) string {
	// Escape single quote to avoid conflicts within the JS string
	escapedSelector := strings.ReplaceAll(selector, `'`, `\'`)
	if selectorType == "xpath" {
		return fmt.Sprintf(`
			(function() {
				var el = document.evaluate('%s', document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null).singleNodeValue;
				if (!el) return false;
				%s
				return true;
			})()
		`, escapedSelector, body)
	}
	return fmt.Sprintf(`
		(function() {
			var sel = '%s';
			var el = document.querySelector(sel);
			
			// Auto-Fuzzy Fallback: If failed and this is an exact href link match, try fuzzy match
			if (!el && sel.includes('[href="')) {
				var fuzzy = sel.replace('[href="', '[href*="');
				el = document.querySelector(fuzzy);
			}
			if (!el) return false;
			%s
			return true;
		})()
	`, escapedSelector, body)
}

// successResult helper buat NodeExecutionResult
func successResult(url string) node.NodeExecutionResult {
	return node.NodeExecutionResult{
		Handle:   "success",
		Response: map[string]string{"url": url},
	}
}
