package chrome

import (
	"context"
	"ekken/internal/features/workflow/node"
	"ekken/internal/logger"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/chromedp/chromedp"
)

func (n *GoogleChromeNode) launch(ctx *node.NodeContext, port int, resolvedProfile string) (node.NodeExecutionResult, error) {
	headless, _ := n.Config["headless"].(bool)
	binPath, _ := n.Config["bin_path"].(string)
	widthF, _ := n.Config["width"].(float64)
	heightF, _ := n.Config["height"].(float64)
	SetConfig(binPath, port, resolvedProfile, headless, int(widthF), int(heightF))
	if err := EnsureBrowser(ctx.Context, true); err != nil {
		return node.NodeExecutionResult{}, fmt.Errorf("failed to launch chrome: %v", err)
	}
	return node.NodeExecutionResult{Handle: "success"}, nil
}

func InitGlobalBrowser(chromeURL string) {
	if chromeURL == "" {
		chromeURL = "http://127.0.0.1:9222"
	}
	GlobalAllocCtx, GlobalCancel = chromedp.NewRemoteAllocator(context.Background(), chromeURL)
	logger.DevPrintf("[Browser] Connected to global Chrome at %s\n", chromeURL)
}

func isChromeReady(port int) bool {
	url := fmt.Sprintf("http://127.0.0.1:%d/json/version", port)
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}
	_, ok := result["webSocketDebuggerUrl"]
	return ok
}

func isPortOpen(port int) bool {
	ln, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 1*time.Second)
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

func getChromePath() string {
	switch runtime.GOOS {
	case "darwin":
		return "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	case "windows":
		paths := []string{
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	case "linux":
		return "google-chrome"
	}
	return ""
}

func EnsureBrowser(ctx context.Context, shouldLaunch bool) error {
	mu.Lock()
	defer mu.Unlock()
	if GlobalAllocCtx != nil && isPortOpen(configPort) && isChromeReady(configPort) {
		return nil
	}
	if !shouldLaunch {
		return fmt.Errorf("browser not running on port %d and shouldLaunch is false", configPort)
	}
	err := startChrome(ctx, configBin, configPort, configProfile, configHeadless, configWidth, configHeight)
	if err != nil {
		return err
	}
	chromeURL := fmt.Sprintf("http://127.0.0.1:%d", configPort)
	InitGlobalBrowser(chromeURL)
	return nil
}

func SetConfig(binPath string, port int, profileDir string, headless bool, width, height int) {
	configBin = binPath
	if port > 0 {
		configPort = port
	}
	configProfile = profileDir
	configHeadless = headless
	if width > 0 {
		configWidth = width
	}
	if height > 0 {
		configHeight = height
	}
}

func startChrome(ctx context.Context, binPath string, port int, profileDir string, headless bool, width, height int) error {
	if isPortOpen(port) {
		fmt.Printf("[Browser] Chrome port %d is already open, verifying if it is a valid debugging instance...\n", port)
		if isChromeReady(port) {
			fmt.Printf("[Browser] Chrome debugging is ready on port %d\n", port)
			return nil
		}
		return fmt.Errorf("port %d is already in use by another application or an invalid Chrome instance", port)
	}
	if binPath == "" {
		binPath = getChromePath()
	}
	if binPath == "" {
		return fmt.Errorf("chrome executable not found. Please set EKKENCHROME_BIN")
	}
	if width <= 0 {
		width = 1920
	}
	if height <= 0 {
		height = 1080
	}
	args := []string{
		fmt.Sprintf("--remote-debugging-port=%d", port),
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-background-timer-throttling",
		"--disable-backgrounding-occluded-windows",
		"--disable-renderer-backgrounding",
		"--disable-ipc-flooding-protection",
		"--disable-features=CalculateNativeWinOcclusion",
		fmt.Sprintf("--window-size=%d,%d", width, height),
	}
	if headless {
		args = append(args, "--headless=new")
	}
	if profileDir == "" {
		profileDir = filepath.Join(os.TempDir(), "ekken-chrome")
	}
	var absPath string
	if filepath.IsAbs(profileDir) {
		absPath = profileDir
	} else {
		homeDir, _ := os.UserHomeDir()
		absPath = filepath.Join(homeDir, ".ekken", "chromium/profiles", profileDir)
	}
	args = append(args, fmt.Sprintf("--user-data-dir=%s", absPath))
	os.MkdirAll(absPath, 0755)
	cmd := exec.Command(binPath, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start chrome: %w", err)
	}
	activeProcs[port] = cmd.Process
	logger.DevPrintf("[Browser] Chrome launched: %s (Port: %d, PID: %d)\n", binPath, port, cmd.Process.Pid)
	for range 30 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if isChromeReady(port) {
				return nil
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
	return fmt.Errorf("timeout waiting for chrome to start on port %d", port)
}
