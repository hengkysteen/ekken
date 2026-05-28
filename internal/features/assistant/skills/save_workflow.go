package skills

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type SaveWorkflow struct{}

func (s *SaveWorkflow) GetID() string   { return "save_workflow" }
func (s *SaveWorkflow) GetName() string { return "Save Workflow" }
func (s *SaveWorkflow) GetDescription() string {
	return "save temp to user db so user can run it in ekken workflow editor"
}

func (s *SaveWorkflow) Execute(args map[string]interface{}) (string, error) {
	tempID, ok := args["id"].(string)
	if !ok || tempID == "" {
		return "", fmt.Errorf("missing required parameter 'id'")
	}

	// 1. Get system config to locate data_dir
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(GetValidationAPIURL("/api/system/config"))
	if err != nil {
		return "", fmt.Errorf("failed to fetch system config: %w", err)
	}
	defer resp.Body.Close()

	var configResp struct {
		Data struct {
			DataDir string `json:"data_dir"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&configResp); err != nil {
		return "", fmt.Errorf("failed to decode system config: %w", err)
	}

	dataDir := configResp.Data.DataDir
	if dataDir == "" {
		return "", fmt.Errorf("data_dir not found in system config")
	}

	// 2. Read the temp json workflow file
	tempDir := filepath.Join(dataDir, "temp", "workflows")
	jsonPath := filepath.Join(tempDir, fmt.Sprintf("%s.json", tempID))
	jsonBytes, err := os.ReadFile(jsonPath)
	if err != nil {
		return "", fmt.Errorf("failed to read temporary JSON workflow for ID '%s': %w", tempID, err)
	}

	// 3. POST the workflow to /api/workflows
	postResp, err := client.Post(GetValidationAPIURL("/api/workflows"), "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("failed to call save workflow API: %w", err)
	}
	defer postResp.Body.Close()

	bodyBytes, err := io.ReadAll(postResp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
		Data  struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("failed to decode save response: %w. Body: %s", err, string(bodyBytes))
	}

	if !result.OK {
		return "", fmt.Errorf("failed to save workflow: %s", result.Error)
	}

	if err := removeTempWorkflowFiles(tempDir, tempID); err != nil {
		return "", err
	}

	return fmt.Sprintf("Workflow '%s' saved successfully. inform the user and stop.", result.Data.Name), nil
}

func removeTempWorkflowFiles(tempDir, tempID string) error {
	for _, ext := range []string{".json", ".yaml"} {
		path := filepath.Join(tempDir, fmt.Sprintf("%s%s", tempID, ext))
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove temporary workflow file %s: %w", path, err)
		}
	}
	return nil
}

func init() {
	Register(&SaveWorkflow{})
}
