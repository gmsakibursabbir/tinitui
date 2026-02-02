package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	RepoOwner = "gmsakibursabbir"
	RepoName  = "tinitui"
)

type Release struct {
	TagName string `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name        string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func GetLatestVersion() (string, *Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", RepoOwner, RepoName)
	resp, err := http.Get(url)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("failed to check update: %s", resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", nil, err
	}

	return release.TagName, &release, nil
}

func IsNewer(current, latest string) bool {
    // Basic comparison assuming vX.Y.Z
    // For robust comparison we might want a semver lib, but text compare works for strict format
    // Ignoring 'v' prefix
    // Ignoring 'v' prefix
    c := strings.TrimPrefix(current, "v")
    l := strings.TrimPrefix(latest, "v")
    return compareVersions(c, l)
}

func compareVersions(v1, v2 string) bool {
	p1 := strings.Split(v1, ".")
	p2 := strings.Split(v2, ".")
	len1 := len(p1)
	len2 := len(p2)
	maxLen := len1
	if len2 > maxLen { maxLen = len2 }

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len1 { fmt.Sscanf(p1[i], "%d", &n1) }
		if i < len2 { fmt.Sscanf(p2[i], "%d", &n2) }
		if n2 > n1 { return true }
		if n1 > n2 { return false }
	}
	return false
}

func Update(release *Release) error {
	// Find asset
	goOS := runtime.GOOS
	goArch := runtime.GOARCH

	// Expected name pattern: tinitui-{os}-{arch}
	// e.g. tinitui-linux-amd64
	targetName := fmt.Sprintf("tinitui-%s-%s", goOS, goArch)
	if goOS == "windows" {
		targetName += ".exe"
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == targetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no binary found for %s/%s", goOS, goArch)
	}

	// Download
	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Temp file
	tmpFile, err := os.CreateTemp("", "tinytui-update-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	// Validate basic
	info, err := os.Stat(tmpFile.Name())
	if err != nil || info.Size() == 0 {
		return fmt.Errorf("download failed (empty file)")
	}

	// Chmod
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return err
	}

	// Replace
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return err
	}

	// Safe rename with fallback
	if err := moveFile(tmpFile.Name(), exePath); err != nil {
		return err
	}

	return nil
}

// moveFile attempts to rename, falling back to copy-delete if cross-device link error occurs
func moveFile(source, destination string) error {
	err := os.Rename(source, destination)
	if err == nil {
		return nil
	}

	// If rename failed, try copy and delete (likely cross-device)
	// Open source
	srcFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create dest (atomic-ish: create temp next to dest then rename)
	destDir := filepath.Dir(destination)
	tmpDest, err := os.CreateTemp(destDir, ".tinitui-update-*")
	if err != nil {
		// Fallback to direct create if temp fails (permissions?)
		// But let's try direct create as last resort or just error.
		return fmt.Errorf("failed to create temp file in target dir: %w", err)
	}
	defer os.Remove(tmpDest.Name()) // Clean up if we fail before rename

	// Copy
	if _, err := io.Copy(tmpDest, srcFile); err != nil {
		tmpDest.Close()
		return err
	}
	
	// Preserve permissions
	info, err := srcFile.Stat()
	if err == nil {
		tmpDest.Chmod(info.Mode())
	} else {
		tmpDest.Chmod(0755)
	}
	
	tmpDest.Close()

	// Atomic Rename in target fs
	if err := os.Rename(tmpDest.Name(), destination); err != nil {
		return err
	}
	
	// Clean up source
	os.Remove(source)
	
	return nil
}
