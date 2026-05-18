package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileSystem interface {
	ReadFile(name string) ([]byte, error)
	Glob(pattern string) (matches []string, err error)
	MkdirAll(path string, perm os.FileMode) error
	Stat(name string) (os.FileInfo, error)
}

type osFS struct{}

func (osFS) ReadFile(name string) ([]byte, error)         { return os.ReadFile(name) }
func (osFS) Glob(pattern string) ([]string, error)        { return filepath.Glob(pattern) }
func (osFS) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }
func (osFS) Stat(name string) (os.FileInfo, error)        { return os.Stat(name) }

func normalizeVersion(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

func resolveCommand(baseDir, command string, args []string) (string, []string) {
	if strings.HasPrefix(command, ".") {
		return filepath.Join(baseDir, command), args
	}
	return command, args
}

func cloneMap(in map[string]interface{}) map[string]interface{} {
	if in == nil {
		return map[string]interface{}{}
	}
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// validateManifest validates common plugin manifest fields.
func validateManifest(manifest Manifest) error {
	if manifest.Kind == "" {
		return fmt.Errorf("plugin.kind is required")
	}

	if len(manifest.Spec) == 0 {
		return fmt.Errorf("plugin.spec is required")
	}

	return nil
}
