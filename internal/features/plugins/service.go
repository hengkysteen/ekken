package plugins

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"ekken/internal/features/plugins/hub"
	"ekken/internal/features/plugins/kind"
)

const defaultPluginRegistryURL = "https://raw.githubusercontent.com/hengkysteen/ekken-plugins/refs/heads/main/hub/catalog.json"

func NewPluginService(appVersion, pluginDir string) (*PluginService, error) {
	manager := NewManager(appVersion, pluginDir, 60*time.Second)
	if err := manager.Load(); err != nil {
		return nil, err
	}
	return &PluginService{
		manager:      manager,
		registryURL:  defaultPluginRegistryURL,
		httpClient:   &http.Client{Timeout: 5 * time.Minute},
		installTasks: make(map[string]*installTaskState),
	}, nil
}

func (s *PluginService) List() []PluginList {
	return s.manager.List()
}

func (s *PluginService) Reload() error {
	return s.manager.Load()
}

func (s *PluginService) Manage(id string, action string) error {
	return s.manager.Manage(id, action)
}

func (s *PluginService) Registry(ctx context.Context) (hub.RegistryResponse, error) {
	registry, err := hub.NewClient(s.registryURL, s.httpClient).Fetch(ctx)
	if err != nil {
		return hub.RegistryResponse{}, err
	}
	s.mergeLocalPluginStatus(&registry)
	return registry, nil
}

func (s *PluginService) Install(ctx context.Context, id string) (InstallTask, error) {
	s.installMu.Lock()
	if running, ok := s.installTasks[id]; ok && isInstallRunning(running.task.Status) {
		task := running.task
		s.installMu.Unlock()
		return task, nil
	}
	s.installMu.Unlock()

	registry, err := hub.NewClient(s.registryURL, s.httpClient).Fetch(ctx)
	if err != nil {
		return InstallTask{}, err
	}

	plugin, ok := findRegistryPlugin(registry.Plugins, id)
	if !ok {
		return InstallTask{}, fmt.Errorf("registry plugin not found: %s", id)
	}
	if s.isInstalled(id) {
		return InstallTask{}, fmt.Errorf("plugin already installed: %s", id)
	}

	artifact, ok := selectArtifact(plugin.Artifacts)
	if !ok {
		return InstallTask{}, fmt.Errorf("plugin %s has no artifact for %s/%s", id, runtime.GOOS, runtime.GOARCH)
	}

	installCtx, cancel := context.WithCancel(context.Background())
	task := InstallTask{PluginID: id, Status: InstallQueued}
	s.installMu.Lock()
	s.installTasks[id] = &installTaskState{task: task, cancel: cancel}
	s.installMu.Unlock()

	go s.runInstall(installCtx, id, plugin, artifact)
	return task, nil
}

func (s *PluginService) InstallStatus(id string) (InstallTask, bool) {
	s.installMu.RLock()
	defer s.installMu.RUnlock()
	task, ok := s.installTasks[id]
	if !ok {
		return InstallTask{}, false
	}
	return task.task, true
}

func (s *PluginService) StopInstall(id string) (InstallTask, error) {
	s.installMu.RLock()
	task, ok := s.installTasks[id]
	s.installMu.RUnlock()
	if !ok {
		return InstallTask{}, fmt.Errorf("install task not found: %s", id)
	}
	if !isInstallRunning(task.task.Status) {
		return task.task, nil
	}
	task.cancel()
	s.updateInstallTask(id, InstallTask{PluginID: id, Status: InstallCanceled, Error: "install canceled"})
	return s.InstallStatusOrZero(id), nil
}

func (s *PluginService) Uninstall(id string) error {
	return s.manager.Uninstall(id)
}

func (s *PluginService) InstallStatusOrZero(id string) InstallTask {
	task, _ := s.InstallStatus(id)
	return task
}

