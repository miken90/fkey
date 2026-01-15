package services

// Auto-updater service using GitHub Releases API
// Checks for new versions and notifies user

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	GitHubOwner      = "miken90"
	GitHubRepo       = "gonhanh.org"
	GitHubAPIURL     = "https://api.github.com/repos/%s/%s/releases/latest"
	CheckInterval    = 24 * time.Hour // Check once per day
	UserAgent        = "FKey-Updater/1.0"
)

// Release represents a GitHub release
type Release struct {
	TagName     string  `json:"tag_name"`
	Name        string  `json:"name"`
	Body        string  `json:"body"`
	HTMLURL     string  `json:"html_url"`
	PublishedAt string  `json:"published_at"`
	Assets      []Asset `json:"assets"`
}

// Asset represents a release asset (downloadable file)
type Asset struct {
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
	ContentType        string `json:"content_type"`
}

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	Available      bool   `json:"available"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseNotes   string `json:"releaseNotes"`
	DownloadURL    string `json:"downloadURL"`
	ReleaseURL     string `json:"releaseURL"`
	AssetName      string `json:"assetName"`
	AssetSize      int64  `json:"assetSize"`
}

// UpdaterService manages auto-update checks
type UpdaterService struct {
	currentVersion string
	lastCheck      time.Time
	cachedInfo     *UpdateInfo
}

// NewUpdaterService creates a new updater service
func NewUpdaterService(currentVersion string) *UpdaterService {
	return &UpdaterService{
		currentVersion: currentVersion,
	}
}

// CheckForUpdates checks GitHub for a newer version
func (u *UpdaterService) CheckForUpdates(force bool) (*UpdateInfo, error) {
	// Use cache if checked recently (unless forced)
	if !force && u.cachedInfo != nil && time.Since(u.lastCheck) < CheckInterval {
		return u.cachedInfo, nil
	}

	release, err := u.fetchLatestRelease()
	if err != nil {
		return nil, err
	}

	info := u.compareVersions(release)
	u.cachedInfo = info
	u.lastCheck = time.Now()

	return info, nil
}

// fetchLatestRelease gets the latest release from GitHub API
func (u *UpdaterService) fetchLatestRelease() (*Release, error) {
	url := fmt.Sprintf(GitHubAPIURL, GitHubOwner, GitHubRepo)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("no releases found")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	return &release, nil
}

// compareVersions compares current version with release
func (u *UpdaterService) compareVersions(release *Release) *UpdateInfo {
	info := &UpdateInfo{
		CurrentVersion: u.currentVersion,
		LatestVersion:  release.TagName,
		ReleaseNotes:   release.Body,
		ReleaseURL:     release.HTMLURL,
	}

	// Find Windows asset
	for _, asset := range release.Assets {
		if u.IsWindowsAsset(asset.Name) {
			info.DownloadURL = asset.BrowserDownloadURL
			info.AssetName = asset.Name
			info.AssetSize = asset.Size
			break
		}
	}

	// Compare versions (strip 'v' prefix if present)
	current := strings.TrimPrefix(u.currentVersion, "v")
	latest := strings.TrimPrefix(release.TagName, "v")

	// Simple version comparison (works for semver)
	info.Available = u.IsNewerVersion(current, latest)

	return info
}

// IsWindowsAsset checks if asset is for Windows (exported for testing)
func (u *UpdaterService) IsWindowsAsset(name string) bool {
	name = strings.ToLower(name)
	
	// Exclude non-Windows platforms explicitly
	if strings.Contains(name, "darwin") || strings.Contains(name, "macos") || 
	   strings.Contains(name, "linux") || strings.Contains(name, "mac") {
		return false
	}
	
	return strings.Contains(name, "windows") ||
		strings.Contains(name, "win64") ||
		strings.Contains(name, "win32") ||
		strings.HasSuffix(name, ".exe") ||
		(strings.HasSuffix(name, ".zip") && (strings.Contains(name, "win") || strings.Contains(name, "fkey")))
}

// IsNewerVersion compares two semver strings (exported for testing)
func (u *UpdaterService) IsNewerVersion(current, latest string) bool {
	// Strip 'v' prefix if present
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")
	
	// Remove any suffix like "-wails", "-beta", etc. for comparison
	current = strings.Split(current, "-")[0]
	latest = strings.Split(latest, "-")[0]

	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	for i := 0; i < 3; i++ {
		var c, l int
		if i < len(currentParts) {
			fmt.Sscanf(currentParts[i], "%d", &c)
		}
		if i < len(latestParts) {
			fmt.Sscanf(latestParts[i], "%d", &l)
		}

		if l > c {
			return true
		}
		if l < c {
			return false
		}
	}

	return false
}

// DownloadUpdate downloads the update to temp directory
func (u *UpdaterService) DownloadUpdate(downloadURL string, progressCb func(downloaded, total int64)) (string, error) {
	client := &http.Client{Timeout: 5 * time.Minute}
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download error: %d", resp.StatusCode)
	}

	// Create temp file
	tempDir := os.TempDir()
	fileName := filepath.Base(downloadURL)
	tempFile := filepath.Join(tempDir, "fkey-update-"+fileName)

	out, err := os.Create(tempFile)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()

	// Download with progress
	var downloaded int64
	total := resp.ContentLength
	buf := make([]byte, 32*1024)

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			out.Write(buf[:n])
			downloaded += int64(n)
			if progressCb != nil {
				progressCb(downloaded, total)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("download error: %w", err)
		}
	}

	return tempFile, nil
}

// OpenReleasePage opens the release page in browser
func (u *UpdaterService) OpenReleasePage(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

// GetCurrentVersion returns the current version
func (u *UpdaterService) GetCurrentVersion() string {
	return u.currentVersion
}
