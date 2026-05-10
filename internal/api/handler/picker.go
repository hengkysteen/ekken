package handler

import (
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) OpenFilePicker(c *gin.Context) {
	multiple := c.Query("multiple") == "true"
	directory := c.Query("directory") == "true"
	
	// Determine title on Backend side
	title := "Choose File"
	if directory {
		title = "Choose Folder"
	} else if multiple {
		title = "Choose Files"
	}

	var files []string
	var status string
	var err error

	switch runtime.GOOS {
	case "darwin":
		files, status, err = openPickerMac(multiple, directory, title)
	case "windows":
		files, status, err = openPickerWindows(multiple, directory, title)
	case "linux":
		files, status, err = openPickerLinux(multiple, directory, title)
	default:
		status = "error"
		files = []string{fmt.Sprintf("unsupported operating system: %s", runtime.GOOS)}
	}

	if err != nil && status == "" {
		status = "error"
		files = []string{err.Error()}
	}

	isOk := status == "ok" || status == "canceled"

	c.JSON(http.StatusOK, gin.H{
		"ok":     isOk,
		"status": status,
		"data":   files,
	})
}

func openPickerMac(multiple bool, directory bool, title string) ([]string, string, error) {
	cmdType := "file"
	if directory {
		cmdType = "folder"
	}

	script := fmt.Sprintf(`POSIX path of (choose %s with prompt "%s")`, cmdType, title)
	if multiple {
		script = fmt.Sprintf(`
		set theItems to choose %s with prompt "%s" with multiple selections allowed
		set resultPaths to ""
		repeat with anItem in theItems
			set resultPaths to resultPaths & POSIX path of anItem & "\n"
		end repeat
		return resultPaths
		`, cmdType, title)
	}
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return nil, "canceled", nil
		}
		return nil, "error", err
	}
	
	output := strings.TrimSpace(string(out))
	if output == "" {
		return []string{}, "ok", nil
	}
	
	paths := strings.Split(output, "\n")
	var finalPaths []string
	for _, p := range paths {
		if strings.TrimSpace(p) != "" {
			finalPaths = append(finalPaths, strings.TrimSpace(p))
		}
	}
	
	return finalPaths, "ok", nil
}

func openPickerWindows(multiple bool, directory bool, title string) ([]string, string, error) {
	if directory {
		// FolderBrowserDialog doesn't support multiple selection easily in standard Forms
		psScript := fmt.Sprintf(`
		Add-Type -AssemblyName System.Windows.Forms
		$f = New-Object System.Windows.Forms.FolderBrowserDialog
		$f.Description = "%s"
		if ($f.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
			$f.SelectedPath
		} else {
			"__CANCELED__"
		}
		`, title)
		cmd := exec.Command("powershell", "-Command", psScript)
		out, err := cmd.Output()
		if err != nil {
			return nil, "error", err
		}
		result := strings.TrimSpace(string(out))
		if result == "__CANCELED__" || result == "" {
			return nil, "canceled", nil
		}
		return []string{result}, "ok", nil
	}

	multiFlag := "$false"
	resultLine := "$f.FileName"
	if multiple {
		multiFlag = "$true"
		resultLine = "$f.FileNames | ForEach-Object { $_ }"
	}

	psScript := fmt.Sprintf(`
	Add-Type -AssemblyName System.Windows.Forms
	$f = New-Object System.Windows.Forms.OpenFileDialog
	$f.Filter = "All Files (*.*)|*.*"
	$f.Title = "%s"
	$f.ShowHelp = $true
	$f.Multiselect = %s
	if ($f.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
		%s
	} else {
		"__CANCELED__"
	}
	`, title, multiFlag, resultLine)
	cmd := exec.Command("powershell", "-Command", psScript)
	out, err := cmd.Output()
	if err != nil {
		return nil, "error", err
	}
	result := strings.TrimSpace(string(out))
	if result == "__CANCELED__" || result == "" {
		return nil, "canceled", nil
	}
	
	paths := strings.Split(result, "\r\n")
	var finalPaths []string
	for _, p := range paths {
		if strings.TrimSpace(p) != "" {
			finalPaths = append(finalPaths, strings.TrimSpace(p))
		}
	}
	return finalPaths, "ok", nil
}

func openPickerLinux(multiple bool, directory bool, title string) ([]string, string, error) {
	if _, err := exec.LookPath("zenity"); err == nil {
		args := []string{"--file-selection", fmt.Sprintf("--title=%s", title)}
		if directory {
			args = append(args, "--directory")
		}
		if multiple {
			args = append(args, "--multiple", "--separator=\n")
		}
		cmd := exec.Command("zenity", args...)
		out, err := cmd.Output()
		if err != nil {
			return nil, "canceled", nil
		}
		paths := strings.Split(strings.TrimSpace(string(out)), "\n")
		return paths, "ok", nil
	}

	if _, err := exec.LookPath("kdialog"); err == nil {
		cmdName := "--getopenfilename"
		if directory {
			cmdName = "--getexistingdirectory"
		}
		args := []string{cmdName, ".", "All Files (*)", fmt.Sprintf("--title=%s", title)}
		if multiple && !directory { // kdialog directory picker usually doesn't support multiple
			args = append(args, "--multiple")
		}
		cmd := exec.Command("kdialog", args...)
		out, err := cmd.Output()
		if err != nil {
			return nil, "canceled", nil
		}
		paths := strings.Split(strings.TrimSpace(string(out)), "\n")
		return paths, "ok", nil
	}

	return nil, "error", fmt.Errorf("NATIVE_PICKER_NOT_FOUND: Zenity or KDialog is required for native file picking on Linux. Please install it first.")
}
