package chromedpnode

import (
	"context"
	"ekken/internal/features/workflow/node"
	"ekken/internal/logger"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

func (n *BrowserNode) input(ctx *node.NodeContext, tabCtx context.Context) (node.NodeExecutionResult, error) {
	selector, queryOpt, err := n.getSelector(ctx)
	if err != nil {
		return node.NodeExecutionResult{}, err
	}
	valueRaw, _ := node.FieldValue(n.Action, "value").(string)
	value := node.ParseTemplate(valueRaw, ctx.Variables)
	logger.DevPrintf("[Browser] Typing into %s (length: %d)\n", selector, len(value))

	err = n.humanType(tabCtx, selector, queryOpt, value)
	if err != nil {
		return node.NodeExecutionResult{}, err
	}
	return node.NodeExecutionResult{Handle: "success"}, nil
}

func (n *BrowserNode) humanType(tabCtx context.Context, selector string, queryOpt chromedp.QueryOption, text string) error {
	// 1. Ensure the element is ready, visible, and focused.
	err := chromedp.Run(tabCtx,
		chromedp.WaitReady(selector, queryOpt),
		chromedp.ScrollIntoView(selector, queryOpt),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Focus(selector, queryOpt),
	)
	if err != nil {
		return fmt.Errorf("failed to focus element: %w", err)
	}

	// 2. Clear the field via JS, which is more reliable than document.activeElement.
	clearJS := fmt.Sprintf(`(function() {
    var el = document.querySelector(%q);
    if (el) {
        el.value = '';
        el.dispatchEvent(new Event('input', { bubbles: true }));
        el.dispatchEvent(new Event('change', { bubbles: true }));
    }
})()`, selector)

	if err := chromedp.Run(tabCtx, chromedp.Evaluate(clearJS, nil)); err != nil {
		logger.DevPrintf("[Browser] Failed to clear field: %v (continuing)\n", err)
	}

	// 3. Build all typing actions in one slice.
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	var actions []chromedp.Action

	for i, ch := range text {
		select {
		case <-tabCtx.Done():
			return fmt.Errorf("typing cancelled: %w", tabCtx.Err())
		default:
		}

		char := string(ch)
		if rng.Float64() < 0.05 && i < len(text)-1 {
			wrongChar := randomTypoChar(char, rng)
			logger.DevPrintf("[Browser] Typo: typed '%s' instead of '%s', correcting...\n", wrongChar, char)

			actions = append(actions, chromedp.SendKeys(selector, wrongChar, queryOpt))
			actions = append(actions, chromedp.Sleep(randomDelay(150, 300, rng)))
			actions = append(actions, chromedp.KeyEvent(kb.Backspace))
			actions = append(actions, chromedp.Sleep(randomDelay(80, 150, rng)))
			actions = append(actions, chromedp.SendKeys(selector, char, queryOpt))
		} else {
			actions = append(actions, chromedp.SendKeys(selector, char, queryOpt))
		}
		delay := typingDelay(i, len(text), rng)
		actions = append(actions, chromedp.Sleep(delay))
	}

	// 4. Execute all actions in one CDP transaction.
	// This prevents focus loss, background throttling, and race conditions.
	if err := chromedp.Run(tabCtx, actions...); err != nil {
		return fmt.Errorf("typing simulation failed: %w", err)
	}

	// 5. Press Enter if configured.
	if pressEnter, ok := node.FieldValue(n.Action, "press_enter").(bool); ok && pressEnter {
		time.Sleep(200 * time.Millisecond)
		if err := chromedp.Run(tabCtx, chromedp.SendKeys(selector, kb.Enter, queryOpt)); err != nil {
			logger.DevPrintf("[Browser] Failed to press Enter: %v\n", err)
		}
	}
	return nil
}

func randomTypoChar(correctChar string, rng *rand.Rand) string {
	neighbors := map[string][]string{
		"a": {"s", "w", "z"}, "b": {"n", "v", "g"}, "c": {"x", "v", "d"},
		"d": {"s", "f", "e"}, "e": {"w", "r", "d"}, "f": {"d", "g", "r"},
		"g": {"f", "h", "t"}, "h": {"g", "j", "y"}, "i": {"o", "u", "k"},
		"j": {"h", "k", "u"}, "k": {"j", "l", "i"}, "l": {"k", "o", "p"},
		"m": {"n", "k", "j"}, "n": {"b", "m", "h"}, "o": {"i", "p", "l"},
		"p": {"o", "l", "["}, "q": {"w", "a", "1"}, "r": {"e", "t", "f"},
		"s": {"a", "d", "w"}, "t": {"r", "y", "g"}, "u": {"y", "i", "j"},
		"v": {"c", "b", "f"}, "w": {"q", "e", "s"}, "x": {"z", "c", "s"},
		"y": {"t", "u", "h"}, "z": {"a", "x", "s"},
	}
	lower := strings.ToLower(correctChar)
	if list, ok := neighbors[lower]; ok && len(list) > 0 {
		wrong := list[rng.Intn(len(list))]
		if correctChar == strings.ToUpper(correctChar) && correctChar != lower {
			return strings.ToUpper(wrong)
		}
		return wrong
	}
	randomRune := 'a' + rune(rng.Intn(26))
	if correctChar == strings.ToUpper(correctChar) && correctChar != strings.ToLower(correctChar) {
		return strings.ToUpper(string(randomRune))
	}
	return string(randomRune)
}

func typingDelay(typedCount, totalLen int, rng *rand.Rand) time.Duration {
	minMs, maxMs := 40, 120
	if typedCount < 3 {
		minMs, maxMs = 150, 350
	} else if typedCount > totalLen-5 {
		minMs, maxMs = 70, 150
	}
	delay := rng.Intn(maxMs-minMs) + minMs
	return time.Duration(delay) * time.Millisecond
}

func randomDelay(minMs, maxMs int, rng *rand.Rand) time.Duration {
	return time.Duration(rng.Intn(maxMs-minMs)+minMs) * time.Millisecond
}
