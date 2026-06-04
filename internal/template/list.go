package template

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
)

// List prints the templates already downloaded into the local cache.
// The cache layout is "{templateDir}/{owner}/{repo}/...", so this walks
// two levels deep to surface "owner/repo" entries.
func List() error {
	owners, err := os.ReadDir(templateDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No templates cached locally.")
			return nil
		}
		return fmt.Errorf("read %s: %w", templateDir, err)
	}

	var refs []string
	for _, owner := range owners {
		if !owner.IsDir() {
			continue
		}
		repos, err := os.ReadDir(filepath.Join(templateDir, owner.Name()))
		if err != nil {
			continue
		}
		for _, repo := range repos {
			if !repo.IsDir() {
				continue
			}
			refs = append(refs, owner.Name()+"/"+repo.Name())
		}
	}

	if len(refs) == 0 {
		fmt.Println("No templates cached locally.")
		return nil
	}

	sort.Strings(refs)
	fmt.Println("Cached templates:")
	for _, r := range refs {
		fmt.Printf("  %s\n", r)
	}
	return nil
}

// ListRemote prints the templates available at the root of a remote
// GitHub repository. The ref must be of the form "owner/repo"; listing
// a subpath is not supported. Each top-level directory is treated as a
// candidate template; a terox.json inside it (if present) supplies the
// description shown in the output.
func ListRemote(ref string) error {
	owner, repo, subpath, err := splitRef(ref)
	if err != nil {
		return err
	}
	if subpath != "" {
		return fmt.Errorf("listing a subpath is not supported; use just 'owner/repo'")
	}

	dirs, err := fetchRepoTopLevelDirs(owner, repo)
	if err != nil {
		return err
	}
	dirs = filterListableDirs(dirs)
	if len(dirs) == 0 {
		fmt.Printf("No templates found in %s/%s.\n", owner, repo)
		return nil
	}

	sort.Strings(dirs)
	entries := make([]catalogueEntry, 0, len(dirs))
	for _, d := range dirs {
		desc, err := fetchTemplateDescription(owner, repo, d)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
		}
		entries = append(entries, catalogueEntry{Name: d, Description: desc})
	}

	printCatalogue(os.Stdout, owner+"/"+repo, entries)
	return nil
}

type catalogueEntry struct {
	Name        string
	Description string
}

// nonTemplateDirs is the small denylist of repo-root directories that
// are almost never templates and would otherwise pollute the listing.
var nonTemplateDirs = map[string]struct{}{
	".github":      {},
	".vscode":      {},
	"docs":         {},
	"node_modules": {},
}

// filterListableDirs drops repo-root entries that are conventionally
// not templates: dot-prefixed dirs, underscore-prefixed dirs, and a
// small denylist of common non-template folders.
func filterListableDirs(dirs []string) []string {
	out := make([]string, 0, len(dirs))
	for _, d := range dirs {
		if strings.HasPrefix(d, ".") || strings.HasPrefix(d, "_") {
			continue
		}
		if _, skip := nonTemplateDirs[d]; skip {
			continue
		}
		out = append(out, d)
	}
	return out
}

func fetchRepoTopLevelDirs(owner, repo string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents", owner, repo)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("list contents of %s/%s: %w", owner, repo, err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, fmt.Errorf("repository %s/%s not found", owner, repo)
	case http.StatusForbidden:
		return nil, fmt.Errorf("list contents of %s/%s: GitHub API rate limit reached (HTTP 403)", owner, repo)
	default:
		return nil, fmt.Errorf("list contents of %s/%s: HTTP %d", owner, repo, resp.StatusCode)
	}

	var items []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("parse contents response for %s/%s: %w", owner, repo, err)
	}

	var dirs []string
	for _, it := range items {
		if it.Type == "dir" {
			dirs = append(dirs, it.Name)
		}
	}
	return dirs, nil
}

// fetchTemplateDescription returns the description from a template's
// terox.json. A missing manifest is not an error — the template simply
// has no description.
func fetchTemplateDescription(owner, repo, dir string) (string, error) {
	url := fmt.Sprintf(
		"https://raw.githubusercontent.com/%s/%s/HEAD/%s/%s",
		owner, repo, dir, ManifestFilename,
	)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetch manifest for %s: %w", dir, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch manifest for %s: HTTP %d", dir, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read manifest for %s: %w", dir, err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return "", fmt.Errorf("parse manifest for %s: %w", dir, err)
	}
	return m.Description, nil
}

func printCatalogue(w io.Writer, ref string, entries []catalogueEntry) {
	_, _ = fmt.Fprintf(w, "%s:\n", ref)
	tw := tabwriter.NewWriter(w, 2, 2, 2, ' ', 0)
	for _, e := range entries {
		if e.Description != "" {
			_, _ = fmt.Fprintf(tw, "  %s\t%s\n", e.Name, e.Description)
		} else {
			_, _ = fmt.Fprintf(tw, "  %s\t\n", e.Name)
		}
	}
	_ = tw.Flush()
}