func (s *PluginService) mergeLocalPluginStatus(registry *hub.RegistryResponse) {
	local := s.manager.List()
	for i := range registry.Plugins {
		for _, summary := range local {
			if summary.ID == registry.Plugins[i].ID || filepath.Base(summary.SourcePath) == registry.Plugins[i].ID {
				registry.Plugins[i].IsInstalled = summary.IsInstalled
				registry.Plugins[i].IsEnabled = summary.IsEnabled
				registry.Plugins[i].Status = summary.Status
				break
			}
		}
	}
}

func (s *PluginService) isInstalled(id string) bool {
	for _, summary := range s.manager.List() {
		if summary.ID == id || filepath.Base(summary.SourcePath) == id {
			return true
		}
	}
	return false
}

func (s *PluginService) runInstall(ctx context.Context, id string, plugin hub.RegistryPluginSummary, artifact hub.RegistryArtifact) {
	if err := s.installPlugin(ctx, id, plugin, artifact); err != nil {
		status := InstallFailed
		if ctx.Err() != nil {
			status = InstallCanceled
		}
		s.updateInstallTask(id, InstallTask{PluginID: id, Status: status, Error: err.Error()})
		return
	}
	s.updateInstallTask(id, InstallTask{PluginID: id, Status: InstallCompleted, Progress: 1})
}

