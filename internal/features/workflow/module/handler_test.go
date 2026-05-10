package module

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ekken/internal/features/workflow"

	"github.com/gin-gonic/gin"
)

type mockRuntime struct {
	workflow.RuntimeServicer
	status func(id string) workflow.WorkflowRunStatus
	running func() []workflow.WorkflowRunStatus
}

func (m *mockRuntime) Status(id string) workflow.WorkflowRunStatus {
	return m.status(id)
}

func (m *mockRuntime) Running() []workflow.WorkflowRunStatus {
	return m.running()
}

type mockSSE struct {
	workflow.SSEServicer
	subscribe func(id string) (string, <-chan workflow.SSEMessage)
	unsubscribe func(id, subID string)
	subscribeGlobal func() (string, <-chan workflow.SSEMessage)
	unsubscribeGlobal func(subID string)
}

func (m *mockSSE) Subscribe(id string) (string, <-chan workflow.SSEMessage) {
	return m.subscribe(id)
}
func (m *mockSSE) Unsubscribe(id, subID string) {
	m.unsubscribe(id, subID)
}
func (m *mockSSE) SubscribeGlobal() (string, <-chan workflow.SSEMessage) {
	return m.subscribeGlobal()
}
func (m *mockSSE) UnsubscribeGlobal(subID string) {
	m.unsubscribeGlobal(subID)
}

type CloseNotifyingRecorder struct {
	*httptest.ResponseRecorder
	closed chan bool
}

func (c *CloseNotifyingRecorder) CloseNotify() <-chan bool {
	return c.closed
}

func (c *CloseNotifyingRecorder) Flush() {}

func TestWorkflowHandler_SSEStream_InitialSync(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expectedStatus := workflow.WorkflowRunStatus{
		ID:     "test-wf",
		Status: "running",
	}

	mockR := &mockRuntime{
		status: func(id string) workflow.WorkflowRunStatus {
			return expectedStatus
		},
	}

	ch := make(chan workflow.SSEMessage)
	mockS := &mockSSE{
		subscribe: func(id string) (string, <-chan workflow.SSEMessage) {
			return "sub-1", ch
		},
		unsubscribe: func(id, subID string) {},
	}

	h := &WorkflowHandler{
		runtime: mockR,
		sse:     mockS,
	}

	w := &CloseNotifyingRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		closed:           make(chan bool),
	}
	c, _ := gin.CreateTestContext(w)
	
	// Create a context that we can cancel to stop the stream
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/api/workflow/test-wf/status/stream", nil).WithContext(ctx)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "test-wf"}}

	// Start the handler in a goroutine because it blocks on c.Stream
	go func() {
		h.SSEStream(c)
	}()

	// Wait a bit for initial sync to be written
	time.Sleep(100 * time.Millisecond)
	close(w.closed) // Trigger CloseNotify
	cancel()        // Stop the stream
	time.Sleep(50 * time.Millisecond)

	body := w.Body.String()
	if !strings.Contains(body, "event:status_update") {
		t.Errorf("Expected initial status_update event, but not found in body: %s", body)
	}

	expectedJSON, _ := json.Marshal(expectedStatus)
	if !strings.Contains(body, string(expectedJSON)) {
		t.Errorf("Expected status data %s, but not found in body: %s", string(expectedJSON), body)
	}
}

func TestWorkflowHandler_WorkflowsStatus_InitialSync(t *testing.T) {
	gin.SetMode(gin.TestMode)

	runningWfs := []workflow.WorkflowRunStatus{
		{ID: "wf-1", Status: "running"},
		{ID: "wf-2", Status: "running"},
	}

	mockR := &mockRuntime{
		running: func() []workflow.WorkflowRunStatus {
			return runningWfs
		},
	}

	ch := make(chan workflow.SSEMessage)
	mockS := &mockSSE{
		subscribeGlobal: func() (string, <-chan workflow.SSEMessage) {
			return "global-1", ch
		},
		unsubscribeGlobal: func(subID string) {},
	}

	h := &WorkflowHandler{
		runtime: mockR,
		sse:     mockS,
	}

	w := &CloseNotifyingRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		closed:           make(chan bool),
	}
	ctx, cancel := context.WithCancel(context.Background())
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/workflows/status", nil).WithContext(ctx)
	c.Request = req

	go func() {
		h.WorkflowsStatus(c)
	}()

	time.Sleep(100 * time.Millisecond)
	close(w.closed)
	cancel()
	time.Sleep(50 * time.Millisecond)

	body := w.Body.String()
	for _, wf := range runningWfs {
		data, _ := json.Marshal(wf)
		if !strings.Contains(body, string(data)) {
			t.Errorf("Expected running workflow data %s, but not found in body", string(data))
		}
	}
}
