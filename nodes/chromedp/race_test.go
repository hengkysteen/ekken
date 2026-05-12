package chromedpnode

import (
	"context"
	"ekken/internal/features/workflow/node"
	"sync"
	"testing"
)

// TestConcurrentMetadataAccess mencoba membuktikan apakah ada data race
// saat multiple browser nodes mengakses ctx.Metadata secara bersamaan
func TestConcurrentMetadataAccess(t *testing.T) {
	// Setup shared NodeContext (seperti di workflow yang sama)
	ctx := &node.NodeContext{
		Context:   context.Background(),
		Variables: make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
		OnCleanup: make([]func(), 0),
	}

	// Simulasi 10 browser nodes berjalan concurrent
	// (misal: workflow dengan banyak tab browser yang dibuka bersamaan)
	var wg sync.WaitGroup
	numNodes := 10

	for i := 0; i < numNodes; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Setiap goroutine mencoba akses Metadata
			browserNode := &BrowserNode{
				Action: node.ActionFromMap(map[string]interface{}{
					"action": "navigate",
					"url":    "https://example.com",
				}),
			}

			// Ini akan trigger getOrCreateTab yang akses ctx.Metadata
			_, _ = browserNode.getOrCreateTab(ctx)
		}(i)
	}

	wg.Wait()
	t.Log("Test selesai tanpa panic")
}
