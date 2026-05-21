package plugins

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"ekken/internal/features/plugins/hub"
	"ekken/internal/features/plugins/kind"
)

type PluginList struct {
	ID          string   `json:"id"`
	Icon        string   `json:"icon,omitempty"`
	Manifest    Manifest `json:"manifest"`
	SourcePath  string   `json:"source_path"`
	Status      string   `json:"status"`
	Reason      string   `json:"reason,omitempty"`
	IsInstalled bool     `json:"is_installed"`
	IsEnabled   bool     `json:"is_enabled"`
}

type Manifest struct {
	Source PluginSource    `json:"source"`
	Kind   string          `json:"kind"`
	Spec   json.RawMessage `json:"spec,omitempty"`
}

type PluginService struct {
	manager      *Manager
	registryURL  string
	httpClient   *http.Client
	installMu    sync.RWMutex
	installTasks map[string]*installTaskState
}

type PluginSource string

const (
	SourceLocal PluginSource = "local"
	SourceHub   PluginSource = "hub"
)

type PluginsState struct {
	Plugins map[string]PluginStateEntry `json:"plugins"`
}

type PluginStateEntry struct {
	Enabled bool `json:"enabled"`
}

type runtimePlugin struct {
	manifest     Manifest
	sourcePath   string
	manifestPath string
	status       string
	reason       string
}

func (p *runtimePlugin) kindPlugin(id string) kind.Plugin {
	return kind.Plugin{
		ID:           id,
		Kind:         p.manifest.Kind,
		Spec:         p.manifest.Spec,
		SourcePath:   p.sourcePath,
		ManifestPath: p.manifestPath,
	}
}

type InstallStatus string

const (
	InstallQueued      InstallStatus = "queued"
	InstallDownloading InstallStatus = "downloading"
	InstallVerifying   InstallStatus = "verifying"
	InstallExtracting  InstallStatus = "extracting"
	InstallInstalling  InstallStatus = "installing"
	InstallCompleted   InstallStatus = "completed"
	InstallFailed      InstallStatus = "failed"
	InstallCanceled    InstallStatus = "canceled"
)

type InstallTask struct {
	PluginID      string        `json:"plugin_id"`
	Status        InstallStatus `json:"status"`
	Progress      float64       `json:"progress"`
	BytesReceived int64         `json:"bytes_received"`
	BytesTotal    int64         `json:"bytes_total,omitempty"`
	Error         string        `json:"error,omitempty"`
}

type installTaskState struct {
	task   InstallTask
	cancel func()
}

type PluginServicer interface {
	List() []PluginList
	Reload() error
	Manage(id string, action string) error
	Registry(ctx context.Context) (hub.RegistryResponse, error)
	Install(ctx context.Context, id string) (InstallTask, error)
	InstallStatus(id string) (InstallTask, bool)
	StopInstall(id string) (InstallTask, error)
	Uninstall(id string) error
}
