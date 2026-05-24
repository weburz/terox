package template

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
)

var templateDir = filepath.Join(xdg.DataHome, "terox")

// Cache resolves a "owner/repo" string to a local template directory,
// downloading and extracting the zipball if not already cached.
// When refresh is true, any existing cached copy is removed first so the
// template is re-downloaded.
// Returns the absolute path to the cached template root.
func Cache(repo string, refresh bool) (string, error) {
	owner, repository, err := splitRepo(repo)
	if err != nil {
		return "", err
	}

	// GitHub lowercases owner/repo in zipball folder names, so the cached
	// path is the lowercased version.
	dir := filepath.Join(templateDir, strings.ToLower(owner), strings.ToLower(repository))

	if refresh {
		if err := os.RemoveAll(dir); err != nil {
			return "", fmt.Errorf("refresh cache for %s: %w", repo, err)
		}
	} else {
		if _, err := os.Stat(dir); err == nil {
			return dir, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("stat template path %s: %w", dir, err)
		}
	}

	if refresh {
		fmt.Printf("Refreshing cache; downloading %s...\n", repo)
	} else {
		fmt.Printf("Template not found locally; downloading %s...\n", repo)
	}
	zipPath, err := downloadTemplate(repo)
	if err != nil {
		return "", err
	}
	defer func() { _ = os.Remove(zipPath) }()

	finalDest, err := extractTemplate(zipPath, templateDir)
	if err != nil {
		return "", err
	}
	return finalDest, nil
}

func splitRepo(repo string) (string, string, error) {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("template ref must be 'owner/repo', got %q", repo)
	}
	return parts[0], parts[1], nil
}

func downloadTemplate(repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/zipball", repo)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad server response: %d", resp.StatusCode)
	}

	tempFile, err := os.CreateTemp("", "terox-*.zip")
	if err != nil {
		return "", fmt.Errorf("create temp zip: %w", err)
	}

	if _, err := io.Copy(tempFile, resp.Body); err != nil {
		_ = tempFile.Close()
		return "", fmt.Errorf("write zip: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return "", fmt.Errorf("close zip: %w", err)
	}
	return tempFile.Name(), nil
}

func extractTemplate(zipfile, dest string) (string, error) {
	r, err := zip.OpenReader(zipfile)
	if err != nil {
		return "", fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	var topLevelFolder string
	for _, f := range r.File {
		parts := strings.SplitN(f.Name, "/", 2)
		if len(parts) > 1 && topLevelFolder == "" {
			topLevelFolder = parts[0]
			break
		}
	}
	if topLevelFolder == "" {
		return "", fmt.Errorf("failed to detect the top-level directory in the archive")
	}

	parts := strings.SplitN(topLevelFolder, "-", 3)
	if len(parts) < 2 {
		return "", fmt.Errorf("unexpected folder structure: %s", topLevelFolder)
	}
	owner := parts[0]
	repo := parts[1]

	finalDest := filepath.Join(dest, owner, repo)
	if err := os.MkdirAll(finalDest, 0o755); err != nil {
		return "", fmt.Errorf("create destination directory: %w", err)
	}

	for _, f := range r.File {
		if err := extractOne(f, topLevelFolder, finalDest); err != nil {
			return "", err
		}
	}
	return finalDest, nil
}

func extractOne(f *zip.File, topLevelFolder, finalDest string) error {
	relativePath := strings.TrimPrefix(f.Name, topLevelFolder+"/")
	if relativePath == "" {
		return nil
	}
	filePath := filepath.Join(finalDest, relativePath)

	if f.FileInfo().IsDir() {
		return os.MkdirAll(filePath, f.Mode())
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return fmt.Errorf("create parent dirs: %w", err)
	}

	srcFile, err := f.Open()
	if err != nil {
		return fmt.Errorf("open zip entry %s: %w", f.Name, err)
	}
	defer func() { _ = srcFile.Close() }()

	destFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return fmt.Errorf("create %s: %w", filePath, err)
	}
	defer func() { _ = destFile.Close() }()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("copy %s: %w", f.Name, err)
	}
	return nil
}
