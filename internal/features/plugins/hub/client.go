package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	registryURL string
	httpClient  *http.Client
}

func NewClient(registryURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		registryURL: registryURL,
		httpClient:  httpClient,
	}
}

func (c *Client) Fetch(ctx context.Context) (RegistryResponse, error) {
	if c.registryURL == "" {
		return RegistryResponse{}, fmt.Errorf("plugin registry URL is not configured")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.registryURL, nil)
	if err != nil {
		return RegistryResponse{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return RegistryResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return RegistryResponse{}, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	var registry RegistryResponse
	if err := json.NewDecoder(resp.Body).Decode(&registry); err != nil {
		return RegistryResponse{}, err
	}

	if registry.SchemaVersion == "" {
		return RegistryResponse{}, fmt.Errorf("registry.schema_version is required")
	}
	if registry.SchemaVersion != "1.0" {
		return RegistryResponse{}, fmt.Errorf("unsupported registry schema_version: %s", registry.SchemaVersion)
	}
	if registry.Plugins == nil {
		registry.Plugins = []RegistryPluginSummary{}
	}
	return registry, nil
}
