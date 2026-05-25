package workflow

import (
	"testing"
	"time"
)

func TestWorkflowEventStream_Individual(t *testing.T) {
	sse := NewWorkflowEventStream()
	wfID := "wf-1"

	// Use case: multiple subscribers for one workflow.
	sub1ID, ch1 := sse.Subscribe(wfID)
	sub2ID, ch2 := sse.Subscribe(wfID)

	if sub1ID == sub2ID {
		t.Errorf("SubIDs should be unique, got both %s", sub1ID)
	}

	msg := SSEMessage{Type: "status", Data: "running"}
	sse.Send(wfID, msg)

	// Both subscribers should receive the same message.
	for i, ch := range []<-chan SSEMessage{ch1, ch2} {
		select {
		case received := <-ch:
			if received.Data != msg.Data {
				t.Errorf("Ch%d expected %v, got %v", i+1, msg.Data, received.Data)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Ch%d timeout waiting for message", i+1)
		}
	}

	// Use case: unsubscribe cleanup.
	sse.Unsubscribe(wfID, sub1ID)

	// Send another message.
	sse.Send(wfID, SSEMessage{Type: "status", Data: "updated"})

	// ch1 should not receive anything after unsubscribe.
	select {
	case msg := <-ch1:
		t.Errorf("Sub1 should not receive message after unsubscribe, got %v", msg)
	case <-time.After(50 * time.Millisecond):
		// Success: no message received.
	}

	// ch2 should still receive messages.
	select {
	case <-ch2:
		// Success.
	case <-time.After(100 * time.Millisecond):
		t.Error("Sub2 should still receive message")
	}
}

func TestWorkflowEventStream_Global(t *testing.T) {
	sse := NewWorkflowEventStream()

	// Use Case: Multiple Global Subscribers
	sub1ID, gch1 := sse.SubscribeGlobal()
	sub2ID, gch2 := sse.SubscribeGlobal()

	if sub1ID == sub2ID {
		t.Errorf("Global SubIDs should be unique, got both %s", sub1ID)
	}

	msg := SSEMessage{Type: "global_status", Data: "system_online"}
	sse.SendGlobal(msg)

	for i, ch := range []<-chan SSEMessage{gch1, gch2} {
		select {
		case received := <-ch:
			if received.Type != msg.Type {
				t.Errorf("Global Ch%d mismatch", i+1)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Global Ch%d timeout", i+1)
		}
	}

	// Unsubscribe Global
	sse.UnsubscribeGlobal(sub1ID)
	sse.SendGlobal(SSEMessage{Type: "ping", Data: "pong"})

	select {
	case m := <-gch1:
		t.Errorf("Global Sub1 should not receive message after unsubscribe, got %v", m)
	case <-time.After(50 * time.Millisecond):
		// OK
	}
}

func TestWorkflowEventStream_NonBlocking(t *testing.T) {
	sse := NewWorkflowEventStream()
	wfID := "wf-block-test"

	// Sub 1: fill this channel so it blocks.
	_, _ = sse.Subscribe(wfID)

	// Sub 2: normal channel.
	_, ch2 := sse.Subscribe(wfID)

	// Fill ch1 with enough messages to exercise the non-blocking send path.
	for i := 0; i < 64; i++ {
		sse.Send(wfID, SSEMessage{Type: "fill", Data: i})
	}

	// Drain ch2 so it is empty again, while leaving ch1 full.
	for i := 0; i < 64; i++ {
		<-ch2
	}

	// ch1 is full and ch2 is empty.
	testMsg := SSEMessage{Type: "important", Data: "must_reach_ch2"}

	// This must not hang because Send uses a default case in its select.
	done := make(chan bool)
	go func() {
		sse.Send(wfID, testMsg)
		done <- true
	}()

	select {
	case <-done:
		// Success: Send did not block.
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Send blocked when one channel was full!")
	}

	// ch2 should still receive the latest message even when ch1 is full.
	receivedAny := false
	// Drain ch2 to find our important message
	for i := 0; i < 66; i++ { // Check up to 66 to be safe
		select {
		case m := <-ch2:
			if m.Type == "important" {
				receivedAny = true
			}
		case <-time.After(10 * time.Millisecond):
		}
	}

	if !receivedAny {
		t.Error("Sub2 should have received the message even if Sub1 was full")
	}
}

func TestWorkflowEventStream_Finish(t *testing.T) {
	sse := NewWorkflowEventStream()
	id := "wf-finish"
	_, ch := sse.Subscribe(id)

	sse.Finish(id)

	// The channel should be closed.
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("Channel should be closed after Finish")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for channel close")
	}
}
