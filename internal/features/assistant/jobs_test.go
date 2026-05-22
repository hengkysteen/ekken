package assistant

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestJobManagerReplaysEventsToLateSubscriber(t *testing.T) {
	manager := NewJobManager()
	done := make(chan struct{})

	job, err := manager.Start("conv_test", func(ctx context.Context, sink StreamSink) error {
		if err := sink.Prepare("conv_test", "model", "provider"); err != nil {
			return err
		}
		if err := sink.Send(ChatResponse{
			ConversationID: "conv_test",
			Model:          "model",
			ProviderName:   "provider",
			Message:        MessageContent{Role: "assistant", Content: "hello"},
		}); err != nil {
			return err
		}
		if err := sink.Done("conv_test", "model"); err != nil {
			return err
		}
		close(done)
		return nil
	})
	if err != nil {
		t.Fatalf("start job: %v", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for job")
	}
	waitForJobStatus(t, manager, "conv_test", JobStatusDone)

	ch, unsubscribe := job.Subscribe()
	defer unsubscribe()

	var got strings.Builder
	for event := range ch {
		got.Write(event)
	}

	if !strings.Contains(got.String(), "hello") {
		t.Fatalf("expected replayed content, got %q", got.String())
	}
	if !strings.Contains(got.String(), "data: [DONE]") {
		t.Fatalf("expected replayed done marker, got %q", got.String())
	}
}

func TestJobManagerRejectsDuplicateRunningJob(t *testing.T) {
	manager := NewJobManager()
	release := make(chan struct{})

	_, err := manager.Start("conv_test", func(ctx context.Context, sink StreamSink) error {
		<-release
		return nil
	})
	if err != nil {
		t.Fatalf("start first job: %v", err)
	}

	_, err = manager.Start("conv_test", func(ctx context.Context, sink StreamSink) error {
		return nil
	})
	if !errors.Is(err, ErrJobAlreadyRunning) {
		t.Fatalf("expected ErrJobAlreadyRunning, got %v", err)
	}

	close(release)
}

func TestJobManagerStopCancelsRunningJob(t *testing.T) {
	manager := NewJobManager()
	cancelled := make(chan struct{})

	_, err := manager.Start("conv_test", func(ctx context.Context, sink StreamSink) error {
		<-ctx.Done()
		close(cancelled)
		return ctx.Err()
	})
	if err != nil {
		t.Fatalf("start job: %v", err)
	}

	if err := manager.Stop("conv_test"); err != nil {
		t.Fatalf("stop job: %v", err)
	}

	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for cancellation")
	}

	deadline := time.After(time.Second)
	for {
		snapshot, ok := manager.Snapshot("conv_test")
		if ok && snapshot.Status == JobStatusCanceled && !snapshot.Running {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("job did not become canceled, snapshot=%+v ok=%v", snapshot, ok)
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func waitForJobStatus(t *testing.T, manager *JobManager, convID string, status JobStatus) {
	t.Helper()
	deadline := time.After(time.Second)
	for {
		snapshot, ok := manager.Snapshot(convID)
		if ok && snapshot.Status == status {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("job did not reach status %q, snapshot=%+v ok=%v", status, snapshot, ok)
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
