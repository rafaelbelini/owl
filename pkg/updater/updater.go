package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
)

const (
	Repository   = "rafael/owl"
	LatestURL    = "https://api.github.com/repos/" + Repository + "/releases/latest"
	InstallShURL = "https://raw.githubusercontent.com/" + Repository + "/main/install.sh"
)

// UpdateInfo holds information about available updates
type UpdateInfo struct {
	CurrentVersion string
	LatestVersion  string
	URL            string
	IsUpdate      bool
}

// CheckForUpdate checks GitHub for a newer version
func CheckForUpdate(currentVersion string) (*UpdateInfo, error) {
	if currentVersion == "" {
		currentVersion = "0.0.0"
	}

	latest, err := getLatestVersion()
	if err != nil {
		return nil, err
	}

	info := &UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  latest,
		URL:            fmt.Sprintf("https://github.com/%s/releases/latest", Repository),
		IsUpdate:       compareVersions(currentVersion, latest) < 0,
	}

	return info, nil
}

// getLatestVersion fetches the latest release tag from GitHub
func getLatestVersion() (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", LatestURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent to avoid GitHub API rate limiting issues
	req.Header.Set("User-Agent", "owl-cli")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}

	if err := json.Unmarshal(body, &release); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Remove 'v' prefix if present
	tag := strings.TrimPrefix(release.TagName, "v")
	return tag, nil
}

// Upgrade performs the upgrade to the latest version
func Upgrade(currentVersion string) error {
	info, err := CheckForUpdate(currentVersion)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !info.IsUpdate {
		fmt.Printf("You are running the latest version: %s\n", currentVersion)
		return nil
	}

	fmt.Printf("Upgrading from %s to %s...\n", currentVersion, info.LatestVersion)

	// Detect OS and architecture
	targetOS := runtime.GOOS
	arch := runtime.GOARCH

	// Normalize architecture names
	switch arch {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	default:
		return fmt.Errorf("unsupported architecture: %s", arch)
	}

	// Normalize OS names
	switch targetOS {
	case "linux":
		targetOS = "linux"
	case "darwin":
		targetOS = "darwin"
	case "windows":
		targetOS = "windows"
	default:
		return fmt.Errorf("unsupported operating system: %s", targetOS)
	}

	// Build the binary name
	binaryName := fmt.Sprintf("owl-%s-%s", targetOS, arch)
	if targetOS == "windows" {
		binaryName += ".exe"
	}

	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", Repository, info.LatestVersion, binaryName)

	fmt.Printf("Downloading %s...\n", downloadURL)

	// Create temp file
	tmpFile, err := os.CreateTemp("", "owl-")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Download the binary
	if err := downloadFile(downloadURL, tmpPath); err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}

	// Make it executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Get install directory from config or use default
	installDir := os.Getenv("OWL_INSTALL_DIR")
	if installDir == "" {
		installDir = "/usr/local/bin"
	}

	targetPath := installDir + "/owl"

	// Try to replace the current binary
	if err := os.Rename(tmpPath, targetPath); err != nil {
		// If rename fails (e.g., different filesystem), try copy
		if err := copyFile(tmpPath, targetPath); err != nil {
			return fmt.Errorf("failed to install: %w", err)
		}
		os.Remove(tmpPath)
	}

	fmt.Printf("Successfully upgraded to %s\n", info.LatestVersion)
	fmt.Printf("Installed to: %s\n", targetPath)

	return nil
}

// downloadFile downloads a file from url to path
func downloadFile(urlStr, path string) error {
	// Parse URL to ensure it's valid
	_, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	resp, err := http.Get(urlStr)
	if err != nil {
		return fmt.Errorf("HTTP GET failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
	 return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return dstFile.Sync()
}

// compareVersions compares two version strings (semver-like)
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < len(parts1) || i < len(parts2); i++ {
		var p1, p2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &p1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &p2)
		}

		if p1 < p2 {
			return -1
		}
		if p1 > p2 {
			return 1
		}
	}

	return 0
}
