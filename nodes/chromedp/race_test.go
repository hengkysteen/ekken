package chromedpnode

import (
	"context"
	"ekken/internal/features/workflow/node"
	"sync"
	"testing"
)

// TestConcurrentMetadataAccess checks whether ctx.Metadata has a data race
// when multiple browser nodes access it concurrently.
func TestConcurrentMetadataAccess(t *testing.T) {
	// Set up a shared NodeContext, like nodes in the same workflow.
	ctx := &node.NodeContext{
		Context:   context.Background(),
		Variables: make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
		OnCleanup: make([]func(), 0),
	}

	// Simulate 10 browser nodes running concurrently,
	// such as a workflow that opens many browser tabs at the same time.
	var wg sync.WaitGroup
	numNodes := 10

	for i := 0; i < numNodes; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine attempts to access Metadata.
			browserNode := &BrowserNode{
				Action: node.ActionFromMap(map[string]interface{}{
					"type": "navigate",
					"url":  "https://example.com",
				}),
			}

			// This triggers getOrCreateTab, which accesses ctx.Metadata.
			_, _ = browserNode.getOrCreateTab(ctx)
		}(i)
	}

	wg.Wait()
	t.Log("Test completed without panic")
}
