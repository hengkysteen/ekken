package hub

import "encoding/json"

type RegistryResponse struct {
	SchemaVersion string                  `json:"schema_version"`
	Message       string                  `json:"message,omitempty"`
	Plugins       []RegistryPluginSummary `json:"plugins"`
}

type RegistryPluginSummary struct {
	ID           string             `json:"id"`
	Source       string             `json:"source"`
	Name         string             `json:"name"`
	Kind         string             `json:"kind"`
	Version      string             `json:"version"`
	Description  string             `json:"description,omitempty"`
	KindMeta     json.RawMessage    `json:"kind_meta,omitempty"`
	Repo         RegistryRepo       `json:"repo"`
	Artifacts    []RegistryArtifact `json:"artifacts"`
	IsInstalled  bool               `json:"is_installed"`
	IsEnabled    bool               `json:"is_enabled"`
	Status       string             `json:"status,omitempty"`
	LocalVersion string             `json:"local_version,omitempty"`
}

type RegistryRepo struct {
	URL    string `json:"url"`
	About  string `json:"about,omitempty"`
	Author string `json:"author,omitempty"`
}

type RegistryArtifact struct {
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	DownloadURL string `json:"download_url"`
	Checksum    string `json:"checksum"`
}
