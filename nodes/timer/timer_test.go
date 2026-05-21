package timer

import (
	"ekken/internal/features/workflow/node"
	"testing"
	"time"
)

func makeCtx(iteration int) *node.NodeContext {
	stop := make(chan struct{})
	return &node.NodeContext{
		Stop:      stop,
		Iteration: iteration,
		Variables: map[string]interface{}{},
	}
}

func makeCtxWithStop(iteration int) (*node.NodeContext, chan struct{}) {
	stop := make(chan struct{})
	return &node.NodeContext{
		Stop:      stop,
		Iteration: iteration,
		Variables: map[string]interface{}{},
	}, stop
}

func executor(config map[string]any) *TimerNode {
	return &TimerNode{action: node.ActionFromMap(config)}
}

// --- Manual ---

func TestManual_FirstIteration(t *testing.T) {
	n := executor(map[string]any{"type": "manual"})
	res, err := n.Execute(makeCtx(0))
	if err != nil || res.Handle != "success" {
		t.Fatalf("got err=%v handle=%q, want nil/success", err, res.Handle)
	}
}

func TestManual_SecondIteration_Complete(t *testing.T) {
	n := executor(map[string]any{"type": "manual"})
	_, err := n.Execute(makeCtx(1))
	if err != node.ErrWorkflowComplete {
		t.Fatalf("got %v, want ErrWorkflowComplete", err)
	}
}

// --- Interval ---

func TestInterval_Success(t *testing.T) {
	n := executor(map[string]any{"type": "interval", "interval": float64(0), "count": float64(0)})
	res, err := n.Execute(makeCtx(0))
	if err != nil || res.Handle != "success" {
		t.Fatalf("got err=%v handle=%q", err, res.Handle)
	}
}

func TestInterval_CountReached(t *testing.T) {
	n := executor(map[string]any{"type": "interval", "interval": float64(1), "count": float64(2)})
	_, err := n.Execute(makeCtx(2))
	if err != node.ErrWorkflowComplete {
		t.Fatalf("got %v, want ErrWorkflowComplete", err)
	}
}

func TestInterval_Stop(t *testing.T) {
	ctx, stop := makeCtxWithStop(0)
	n := executor(map[string]any{"type": "interval", "interval": float64(60), "count": float64(0)})

	done := make(chan error, 1)
	go func() {
		_, err := n.Execute(ctx)
		done <- err
	}()

	time.Sleep(20 * time.Millisecond)
	close(stop)

	if err := <-done; err != node.ErrNodeStopped {
		t.Fatalf("got %v, want ErrNodeStopped", err)
	}
}

func TestInterval_NegativeCount_Error(t *testing.T) {
	n := executor(map[string]any{"type": "interval", "count": float64(-1)})
	_, err := n.Execute(makeCtx(0))
	if err == nil {
		t.Fatal("expected error for negative count")
	}
}

func TestInterval_NegativeInterval_Error(t *testing.T) {
	n := executor(map[string]any{"type": "interval", "interval": float64(-1), "count": float64(0)})
	_, err := n.Execute(makeCtx(0))
	if err == nil {
		t.Fatal("expected error for negative interval")
	}
}

func runAndWait(n *TimerNode, ctx *node.NodeContext) {
	done := make(chan struct{})
	go func() {
		n.Execute(ctx)
		close(done)
	}()
	<-done
}

func TestInterval_IsLooping_Unlimited(t *testing.T) {
	ctx := makeCtx(0)
	n := executor(map[string]any{"type": "interval", "interval": float64(0), "count": float64(0)})
	runAndWait(n, ctx)
	if !ctx.IsLooping {
		t.Fatal("expected IsLooping=true for unlimited count")
	}
}

func TestInterval_IsLooping_NotLastIteration(t *testing.T) {
	ctx := makeCtx(0)
	n := executor(map[string]any{"type": "interval", "interval": float64(0), "count": float64(3)})
	runAndWait(n, ctx)
	if !ctx.IsLooping {
		t.Fatal("expected IsLooping=true when not last iteration")
	}
}

func TestInterval_IsLooping_LastIteration(t *testing.T) {
	ctx := makeCtx(1)
	n := executor(map[string]any{"type": "interval", "interval": float64(0), "count": float64(2)})
	runAndWait(n, ctx)
	if ctx.IsLooping {
		t.Fatal("expected IsLooping=false on last iteration")
	}
}

// --- Cron ---

func TestCron_MissingExpression(t *testing.T) {
	n := executor(map[string]any{"type": "cron", "count": float64(0)})
	_, err := n.Execute(makeCtx(0))
	if err == nil {
		t.Fatal("expected error for missing cron expression")
	}
}

func TestCron_InvalidExpression(t *testing.T) {
	n := executor(map[string]any{"type": "cron", "cron": "not-a-cron", "count": float64(0)})
	_, err := n.Execute(makeCtx(0))
	if err == nil {
		t.Fatal("expected error for invalid cron expression")
	}
}

func TestCron_Stop(t *testing.T) {
	ctx, stop := makeCtxWithStop(0)
	// every minute — won't fire during test
	n := executor(map[string]any{"type": "cron", "cron": "0 * * * * *", "count": float64(0)})

	done := make(chan error, 1)
	go func() {
		_, err := n.Execute(ctx)
		done <- err
	}()

	time.Sleep(20 * time.Millisecond)
	close(stop)

	if err := <-done; err != node.ErrNodeStopped {
		t.Fatalf("got %v, want ErrNodeStopped", err)
	}
}

func TestCron_CountReached(t *testing.T) {
	n := executor(map[string]any{"type": "cron", "cron": "0 * * * * *", "count": float64(1)})
	_, err := n.Execute(makeCtx(1))
	if err != node.ErrWorkflowComplete {
		t.Fatalf("got %v, want ErrWorkflowComplete", err)
	}
}
