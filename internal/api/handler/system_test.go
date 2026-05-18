package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ekken/internal/config"

	"github.com/gin-gonic/gin"
)

func TestGetSystemConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &Handler{
		Config: config.Config{
			AppName:    "Ekken Test",
			DataDir:    "/tmp/ekken-data",
			PluginDir:  "/tmp/ekken-data/plugins",
			Address:    "localhost:9090",
			AppVersion: "1.2.3",
			Mode:       "development",
			RepoURL:    "https://example.com/ekken",
			Author:     "tester",
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/system/config", nil)
	c.Request = req

	h.GetSystemConfig(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp struct {
		OK   bool           `json:"ok"`
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("expected ok=true")
	}

	assertEqual := func(key string, want interface{}) {
		t.Helper()
		if got := resp.Data[key]; got != want {
			t.Fatalf("expected %s=%v, got %v", key, want, got)
		}
	}

	assertEqual("app_name", "Ekken Test")
	assertEqual("data_dir", "/tmp/ekken-data")
	assertEqual("plugin_dir", "/tmp/ekken-data/plugins")
	assertEqual("address", "localhost:9090")
	assertEqual("app_version", "1.2.3")
	assertEqual("mode", "development")
	assertEqual("repo_url", "https://example.com/ekken")
	assertEqual("author", "tester")

}
