package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ekken/internal/features/workflow/node"
)

func TestHTTPV2CurlExecutesRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("authorization = %q, want bearer token", got)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if payload["name"] != "Ekken" {
			t.Fatalf("payload name = %v, want Ekken", payload["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()

	n := &HTTPNode{Action: node.ActionFromMap(map[string]any{
		"action": "curl",
		"curl":   "curl -X POST " + server.URL + " -H 'Authorization: Bearer {{token}}' -H 'Content-Type: application/json' -d '{\"name\":\"{{name}}\"}'",
	})}

	result, err := n.Execute(&node.NodeContext{
		Stop:      make(chan struct{}),
		Context:   t.Context(),
		Variables: map[string]any{"token": "test-token", "name": "Ekken"},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Handle != "success" {
		t.Fatalf("handle = %q, want success", result.Handle)
	}

	response, ok := result.Response.(map[string]any)
	if !ok {
		t.Fatalf("response type = %T, want map[string]any", result.Response)
	}
	if response["status_code"] != http.StatusOK {
		t.Fatalf("status_code = %v, want 200", response["status_code"])
	}
	body, ok := response["body"].(map[string]any)
	if !ok {
		t.Fatalf("body type = %T, want map[string]any", response["body"])
	}
	if body["ok"] != true {
		t.Fatalf("body ok = %v, want true", body["ok"])
	}
}

func TestHTTPV2ParsesJSONBodyWithoutJSONContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(`{"data":{"choices":[{"message":"ok"}]}}`))
	}))
	defer server.Close()

	n := &HTTPNode{Action: node.ActionFromMap(map[string]any{
		"action": "curl",
		"curl":   "curl " + server.URL,
	})}

	result, err := n.Execute(&node.NodeContext{
		Stop:      make(chan struct{}),
		Context:   t.Context(),
		Variables: map[string]any{},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	response := result.Response.(map[string]any)
	body, ok := response["body"].(map[string]any)
	if !ok {
		t.Fatalf("body type = %T, want map[string]any", response["body"])
	}
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("data type = %T, want map[string]any", body["data"])
	}
	choices, ok := data["choices"].([]any)
	if !ok {
		t.Fatalf("choices type = %T, want []any", data["choices"])
	}
	if len(choices) != 1 {
		t.Fatalf("choices len = %d, want 1", len(choices))
	}
}

