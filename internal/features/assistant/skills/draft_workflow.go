package skills

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"ekken/internal/features/workflow"

	"github.com/goccy/go-yaml"
)

type DraftWorkflow struct{}

func (s *DraftWorkflow) GetID() string   { return "draft_workflow" }
func (s *DraftWorkflow) GetName() string { return "Draft Workflow" }
func (s *DraftWorkflow) GetDescription() string {
	return "Create a temporary workflow draft from model description"
}

func (s *DraftWorkflow) Execute(args map[string]any) (string, error) {
	// 1. Ambil data dari model (args sudah diparse oleh detector sebagai map)

	// 2. buat 2 variabel utk simpan data
	// wfYaml (data yaml asli dari model)
	wfYaml, err := yaml.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("failed to marshal yaml: %w", err)
	}

	// 5. simpan wfYaml ke temp/workflows/TEMP_ID
	tempID, _ := args["id"].(string)
	if tempID == "" {
		tempID = generateTempID()
	}

	// wfJson (data hasil convert yaml ke json = masih kosong)
	var wfJson workflow.Workflow

	// 3. panggil catalog cari action terkait dari wfYaml utk conversi menjadi wfJson
	if err := ConvertToInternal(args, &wfJson); err != nil {
		return "", err
	}

	// 4. validasi wfJson dgn api
	if err := ValidateWorkflow(wfJson); err != nil {
		return "", err
	}

	if err := s.saveTempWorkflow(tempID, string(wfYaml), wfJson); err != nil {
		return "", err
	}

	return fmt.Sprintf("Workflow draft created with ID: %s. Ask the user if they want to make any changes or save the workflow so they can run it in the editor.", tempID), nil
}

func (s *DraftWorkflow) saveTempWorkflow(id string, yamlContent string, jsonData any) error {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(GetValidationAPIURL("/api/system/config"))
	if err != nil {
		return fmt.Errorf("failed to fetch system config: %w", err)
	}
	defer resp.Body.Close()

	var configResp struct {
		Data struct {
			DataDir string `json:"data_dir"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&configResp); err != nil {
		return fmt.Errorf("failed to decode system config: %w", err)
	}

	dataDir := configResp.Data.DataDir
	if dataDir == "" {
		return fmt.Errorf("data_dir not found in system config")
	}

	dir := filepath.Join(dataDir, "temp", "workflows")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Save YAML file
	yamlPath := filepath.Join(dir, fmt.Sprintf("%s.yaml", id))
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		return fmt.Errorf("failed to save yaml workflow: %w", err)
	}

	// Save JSON file
	jsonPath := filepath.Join(dir, fmt.Sprintf("%s.json", id))
	jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal json for saving: %w", err)
	}
	if err := os.WriteFile(jsonPath, jsonBytes, 0644); err != nil {
		return fmt.Errorf("failed to save json workflow: %w", err)
	}

	return nil
}

func generateTempID() string {
	return fmt.Sprintf("tmp_%d", time.Now().UnixNano())
}

func init() {
	Register(&DraftWorkflow{})
}
