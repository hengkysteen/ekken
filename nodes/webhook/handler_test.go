package webhook

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ekken/internal/features/workflow/node"

	"github.com/gin-gonic/gin"
)

func TestWebhookHandlerDispatchesPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	action := node.Action{
		Type: "on_request",
		Fields: []node.NodeField{
			{Key: "webhook_id", Value: "hook-1"},
			{Key: "methods", Value: []any{"POST"}},
			{Key: "secret", Value: "dev-secret"},
			{Key: "secret_location", Value: "header"},
			{Key: "secret_key", Value: "x-ekken-secret"},
		},
	}
	listener, cleanup, err := node.GlobalEventListeners.Register("hook-1", action, "wf-1")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	defer cleanup()

	h := &Handler{}
	router := gin.New()
	router.Any("/api/webhook/:id", h.Handle)

	req := httptest.NewRequest(http.MethodPost, "/api/webhook/hook-1", bytes.NewBufferString(`{"event":"ping"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-ekken-secret", "dev-secret")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	payload, ok := (<-listener.Payload).(map[string]any)
	if !ok {
		t.Fatalf("payload type = %T, want map[string]any", payload)
	}
	body, ok := payload["body"].(map[string]any)
	if !ok {
		t.Fatalf("payload body = %#v, want JSON object", payload["body"])
	}
	if body["event"] != "ping" {
		t.Fatalf("body.event = %#v, want ping", body["event"])
	}
}

func TestWebhookHandlerRejectsMethodAndSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)

	action := node.Action{
		Type: "on_request",
		Fields: []node.NodeField{
			{Key: "webhook_id", Value: "hook-2"},
			{Key: "methods", Value: []any{"POST"}},
			{Key: "secret", Value: "dev-secret"},
		},
	}
	_, cleanup, err := node.GlobalEventListeners.Register("hook-2", action, "wf-1")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	defer cleanup()

	h := &Handler{}
	router := gin.New()
	router.Any("/api/webhook/:id", h.Handle)

	req := httptest.NewRequest(http.MethodGet, "/api/webhook/hook-2", nil)
	req.Header.Set("x-ekken-secret", "dev-secret")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET status = %d, want 405", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/webhook/hook-2", nil)
	req.Header.Set("x-ekken-secret", "wrong")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("bad secret status = %d, want 401", rec.Code)
	}
}

func TestWebhookHandlerRequiresActiveListener(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousTimeout := listenerWaitTimeout
	listenerWaitTimeout = time.Millisecond
	defer func() {
		listenerWaitTimeout = previousTimeout
	}()

	h := &Handler{}
	router := gin.New()
	router.Any("/api/webhook/:id", h.Handle)

	req := httptest.NewRequest(http.MethodPost, "/api/webhook/not-running", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
