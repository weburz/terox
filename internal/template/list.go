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
	"unicode/utf8"

	"github.com/charmbracelet/x/term"
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
// a subpath is not supported. A top-level directory is a template iff
// it contains a terox.json; the manifest's description, if set, is
// shown next to the name.
//
// When long is true, each template renders as a name line followed by
// its description wrapped under it. Otherwise the output is a compact
// two-column table; descriptions are truncated with an ellipsis to fit
// the terminal width when stdout is a TTY, and left untouched
// otherwise so piped output stays parseable.
func ListRemote(ref string, long bool) error {
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

	sort.Strings(dirs)
	entries := make([]catalogueEntry, 0, len(dirs))
	for _, d := range dirs {
		m, err := fetchTemplateManifest(owner, repo, d)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
			continue
		}
		if m == nil {
			continue
		}
		entries = append(entries, catalogueEntry{Name: d, Description: m.Description})
	}

	if len(entries) == 0 {
		fmt.Printf("No templates found in %s/%s.\n", owner, repo)
		return nil
	}

	width, isTTY := terminalWidth()
	if long {
		wrap := width
		if wrap <= 0 {
			wrap = 100
		}
		printCatalogueLong(os.Stdout, owner+"/"+repo, entries, wrap)
		return nil
	}

	truncWidth := 0
	if isTTY {
		truncWidth = width
	}
	printCatalogueCompact(os.Stdout, owner+"/"+repo, entries, truncWidth)
	return nil
}

type catalogueEntry struct {
	Name        string
	Description string
}

// filterListableDirs is a cheap pre-filter that drops dot- and
// underscore-prefixed dirs before we spend an HTTP roundtrip on each
// candidate. The authoritative check is the manifest fetch in
// ListRemote: a dir is a template iff it has a terox.json.
func filterListableDirs(dirs []string) []string {
	out := make([]string, 0, len(dirs))
	for _, d := range dirs {
		if strings.HasPrefix(d, ".") || strings.HasPrefix(d, "_") {
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

// fetchTemplateManifest returns the parsed terox.json for a candidate
// template directory, or nil if the directory has no manifest (and
// therefore isn't a template at all).
func fetchTemplateManifest(owner, repo, dir string) (*Manifest, error) {
	url := fmt.Sprintf(
		"https://raw.githubusercontent.com/%s/%s/HEAD/%s/%s",
		owner, repo, dir, ManifestFilename,
	)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch manifest for %s: %w", dir, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch manifest for %s: HTTP %d", dir, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read manifest for %s: %w", dir, err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest for %s: %w", dir, err)
	}
	return &m, nil
}

// terminalWidth returns the column width of stdout when it is a TTY.
// (0, false) means stdout is not a terminal (piped or redirected), in
// which case callers should skip truncation to keep the output
// parseable.
func terminalWidth() (int, bool) {
	fd := os.Stdout.Fd()
	if !term.IsTerminal(fd) {
		return 0, false
	}
	w, _, err := term.GetSize(fd)
	if err != nil || w <= 0 {
		return 0, false
	}
	return w, true
}

// printCatalogueCompact renders entries as a two-column table aligned
// on the name column. If maxWidth > 0, lines longer than maxWidth are
// truncated with an ellipsis; maxWidth == 0 disables truncation (used
// for piped output).
func printCatalogueCompact(w io.Writer, ref string, entries []catalogueEntry, maxWidth int) {
	_, _ = fmt.Fprintf(w, "%s:\n", ref)

	nameWidth := 0
	for _, e := range entries {
		if n := utf8.RuneCountInString(e.Name); n > nameWidth {
			nameWidth = n
		}
	}

	for _, e := range entries {
		var line string
		if e.Description != "" {
			line = fmt.Sprintf("  %-*s  %s", nameWidth, e.Name, e.Description)
		} else {
			line = fmt.Sprintf("  %s", e.Name)
		}
		if maxWidth > 0 {
			line = truncateToWidth(line, maxWidth)
		}
		_, _ = fmt.Fprintln(w, line)
	}
}

// printCatalogueLong renders each entry as a name line followed by its
// description wrapped under it. wrapWidth is the absolute column
// budget; the indent is subtracted internally.
func printCatalogueLong(w io.Writer, ref string, entries []catalogueEntry, wrapWidth int) {
	_, _ = fmt.Fprintf(w, "%s:\n", ref)

	const indent = "    "
	body := wrapWidth - len(indent)
	if body < 20 {
		body = 20
	}

	for _, e := range entries {
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintf(w, "  %s\n", e.Name)
		if e.Description == "" {
			continue
		}
		for _, line := range wordWrap(e.Description, body) {
			_, _ = fmt.Fprintf(w, "%s%s\n", indent, line)
		}
	}
}

// truncateToWidth returns s shortened to at most max display columns,
// appending an ellipsis when truncation occurs. Counts runes rather
// than bytes; multibyte characters are treated as one column each,
// which is correct for ASCII and "close enough" for typical Latin text
// in template descriptions.
func truncateToWidth(s string, max int) string {
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	runes := []rune(s)
	return strings.TrimRight(string(runes[:max-1]), " ") + "…"
}

// wordWrap breaks s into lines of at most width runes, splitting on
// whitespace. Words longer than width are emitted on their own line
// rather than broken mid-character — better to overflow occasionally
// than to mangle a URL or identifier.
func wordWrap(s string, width int) []string {
	if width <= 0 {
		return []string{s}
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	var cur strings.Builder
	for _, word := range words {
		switch {
		case cur.Len() == 0:
			cur.WriteString(word)
		case utf8.RuneCountInString(cur.String())+1+utf8.RuneCountInString(word) <= width:
			cur.WriteByte(' ')
			cur.WriteString(word)
		default:
			lines = append(lines, cur.String())
			cur.Reset()
			cur.WriteString(word)
		}
	}
	if cur.Len() > 0 {
		lines = append(lines, cur.String())
	}
	return lines
}
