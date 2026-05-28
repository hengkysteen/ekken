package adapter

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"ekken/internal/features/assistant"
	"ekken/internal/features/plugins/passistant"
)

func TestAdapter_RegisterAndUnregister(t *testing.T) {
	// Ensure passistant.GlobalRegistry is set via init()
	if passistant.GlobalRegistry == nil {
		t.Fatal("expected passistant.GlobalRegistry to be set, but it was nil")
	}

	// Set up temporary ModelManager to test dynamic registration in models.json
	tempDir, err := os.MkdirTemp("", "ekken-adapter-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mm, err := assistant.NewModelManager(tempDir)
	if err != nil {
		t.Fatalf("failed to create model manager: %v", err)
	}

	runnerSpec := passistant.RunnerSpec{
		Command: "echo",
	}

	providerSpec := passistant.ProviderSpec{
		ID:          "adapter-test-prov",
		Name:        "Adapter Test Provider",
		Icon:        "logo.png",
		OfficialURL: "https://example.com",
		Models: []passistant.ModelSpecEntry{
			{Name: "Model 1", Origin: "origin-1", ContextWindow: 2048},
		},
	}

	// Register using the adapter
	err = passistant.GlobalRegistry.Register("adapter-test-prov", runnerSpec, providerSpec, "/dummy/path", 1*time.Second)
	if err != nil {
		t.Fatalf("failed to register provider: %v", err)
	}

	// Verify that models.json is updated on disk and contains the registered plugin model
	modelsFilePath := filepath.Join(tempDir, "models.json")
	if _, err := os.Stat(modelsFilePath); os.IsNotExist(err) {
		t.Errorf("expected models.json to be created on disk, but it was not")
	} else {
		// Load models to ensure it contains our registered provider
		models := mm.GetModels("adapter-test-prov")
		if len(models) != 1 || models[0].Model != "origin-1" || models[0].Name != "Model 1" {
			t.Errorf("expected dynamic registration to add model to models.json: %+v", models)
		}
	}

	// Instantiate the provider in assistant registry
	err = assistant.CreateProvider("adapter-test-prov", nil, nil)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	// Verify that the provider is registered in the assistant registry
	p, err := assistant.GetProvider("adapter-test-prov")
	if err != nil {
		t.Fatalf("provider was not registered in assistant registry: %v", err)
	}

	info := p.Info()
	if info.ID != "adapter-test-prov" || info.Name != "Adapter Test Provider" || info.Logo != "logo.png" {
		t.Errorf("incorrect provider info registered: %+v", info)
	}

	// Verify that the models are registered
	models := assistant.GetDefaultModels("adapter-test-prov")
	if len(models) != 1 || models[0].Name != "Model 1" || models[0].Origin != "origin-1" || models[0].ContextWindow != 2048 {
		t.Errorf("incorrect models registered: %+v", models)
	}

	// Unregister
	err = passistant.GlobalRegistry.Unregister("adapter-test-prov")
	if err != nil {
		t.Fatalf("failed to unregister provider: %v", err)
	}

	// Verify that the provider is removed
	_, err = assistant.GetProvider("adapter-test-prov")
	if err == nil {
		t.Error("expected provider to be unregistered, but it was still found")
	}
}
