// Himiko Discord Bot
// Copyright (C) 2025 Himiko Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package updater

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

const (
	GitHubRepo    = "blubskye/himiko"
	GitHubAPIURL  = "https://api.github.com/repos/" + GitHubRepo + "/releases/latest"
	CurrentVersion = "1.5.7"
)

// Release represents a GitHub release
type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Body    string  `json:"body"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	Available      bool
	CurrentVersion string
	NewVersion     string
	ReleaseNotes   string
	DownloadURL    string
	AssetName      string
	Size           int64
}

// CheckForUpdate checks GitHub for a newer release
func CheckForUpdate() (*UpdateInfo, error) {
	resp, err := http.Get(GitHubAPIURL)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	// Clean version tags (remove 'v' prefix if present)
	newVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(CurrentVersion, "v")

	info := &UpdateInfo{
		CurrentVersion: currentVersion,
		NewVersion:     newVersion,
		ReleaseNotes:   release.Body,
	}

	// Check if newer version
	if !isNewerVersion(currentVersion, newVersion) {
		info.Available = false
		return info, nil
	}

	// Find appropriate asset for current OS/arch
	assetName := getAssetName()
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			info.Available = true
			info.DownloadURL = asset.BrowserDownloadURL
			info.AssetName = asset.Name
			info.Size = asset.Size
			return info, nil
		}
	}

	return nil, fmt.Errorf("no compatible release found for %s/%s", runtime.GOOS, runtime.GOARCH)
}

// getAssetName returns the expected asset name for current platform
func getAssetName() string {
	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("himiko-v%s-windows-%s.zip", CurrentVersion, runtime.GOARCH)
	case "linux":
		return fmt.Sprintf("himiko-v%s-linux-%s.zip", CurrentVersion, runtime.GOARCH)
	case "darwin":
		return fmt.Sprintf("himiko-v%s-darwin-%s.zip", CurrentVersion, runtime.GOARCH)
	default:
		return ""
	}
}

// getExpectedAssetPattern returns a pattern to match assets for current platform
func getExpectedAssetPattern() string {
	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("himiko-v*-windows-%s.zip", runtime.GOARCH)
	case "linux":
		return fmt.Sprintf("himiko-v*-linux-%s.zip", runtime.GOARCH)
	case "darwin":
		return fmt.Sprintf("himiko-v*-darwin-%s.zip", runtime.GOARCH)
	default:
		return ""
	}
}

// matchesAssetPattern checks if an asset name matches our platform
func matchesAssetPattern(name string) bool {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	expectedSuffix := fmt.Sprintf("-%s-%s.zip", goos, goarch)
	return strings.HasPrefix(name, "himiko-v") && strings.HasSuffix(name, expectedSuffix)
}

// CheckForUpdateByPattern checks for updates matching platform pattern
func CheckForUpdateByPattern() (*UpdateInfo, error) {
	resp, err := http.Get(GitHubAPIURL)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	newVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(CurrentVersion, "v")

	info := &UpdateInfo{
		CurrentVersion: currentVersion,
		NewVersion:     newVersion,
		ReleaseNotes:   release.Body,
	}

	if !isNewerVersion(currentVersion, newVersion) {
		info.Available = false
		return info, nil
	}

	// Find asset matching our platform pattern
	for _, asset := range release.Assets {
		if matchesAssetPattern(asset.Name) {
			info.Available = true
			info.DownloadURL = asset.BrowserDownloadURL
			info.AssetName = asset.Name
			info.Size = asset.Size
			return info, nil
		}
	}

	return nil, fmt.Errorf("no compatible release found for %s/%s", runtime.GOOS, runtime.GOARCH)
}

// isNewerVersion compares semver versions
func isNewerVersion(current, new string) bool {
	currentParts := parseVersion(current)
	newParts := parseVersion(new)

	for i := 0; i < 3; i++ {
		if newParts[i] > currentParts[i] {
			return true
		}
		if newParts[i] < currentParts[i] {
			return false
		}
	}
	return false
}

// parseVersion parses a semver string into [major, minor, patch]
func parseVersion(v string) [3]int {
	var parts [3]int
	v = strings.TrimPrefix(v, "v")

	fmt.Sscanf(v, "%d.%d.%d", &parts[0], &parts[1], &parts[2])
	return parts
}

// DownloadUpdate downloads the update to a temporary file
func DownloadUpdate(info *UpdateInfo, progressFn func(downloaded, total int64)) (string, error) {
	resp, err := http.Get(info.DownloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	// Create temp file
	tmpFile, err := os.CreateTemp("", "himiko-update-*.zip")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	// Download with progress
	var downloaded int64
	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := tmpFile.Write(buf[:n])
			if writeErr != nil {
				os.Remove(tmpFile.Name())
				return "", fmt.Errorf("failed to write update: %w", writeErr)
			}
			downloaded += int64(n)
			if progressFn != nil {
				progressFn(downloaded, info.Size)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			os.Remove(tmpFile.Name())
			return "", fmt.Errorf("failed to download update: %w", err)
		}
	}

	return tmpFile.Name(), nil
}

// ApplyUpdate extracts the update and replaces the binary
func ApplyUpdate(zipPath string) error {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	execDir := filepath.Dir(execPath)
	execName := filepath.Base(execPath)

	// Open zip file
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open update archive: %w", err)
	}
	defer r.Close()

	// Find the binary in the zip
	var binaryFile *zip.File
	binaryName := "himiko-linux-amd64"
	if runtime.GOOS == "windows" {
		binaryName = "himiko-windows-amd64.exe"
	} else if runtime.GOOS == "darwin" {
		binaryName = "himiko-darwin-amd64"
	}

	for _, f := range r.File {
		if f.Name == binaryName || strings.HasSuffix(f.Name, "/"+binaryName) {
			binaryFile = f
			break
		}
	}

	if binaryFile == nil {
		return fmt.Errorf("binary not found in update archive")
	}

	// Extract to temp file
	tmpBinary := filepath.Join(execDir, execName+".new")
	if err := extractFile(binaryFile, tmpBinary); err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}

	// Make executable
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tmpBinary, 0755); err != nil {
			os.Remove(tmpBinary)
			return fmt.Errorf("failed to set permissions: %w", err)
		}
	}

	// Backup old binary
	backupPath := execPath + ".old"
	if err := os.Rename(execPath, backupPath); err != nil {
		os.Remove(tmpBinary)
		return fmt.Errorf("failed to backup old binary: %w", err)
	}

	// Move new binary into place
	if err := os.Rename(tmpBinary, execPath); err != nil {
		// Try to restore backup
		os.Rename(backupPath, execPath)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	// Remove backup
	os.Remove(backupPath)

	// Clean up zip
	os.Remove(zipPath)

	return nil
}

// extractFile extracts a single file from a zip
func extractFile(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	outFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, rc)
	return err
}

// GetCurrentVersion returns the current version
func GetCurrentVersion() string {
	return CurrentVersion
}

// RelaunchAfterUpdate relaunches the bot executable after an update
// This uses exec on Unix systems to replace the current process
func RelaunchAfterUpdate() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// On Windows, we need to start a new process and exit
	if runtime.GOOS == "windows" {
		cmd := exec.Command(execPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start new process: %w", err)
		}
		// Exit the current process
		os.Exit(0)
		return nil
	}

	// On Unix systems, use syscall.Exec to replace the current process
	// This preserves the PID and cleanly transitions to the new binary
	return syscall.Exec(execPath, []string{execPath}, os.Environ())
}
