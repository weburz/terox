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

// Cache resolves a template ref to a local directory, downloading and
// extracting the zipball if not already cached. The ref is either
// "owner/repo" (the repo root is the template) or "owner/repo/subpath"
// (a directory inside the repo is the template — useful for monorepos
// like weburz/terox-templates).
//
// When refresh is true, any existing cached copy of the repo is removed
// first so the template is re-downloaded.
// Returns the absolute path to the template directory (the repo root,
// or the resolved subpath within it).
func Cache(ref string, refresh bool) (string, error) {
	owner, repository, subpath, err := splitRef(ref)
	if err != nil {
		return "", err
	}

	repoRef := owner + "/" + repository
	// GitHub lowercases owner/repo in zipball folder names, so the cached
	// path is the lowercased version.
	dir := filepath.Join(templateDir, strings.ToLower(owner), strings.ToLower(repository))

	if refresh {
		if err := os.RemoveAll(dir); err != nil {
			return "", fmt.Errorf("refresh cache for %s: %w", repoRef, err)
		}
	}

	cacheHit := false
	if !refresh {
		if _, err := os.Stat(dir); err == nil {
			cacheHit = true
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("stat template path %s: %w", dir, err)
		}
	}

	if !cacheHit {
		if refresh {
			fmt.Printf("Refreshing cache; downloading %s...\n", repoRef)
		} else {
			fmt.Printf("Template not found locally; downloading %s...\n", repoRef)
		}
		zipPath, err := downloadTemplate(repoRef)
		if err != nil {
			return "", err
		}
		defer func() { _ = os.Remove(zipPath) }()

		if err := extractTemplate(zipPath, dir); err != nil {
			return "", err
		}
	}

	return resolveSubpath(dir, subpath, repoRef)
}

// splitRef parses a template ref of the form "owner/repo" or
// "owner/repo/subpath". The subpath, if present, must be a relative,
// forward-slash path with no empty, "." or ".." segments.
func splitRef(ref string) (owner, repo, subpath string, err error) {
	parts := strings.SplitN(ref, "/", 3)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", "", fmt.Errorf("template ref must be 'owner/repo' or 'owner/repo/subpath', got %q", ref)
	}
	owner = parts[0]
	repo = parts[1]
	if len(parts) == 3 {
		subpath = parts[2]
		if subpath == "" {
			return "", "", "", fmt.Errorf("template ref subpath must not be empty, got %q", ref)
		}
		if err := validateSubpath(subpath); err != nil {
			return "", "", "", fmt.Errorf("invalid subpath in %q: %w", ref, err)
		}
	}
	return owner, repo, subpath, nil
}

// validateSubpath rejects subpaths that escape the repo root or contain
// empty/dot segments. The input is always forward-slash separated since
// it comes from a user-supplied ref, not a filesystem path.
func validateSubpath(p string) error {
	if strings.HasPrefix(p, "/") {
		return fmt.Errorf("subpath must be relative")
	}
	for _, seg := range strings.Split(p, "/") {
		switch seg {
		case "":
			return fmt.Errorf("subpath must not contain empty segments")
		case ".", "..":
			return fmt.Errorf("subpath segment %q is not allowed", seg)
		}
	}
	return nil
}

func resolveSubpath(root, subpath, repoRef string) (string, error) {
	if subpath == "" {
		return root, nil
	}
	dest := filepath.Join(root, filepath.FromSlash(subpath))
	info, err := os.Stat(dest)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("subpath %q not found in %s", subpath, repoRef)
		}
		return "", fmt.Errorf("stat subpath %s: %w", dest, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("subpath %q in %s is not a directory", subpath, repoRef)
	}
	return dest, nil
}

func downloadTemplate(repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/zipball", repo)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

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

// extractTemplate unzips a GitHub zipball into dest. GitHub zipballs
// always have a single top-level folder of the form
// "{owner}-{repo}-{sha-prefix}"; that folder is detected dynamically and
// stripped from each entry's path. dest is supplied by the caller because
// the folder name cannot be reliably split back into owner/repo when
// either contains hyphens (e.g. "weburz-terox-templates-abc123").
func extractTemplate(zipfile, dest string) error {
	r, err := zip.OpenReader(zipfile)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer func() { _ = r.Close() }()

	var topLevelFolder string
	for _, f := range r.File {
		parts := strings.SplitN(f.Name, "/", 2)
		if len(parts) > 1 && topLevelFolder == "" {
			topLevelFolder = parts[0]
			break
		}
	}
	if topLevelFolder == "" {
		return fmt.Errorf("failed to detect the top-level directory in the archive")
	}

	if err := os.MkdirAll(dest, 0o755); err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}

	for _, f := range r.File {
		if err := extractOne(f, topLevelFolder, dest); err != nil {
			return err
		}
	}
	return nil
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
