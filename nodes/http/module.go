package http

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"ekken/internal/api/module"
	"ekken/internal/config"
	"ekken/internal/db"

	"github.com/gin-gonic/gin"
)

type httpModule struct{}

func init() {
	module.RegisterModule(&httpModule{})
}

func (m *httpModule) Name() string { return "http_node" }

func (m *httpModule) Init(_ *db.DB, _ config.Config) error { return nil }

func (m *httpModule) RegisterRoutes(api *gin.RouterGroup) {
	api.POST("/http/test", handleTestRequest)
}

// TestRequestPayload is the incoming payload from the UI for a test HTTP request.
type TestRequestPayload struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
	Timeout int               `json:"timeout"`
}

// handleTestRequest executes an HTTP request on behalf of the UI for live testing.
func handleTestRequest(c *gin.Context) {
	var payload TestRequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"error": "invalid payload: " + err.Error()})
		return
	}

	if payload.URL == "" {
		c.JSON(400, gin.H{"error": "url is required"})
		return
	}

	method := strings.ToUpper(payload.Method)
	if method == "" {
		method = "GET"
	}

	timeoutSec := payload.Timeout
	if timeoutSec <= 0 {
		timeoutSec = 60
	}

	var body io.Reader
	if payload.Body != "" && method != "GET" && method != "HEAD" {
		body = bytes.NewBufferString(payload.Body)
	}

	req, err := http.NewRequest(method, payload.URL, body)
	if err != nil {
		c.JSON(400, gin.H{"error": "failed to create request: " + err.Error()})
		return
	}

	// Set headers from payload
	hasContentType := false
	for k, v := range payload.Headers {
		req.Header.Set(k, v)
		if strings.ToLower(k) == "content-type" {
			hasContentType = true
		}
	}
	if !hasContentType && (method == "POST" || method == "PUT" || method == "PATCH") {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{
		Timeout: time.Duration(timeoutSec) * time.Second,
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(502, gin.H{"error": "request failed: " + err.Error()})
		return
	}
	defer resp.Body.Close()
	elapsed := time.Since(start).Milliseconds()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to read response: " + err.Error()})
		return
	}

	// Flatten response headers (take only first value per key)
	flatHeaders := make(map[string]string)
	for k, vals := range resp.Header {
		if len(vals) > 0 {
			flatHeaders[k] = vals[0]
		}
	}

	c.JSON(200, gin.H{
		"status":      resp.StatusCode,
		"status_text": resp.Status,
		"body":        string(respBody),
		"headers":     flatHeaders,
		"time_ms":     elapsed,
		"size":        len(respBody),
	})
}