func TestHTTPV2KeepsEventStreamBodyRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"The\"},\"finish_reason\":null}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\" answer\"},\"finish_reason\":\"stop\"}]}\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	n := &HTTPNode{Action: node.ActionFromMap(map[string]any{
		"action": "curl",
		"curl":   "curl " + server.URL,
	})}

	result, err := n.Execute(&node.NodeContext{
		Stop:      make(chan struct{}),
		Context:   t.Context(),
		Variables: map[string]any{},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	response := result.Response.(map[string]any)
	if response["content_type"] != "text/event-stream; charset=utf-8" {
		t.Fatalf("content_type = %q, want event stream content type", response["content_type"])
	}
	body, ok := response["body"].(string)
	if !ok {
		t.Fatalf("body type = %T, want string", response["body"])
	}
	if !strings.Contains(body, "data: [DONE]") {
		t.Fatalf("body = %q, want raw SSE stream", body)
	}
}

func TestHTTPV2DefaultsRequestBodyContentTypeToJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	n := &HTTPNode{Action: node.ActionFromMap(map[string]any{
		"action": "curl",
		"curl":   `curl -d '{"name":"Ekken"}' ` + server.URL,
	})}

	_, err := n.Execute(&node.NodeContext{
		Stop:      make(chan struct{}),
		Context:   t.Context(),
		Variables: map[string]any{},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestParseCurlSupportsCompactOptions(t *testing.T) {
	req, err := parseCurl(`curl -XPUT -H'X-Test: yes' -dhello https://example.test`)
	if err != nil {
		t.Fatalf("parseCurl() error = %v", err)
	}
	if req.Method != http.MethodPut {
		t.Fatalf("method = %s, want PUT", req.Method)
	}
	if req.Headers.Get("X-Test") != "yes" {
		t.Fatalf("X-Test = %q, want yes", req.Headers.Get("X-Test"))
	}
	if req.Body != "hello" {
		t.Fatalf("body = %q, want hello", req.Body)
	}
	if req.URL != "https://example.test" {
		t.Fatalf("url = %q", req.URL)
	}
}

func TestParseCurlSupportsWorkflowVariablesAndCredentials(t *testing.T) {
	previousResolver := node.CredentialResolver
	node.CredentialResolver = func(key string) (string, error) {
		if key != "cred.api_token" {
			t.Fatalf("credential key = %q, want cred.api_token", key)
		}
		return "secret-token", nil
	}
	defer func() {
		node.CredentialResolver = previousResolver
	}()

	command := node.ParseTemplate(
		`curl -H 'Authorization: Bearer {{ cred.api_token }}' -H 'X-User: {{user}}' -d '{"page":{{page + 1}},"name":"{{user}}"}' https://example.test/users`,
		map[string]any{
			"user": "Ekken",
			"page": float64(1),
		},
	)

	req, err := parseCurl(command)
	if err != nil {
		t.Fatalf("parseCurl() error = %v", err)
	}
	if req.Headers.Get("Authorization") != "Bearer secret-token" {
		t.Fatalf("Authorization = %q, want credential value", req.Headers.Get("Authorization"))
	}
	if req.Headers.Get("X-User") != "Ekken" {
		t.Fatalf("X-User = %q, want Ekken", req.Headers.Get("X-User"))
	}
	if req.Body != `{"page":2,"name":"Ekken"}` {
		t.Fatalf("body = %q", req.Body)
	}
	if req.Method != http.MethodPost {
		t.Fatalf("method = %s, want POST", req.Method)
	}
}

func TestParseCurlSupportsPostmanStyleMultilineCommand(t *testing.T) {
	req, err := parseCurl(`curl --location 'https://example.test/users' \
--header 'Authorization: Bearer token' \
--header 'Content-Type: application/json' \
--data-raw '{"name":"Ekken"}'`)
	if err != nil {
		t.Fatalf("parseCurl() error = %v", err)
	}
	if req.URL != "https://example.test/users" {
		t.Fatalf("url = %q, want https://example.test/users", req.URL)
	}
	if req.Method != http.MethodPost {
		t.Fatalf("method = %s, want POST", req.Method)
	}
	if req.Headers.Get("Authorization") != "Bearer token" {
		t.Fatalf("Authorization = %q, want bearer token", req.Headers.Get("Authorization"))
	}
	if req.Headers.Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", req.Headers.Get("Content-Type"))
	}
	if req.Body != `{"name":"Ekken"}` {
		t.Fatalf("body = %q", req.Body)
	}
}

func TestParseCurlSupportsBrunoStyleURLFlag(t *testing.T) {
	req, err := parseCurl(`curl --request PATCH \
--url 'https://example.test/users/1' \
--header 'Content-Type: application/json' \
--data '{"active":true}'`)
	if err != nil {
		t.Fatalf("parseCurl() error = %v", err)
	}
	if req.URL != "https://example.test/users/1" {
		t.Fatalf("url = %q, want https://example.test/users/1", req.URL)
	}
	if req.Method != http.MethodPatch {
		t.Fatalf("method = %s, want PATCH", req.Method)
	}
	if req.Body != `{"active":true}` {
		t.Fatalf("body = %q", req.Body)
	}
}

func TestHTTPV2SpecHasGlobalFieldsAndResponseVar(t *testing.T) {
	spec, ok := node.GlobalRegistry.GetSpec("httpv2")
	if !ok {
		t.Fatal("httpv2 spec is not registered")
	}

	if len(spec.Actions) != 1 {
		t.Fatalf("actions len = %d, want 1", len(spec.Actions))
	}
	action := spec.Actions[0]
	if action.Key != "curl" {
		t.Fatalf("action key = %q, want curl", action.Key)
	}
	if action.ResponseVar == "" {
		t.Fatal("response var is empty")
	}
	if action.ResponseVar != "httpv2.curl_" {
		t.Fatalf("response var = %q, want httpv2.curl_", action.ResponseVar)
	}
	if len(action.Fields) != 1 || action.Fields[0].Key != "curl" {
		t.Fatalf("action fields = %#v, want only curl", action.Fields)
	}

	fields := map[string]node.NodeField{}
	for _, field := range spec.GlobalFields {
		fields[field.Key] = field
	}
	if fields["timeout"].Default != 60 {
		t.Fatalf("timeout default = %v, want 60", fields["timeout"].Default)
	}
	if fields["retry_count"].Default != 0 {
		t.Fatalf("retry_count default = %v, want 0", fields["retry_count"].Default)
	}
}
