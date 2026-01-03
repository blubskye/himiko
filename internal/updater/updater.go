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
	CurrentVersion = "1.8.0"
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

// IsGitRepository checks if the current directory is a git repository
func IsGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

// CheckForSourceUpdate checks if there are new commits on the remote repository
func CheckForSourceUpdate() (*UpdateInfo, error) {
	if !IsGitRepository() {
		return nil, fmt.Errorf("not in a git repository")
	}

	// Fetch latest changes from remote
	fetchCmd := exec.Command("git", "fetch", "origin")
	if err := fetchCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to fetch from remote: %w", err)
	}

	// Get current branch
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, err := branchCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}
	branch := strings.TrimSpace(string(branchOut))

	// Check if there are new commits
	revListCmd := exec.Command("git", "rev-list", "HEAD..origin/"+branch, "--count")
	revListOut, err := revListCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}

	behindCount := strings.TrimSpace(string(revListOut))
	available := behindCount != "0"

	// Get remote version tag if available
	tagCmd := exec.Command("git", "describe", "--tags", "--abbrev=0", "origin/"+branch)
	tagOut, _ := tagCmd.Output()
	newVersion := strings.TrimSpace(string(tagOut))
	if newVersion == "" {
		newVersion = "latest"
	} else {
		newVersion = strings.TrimPrefix(newVersion, "v")
	}

	// Get commit messages since last update
	logCmd := exec.Command("git", "log", "HEAD..origin/"+branch, "--oneline", "--no-decorate")
	logOut, _ := logCmd.Output()
	releaseNotes := string(logOut)
	if releaseNotes == "" {
		releaseNotes = "No new commits"
	}

	info := &UpdateInfo{
		Available:      available,
		CurrentVersion: CurrentVersion,
		NewVersion:     newVersion,
		ReleaseNotes:   releaseNotes,
	}

	return info, nil
}

// ApplySourceUpdate pulls the latest code and rebuilds the binary
func ApplySourceUpdate(progressFn func(step, total int, message string)) error {
	if !IsGitRepository() {
		return fmt.Errorf("not in a git repository")
	}

	totalSteps := 5
	currentStep := 0

	// Step 1: Stash any local changes
	currentStep++
	if progressFn != nil {
		progressFn(currentStep, totalSteps, "Stashing local changes...")
	}
	stashCmd := exec.Command("git", "stash", "push", "-m", "Auto-update stash")
	stashCmd.Run() // Ignore errors - may have nothing to stash

	// Step 2: Pull latest changes
	currentStep++
	if progressFn != nil {
		progressFn(currentStep, totalSteps, "Pulling latest changes...")
	}
	pullCmd := exec.Command("git", "pull", "origin")
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr
	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("failed to pull updates: %w", err)
	}

	// Step 3: Get executable path
	currentStep++
	if progressFn != nil {
		progressFn(currentStep, totalSteps, "Determining build path...")
	}
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Step 4: Build new binary
	currentStep++
	if progressFn != nil {
		progressFn(currentStep, totalSteps, "Building from source...")
	}

	tmpBinary := execPath + ".new"
	buildCmd := exec.Command("go", "build", "-o", tmpBinary, "./cmd/himiko")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("failed to build: %w", err)
	}

	// Step 5: Replace binary
	currentStep++
	if progressFn != nil {
		progressFn(currentStep, totalSteps, "Installing new binary...")
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

	// Make executable (Unix)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(execPath, 0755); err != nil {
			fmt.Printf("Warning: Failed to set permissions: %v\n", err)
		}
	}

	// Remove backup
	os.Remove(backupPath)

	if progressFn != nil {
		progressFn(totalSteps, totalSteps, "Update complete!")
	}

	return nil
}
