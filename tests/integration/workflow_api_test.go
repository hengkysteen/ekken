package integration

import (
	"bytes"
	"ekken/internal/api"
	"ekken/internal/config"
	"ekken/internal/db"
	_ "ekken/internal/features"
	_ "ekken/nodes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func setupTestEnvironment(t *testing.T) (*api.Server, *db.DB, func()) {
	tempDir, err := os.MkdirTemp("", "ekken_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	cfg := config.Config{
		DataDir: tempDir,
		Mode:    "test",
		Port:    0,
	}
	database, err := db.Open(cfg.DataDir)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	server := api.NewServer(cfg, database)
	cleanup := func() {
		database.Close()
		os.RemoveAll(tempDir)
	}
	return server, database, cleanup
}
func TestIntegration_CreateWorkflowAPI(t *testing.T) {
	server, _, cleanup := setupTestEnvironment(t)
	defer cleanup()
	router := server.Engine()
	payload := []byte(`{
		"name": "Test Workflow API",
		"nodes": []
	}`)
	req, _ := http.NewRequest("POST", "/api/workflows", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}
	if response["ok"] != true {
		t.Errorf("Expected response ok to be true, got %v", response["ok"])
	}
	reqGet, _ := http.NewRequest("GET", "/api/workflows", nil)
	wGet := httptest.NewRecorder()
	router.ServeHTTP(wGet, reqGet)
	if wGet.Code != http.StatusOK {
		t.Errorf("Expected GET status 200, got %d", wGet.Code)
	}
}
