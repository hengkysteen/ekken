package webhook

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ekken/internal/api"
	"ekken/internal/features/workflow/node"

	"github.com/gin-gonic/gin"
)

const maxWebhookBodyBytes = 1 << 20

var listenerWaitTimeout = 3 * time.Second

type Handler struct{}

func (h *Handler) Handle(c *gin.Context) {
	webhookID := strings.TrimSpace(c.Param("id"))
	if webhookID == "" {
		c.JSON(http.StatusNotFound, api.Response{OK: false, Error: "webhook listener not active"})
		return
	}

	listener, ok := waitForListener(c.Request.Context(), webhookID, listenerWaitTimeout)
	if !ok {
		c.JSON(http.StatusNotFound, api.Response{OK: false, Error: "webhook listener not active"})
		return
	}

	if !methodAllowed(listener.Action, c.Request.Method) {
		c.JSON(http.StatusMethodNotAllowed, api.Response{OK: false, Error: "method not allowed"})
		return
	}
	if !secretValid(listener.Action, c) {
		c.JSON(http.StatusUnauthorized, api.Response{OK: false, Error: "invalid webhook secret"})
		return
	}

	payload, err := buildPayload(c, webhookID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}

	if err := node.GlobalEventListeners.Dispatch(webhookID, payload); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(strings.ToLower(err.Error()), "busy") {
			status = http.StatusConflict
		}
		c.JSON(status, api.Response{OK: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, api.Response{OK: true})
}

func waitForListener(ctx context.Context, webhookID string, timeout time.Duration) (*node.EventListener, bool) {
	if listener, ok := node.GlobalEventListeners.Get(webhookID); ok {
		return listener, true
	}
	if timeout <= 0 {
		return nil, false
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, false
		case <-timer.C:
			return nil, false
		case <-ticker.C:
			if listener, ok := node.GlobalEventListeners.Get(webhookID); ok {
				return listener, true
			}
		}
	}
}

func methodAllowed(action node.Action, method string) bool {
	allowed := methods(action)
	for _, m := range allowed {
		if strings.EqualFold(m, method) {
			return true
		}
	}
	return false
}

func methods(action node.Action) []string {
	raw := node.FieldValue(action, "methods")
	methods := make([]string, 0)
	switch v := raw.(type) {
	case []string:
		methods = append(methods, v...)
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok {
				methods = append(methods, s)
			}
		}
	case string:
		for _, part := range strings.Split(v, ",") {
			methods = append(methods, part)
		}
	}
	if len(methods) == 0 {
		methods = []string{http.MethodPost}
	}

	clean := methods[:0]
	for _, method := range methods {
		method = strings.ToUpper(strings.TrimSpace(method))
		if method != "" {
			clean = append(clean, method)
		}
	}
	if len(clean) == 0 {
		return []string{http.MethodPost}
	}
	return clean
}

func secretValid(action node.Action, c *gin.Context) bool {
	secret, _ := node.FieldValue(action, "secret").(string)
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return true
	}

	location, _ := node.FieldValue(action, "secret_location").(string)
	location = strings.ToLower(strings.TrimSpace(location))
	if location == "" {
		location = "header"
	}

	key, _ := node.FieldValue(action, "secret_key").(string)
	key = strings.TrimSpace(key)
	if key == "" {
		key = "x-ekken-secret"
	}

	var got string
	if location == "query" {
		got = c.Query(key)
	} else {
		got = c.GetHeader(key)
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(secret)) == 1
}

func buildPayload(c *gin.Context, webhookID string) (map[string]any, error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxWebhookBodyBytes)
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body")
	}

	contentType := c.GetHeader("Content-Type")
	payload := map[string]any{
		"webhook_id":   webhookID,
		"method":       c.Request.Method,
		"headers":      normalizedHeaders(c.Request.Header),
		"query":        c.Request.URL.Query(),
		"raw_body":     string(raw),
		"content_type": contentType,
		"remote_ip":    c.ClientIP(),
		"received_at":  time.Now().UTC().Format(time.RFC3339Nano),
	}

	if isJSONContent(contentType) && len(raw) > 0 {
		var parsed any
		if err := json.Unmarshal(raw, &parsed); err != nil {
			payload["body"] = nil
			payload["parse_error"] = err.Error()
		} else {
			payload["body"] = parsed
		}
	} else {
		payload["body"] = nil
	}

	return payload, nil
}

func normalizedHeaders(headers http.Header) map[string][]string {
	out := make(map[string][]string, len(headers))
	for key, values := range headers {
		copied := make([]string, len(values))
		copy(copied, values)
		out[strings.ToLower(key)] = copied
	}
	return out
}

func isJSONContent(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "application/json") || strings.Contains(contentType, "+json")
}
