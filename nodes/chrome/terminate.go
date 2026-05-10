package chrome

import (
	"ekken/internal/features/workflow/node"
	"ekken/internal/logger"
	"fmt"
	"syscall"
	"time"
)

func (n *GoogleChromeNode) terminate(port int) (node.NodeExecutionResult, error) {
	if err := StopChrome(port); err != nil {
		logger.DevPrintf("[Chrome] Warning during Chrome termination: %v\n", err)
	}
	return node.NodeExecutionResult{Handle: "success"}, nil
}

func StopChrome(port int) error {
	proc, ok := activeProcs[port]
	if !ok {
		if port == configPort && GlobalCancel != nil {
			GlobalCancel()
			GlobalAllocCtx = nil
			return nil
		}
		return fmt.Errorf("no known chrome process running on port %d", port)
	}
	if err := proc.Signal(syscall.SIGTERM); err == nil {
		done := make(chan error, 1)
		go func() {
			_, err := proc.Wait()
			done <- err
		}()
		select {
		case <-done:
			delete(activeProcs, port)
			if port == configPort {
				if GlobalCancel != nil {
					GlobalCancel()
				}
				GlobalAllocCtx = nil
				GlobalCancel = nil
			}
			return nil
		case <-time.After(3 * time.Second):
		}
	}
	if err := proc.Kill(); err != nil {
		return fmt.Errorf("failed to kill chrome on port %d: %v", port, err)
	}
	proc.Wait()
	delete(activeProcs, port)
	if port == configPort {
		if GlobalCancel != nil {
			GlobalCancel()
		}
		GlobalAllocCtx = nil
		GlobalCancel = nil
	}
	fmt.Printf("[Browser] Successfully stopped chrome on port %d\n", port)
	return nil
}