func (s *PluginService) installPlugin(ctx context.Context, id string, plugin hub.RegistryPluginSummary, artifact hub.RegistryArtifact) error {
	if err := validateRegistryID(id); err != nil {
		return err
	}

	tempRoot, err := os.MkdirTemp("", "ekken-plugin-install-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempRoot)

	archivePath := filepath.Join(tempRoot, "artifact.zip")
	s.updateInstallTask(id, InstallTask{PluginID: id, Status: InstallDownloading})
	if err := s.downloadArchive(ctx, id, artifact.DownloadURL, archivePath); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	s.updateInstallTask(id, InstallTask{PluginID: id, Status: InstallVerifying})
	if err := verifyChecksum(archivePath, artifact.Checksum); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	extractDir := filepath.Join(tempRoot, "extract")
	s.updateInstallTask(id, InstallTask{PluginID: id, Status: InstallExtracting})
	if err := extractZip(archivePath, extractDir); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	manifestPath := filepath.Join(extractDir, "plugin.json")
	rawManifest, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("missing plugin.json: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(rawManifest, &manifest); err != nil {
		return fmt.Errorf("invalid plugin.json: %w", err)
	}
	if manifest.Source == "" {
		manifest.Source = SourceHub
	}
	if err := validateManifest(manifest); err != nil {
		return err
	}
	if plugin.Kind != "" && manifest.Kind != plugin.Kind {
		return fmt.Errorf("artifact kind %q does not match registry kind %q", manifest.Kind, plugin.Kind)
	}

	handler := kind.GetHandler(manifest.Kind, kind.Config{
		ExecTimeout: s.manager.execTimeout,
	})
	if handler == nil {
		return fmt.Errorf("unsupported plugin kind: %s", manifest.Kind)
	}

	runtimePlugin := &runtimePlugin{
		manifest:     manifest,
		sourcePath:   extractDir,
		manifestPath: manifestPath,
	}
	if err := handler.Validate(runtimePlugin.kindPlugin(id)); err != nil {
		return err
	}

	finalDir := filepath.Join(s.manager.pluginDir, manifest.Kind+"s", id)
	s.updateInstallTask(id, InstallTask{PluginID: id, Status: InstallInstalling})
	if err := os.MkdirAll(filepath.Dir(finalDir), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(finalDir); err == nil {
		return fmt.Errorf("final plugin directory already exists: %s", finalDir)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(extractDir, finalDir); err != nil {
		return err
	}
	return s.manager.Load()
}

func (s *PluginService) downloadArchive(ctx context.Context, id, downloadURL, archivePath string) error {
	if downloadURL == "" {
		return fmt.Errorf("artifact.download_url is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	out, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer out.Close()

	progress := &installProgressWriter{
		id:    id,
		total: resp.ContentLength,
		svc:   s,
	}
	if _, err := io.Copy(out, io.TeeReader(resp.Body, progress)); err != nil {
		return err
	}
	return nil
}

type installProgressWriter struct {
	id       string
	total    int64
	received int64
	svc      *PluginService
}

func (w *installProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	w.received += int64(n)
	progress := 0.0
	if w.total > 0 {
		progress = float64(w.received) / float64(w.total)
	}
	w.svc.updateInstallTask(w.id, InstallTask{
		PluginID:      w.id,
		Status:        InstallDownloading,
		Progress:      progress,
		BytesReceived: w.received,
		BytesTotal:    w.total,
	})
	return n, nil
}

func (s *PluginService) updateInstallTask(id string, task InstallTask) {
	s.installMu.Lock()
	defer s.installMu.Unlock()
	current, ok := s.installTasks[id]
	if !ok {
		s.installTasks[id] = &installTaskState{task: task}
		return
	}
	if task.PluginID == "" {
		task.PluginID = id
	}
	current.task = task
}

func findRegistryPlugin(plugins []hub.RegistryPluginSummary, id string) (hub.RegistryPluginSummary, bool) {
	for _, plugin := range plugins {
		if plugin.ID == id {
			return plugin, true
		}
	}
	return hub.RegistryPluginSummary{}, false
}

func selectArtifact(artifacts []hub.RegistryArtifact) (hub.RegistryArtifact, bool) {
	for _, artifact := range artifacts {
		if artifact.OS == runtime.GOOS && artifact.Arch == runtime.GOARCH {
			return artifact, true
		}
	}
	for _, artifact := range artifacts {
		if (artifact.OS == "any" || artifact.OS == runtime.GOOS) && (artifact.Arch == "any" || artifact.Arch == runtime.GOARCH) {
			return artifact, true
		}
	}
	return hub.RegistryArtifact{}, false
}

func isInstallRunning(status InstallStatus) bool {
	switch status {
	case InstallQueued, InstallDownloading, InstallVerifying, InstallExtracting, InstallInstalling:
		return true
	default:
		return false
	}
}

func validateRegistryID(id string) error {
	if id == "" {
		return fmt.Errorf("plugin id is required")
	}
	if id != filepath.Base(id) || strings.Contains(id, "..") {
		return fmt.Errorf("invalid plugin id: %s", id)
	}
	return nil
}

func verifyChecksum(path string, checksum string) error {
	algorithm, expected, ok := strings.Cut(checksum, ":")
	if !ok || algorithm == "" || expected == "" {
		return fmt.Errorf("checksum must use format sha256:<hex>")
	}
	if algorithm != "sha256" {
		return fmt.Errorf("unsupported checksum algorithm: %s", algorithm)
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}
	actual := hex.EncodeToString(hasher.Sum(nil))
	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("checksum mismatch")
	}
	return nil
}

func extractZip(archivePath, destDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	cleanDest, err := filepath.Abs(destDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cleanDest, 0o755); err != nil {
		return err
	}

	for _, file := range reader.File {
		target := filepath.Join(cleanDest, file.Name)
		cleanTarget, err := filepath.Abs(target)
		if err != nil {
			return err
		}
		if cleanTarget != cleanDest && !strings.HasPrefix(cleanTarget, cleanDest+string(os.PathSeparator)) {
			return fmt.Errorf("zip path traversal: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(cleanTarget, file.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(cleanTarget), 0o755); err != nil {
			return err
		}

		src, err := file.Open()
		if err != nil {
			return err
		}
		dst, err := os.OpenFile(cleanTarget, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.Mode())
		if err != nil {
			src.Close()
			return err
		}
		_, copyErr := io.Copy(dst, src)
		closeErr := dst.Close()
		src.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}
