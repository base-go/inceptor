package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	githubRepo     = "base-go/inceptor"
	releaseBaseURL = "https://github.com/base-go/inceptor/releases/latest/download"
)

// handleGetVersion returns current and latest version
func (s *Server) handleGetVersion(c *gin.Context) {
	current := s.version

	// Fetch latest version from GitHub releases
	latest := current
	updateAvailable := false

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo))
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		var release struct {
			TagName string `json:"tag_name"`
		}
		if json.NewDecoder(resp.Body).Decode(&release) == nil && release.TagName != "" {
			latest = strings.TrimPrefix(release.TagName, "v")
			// Compare versions semantically
			if compareVersions(latest, current) > 0 {
				updateAvailable = true
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"current":         current,
		"latest":          latest,
		"updateAvailable": updateAvailable,
	})
}

// compareVersions compares two semver strings, returns 1 if a > b, -1 if a < b, 0 if equal
func compareVersions(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := 0; i < maxLen; i++ {
		var aNum, bNum int
		if i < len(aParts) {
			fmt.Sscanf(aParts[i], "%d", &aNum)
		}
		if i < len(bParts) {
			fmt.Sscanf(bParts[i], "%d", &bNum)
		}
		if aNum > bNum {
			return 1
		}
		if aNum < bNum {
			return -1
		}
	}
	return 0
}

// handleSystemUpdate triggers a self-update
func (s *Server) handleSystemUpdate(c *gin.Context) {
	// Determine binary path and architecture
	execPath, err := os.Executable()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot determine executable path"})
		return
	}

	// Use runtime OS and architecture
	goos := runtime.GOOS
	arch := runtime.GOARCH
	if arch == "" {
		arch = "amd64"
	}

	// Download URL
	downloadURL := fmt.Sprintf("%s/inceptor-%s-%s", releaseBaseURL, goos, arch)

	// Download new binary to temp file
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(downloadURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download update: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Download failed with status: %d", resp.StatusCode)})
		return
	}

	// Write to temp file
	tmpFile, err := os.CreateTemp("", "inceptor-update-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp file: " + err.Error()})
		return
	}
	tmpPath := tmpFile.Name()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write update: " + err.Error()})
		return
	}
	tmpFile.Close()

	// Make executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set permissions: " + err.Error()})
		return
	}

	// Replace current binary (atomic move)
	if err := os.Rename(tmpPath, execPath); err != nil {
		// Try copy if rename fails (cross-device or permission issue)
		srcFile, err := os.Open(tmpPath)
		if err != nil {
			os.Remove(tmpPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open temp file: " + err.Error()})
			return
		}
		defer srcFile.Close()

		dstFile, err := os.Create(execPath)
		if err != nil {
			os.Remove(tmpPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to replace binary (permission denied?): " + err.Error()})
			return
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			os.Remove(tmpPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write binary: " + err.Error()})
			return
		}
		os.Remove(tmpPath)
	}

	// Send response first, then trigger restart in background
	c.JSON(http.StatusOK, gin.H{
		"status":  "updated",
		"message": "Update complete. Restarting service...",
	})

	// Restart properly via service manager
	go func() {
		time.Sleep(1 * time.Second) // Give time for response to be sent

		if runtime.GOOS == "darwin" {
			// macOS: try system daemon first, then user daemon
			if err := exec.Command("launchctl", "kickstart", "-k", "system/com.inceptor").Run(); err != nil {
				// Fallback to user daemon
				exec.Command("launchctl", "kickstart", "-k", fmt.Sprintf("gui/%d/com.inceptor", os.Getuid())).Run()
			}
		} else {
			// Linux: try system service first, then user service
			if err := exec.Command("systemctl", "restart", "inceptor").Run(); err != nil {
				exec.Command("systemctl", "--user", "restart", "inceptor").Run()
			}
		}
	}()
}
