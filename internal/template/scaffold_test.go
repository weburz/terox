package template

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestSplitRef(t *testing.T) {
	tests := []struct {
		in          string
		wantOwner   string
		wantRepo    string
		wantSubpath string
		wantErr     bool
	}{
		{"foo/bar", "foo", "bar", "", false},
		{"Weburz/simple-website-template", "Weburz", "simple-website-template", "", false},
		{"weburz/terox-templates/nuxt-starter", "weburz", "terox-templates", "nuxt-starter", false},
		{"weburz/terox-templates/group/nested", "weburz", "terox-templates", "group/nested", false},
		{"", "", "", "", true},
		{"nofslash", "", "", "", true},
		{"/bar", "", "", "", true},
		{"foo/", "", "", "", true},
		{"/", "", "", "", true},
		{"foo/bar/", "", "", "", true},       // trailing slash → empty subpath
		{"foo/bar//baz", "", "", "", true},   // empty segment in subpath
		{"foo/bar/..", "", "", "", true},     // escape attempt
		{"foo/bar/../etc", "", "", "", true}, // escape attempt
		{"foo/bar/./baz", "", "", "", true},  // dot segment
		{"foo/bar//", "", "", "", true},      // empty segment
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			owner, repo, subpath, err := splitRef(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err: got %v wantErr=%v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if owner != tc.wantOwner || repo != tc.wantRepo || subpath != tc.wantSubpath {
				t.Errorf("got (%q, %q, %q) want (%q, %q, %q)",
					owner, repo, subpath, tc.wantOwner, tc.wantRepo, tc.wantSubpath)
			}
		})
	}
}

func TestResolveSubpath(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "nuxt-starter", "pages"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "loose-file.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	t.Run("empty subpath returns root", func(t *testing.T) {
		got, err := resolveSubpath(root, "", "owner/repo")
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if got != root {
			t.Errorf("got %q want %q", got, root)
		}
	})

	t.Run("existing directory resolves", func(t *testing.T) {
		got, err := resolveSubpath(root, "nuxt-starter", "owner/repo")
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		want := filepath.Join(root, "nuxt-starter")
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("nested directory resolves", func(t *testing.T) {
		got, err := resolveSubpath(root, "nuxt-starter/pages", "owner/repo")
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		want := filepath.Join(root, "nuxt-starter", "pages")
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("missing subpath errors", func(t *testing.T) {
		if _, err := resolveSubpath(root, "does-not-exist", "owner/repo"); err == nil {
			t.Fatal("expected error for missing subpath")
		}
	})

	t.Run("file subpath errors", func(t *testing.T) {
		if _, err := resolveSubpath(root, "loose-file.txt", "owner/repo"); err == nil {
			t.Fatal("expected error when subpath points to a file")
		}
	})
}

// makeFakeZip builds a zip that mimics a GitHub zipball: one top-level
// directory named `topFolder` containing the supplied path→contents map.
func makeFakeZip(t *testing.T, topFolder string, files map[string]string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "fake-*.zip")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	w := zip.NewWriter(f)
	for relPath, content := range files {
		fw, err := w.Create(topFolder + "/" + relPath)
		if err != nil {
			t.Fatalf("zip.Create: %v", err)
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			t.Fatalf("zip.Write: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("zip.Close: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("file.Close: %v", err)
	}
	return f.Name()
}

func TestExtractTemplate_StripsTopLevelFolder(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "weburz", "crisp")
	// GitHub's zipballs name the top folder like "weburz-crisp-abc123def".
	zipPath := makeFakeZip(t, "weburz-crisp-abc123def", map[string]string{
		"README.md":      "hello\n",
		"sub/nested.txt": "nested body\n",
	})

	if err := extractTemplate(zipPath, dest); err != nil {
		t.Fatalf("extractTemplate: %v", err)
	}

	readme, err := os.ReadFile(filepath.Join(dest, "README.md"))
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	if string(readme) != "hello\n" {
		t.Errorf("README content: got %q", readme)
	}

	nested, err := os.ReadFile(filepath.Join(dest, "sub", "nested.txt"))
	if err != nil {
		t.Fatalf("read nested: %v", err)
	}
	if string(nested) != "nested body\n" {
		t.Errorf("nested content: got %q", nested)
	}
}

func TestExtractTemplate_HyphenatedRepoName(t *testing.T) {
	// Regression: previously the destination was guessed by parsing the
	// top-level folder name with strings.SplitN("-", 3), which mapped
	// "weburz-terox-templates-abc123" to owner=weburz,repo=terox. The
	// caller now controls the destination, so hyphenated repo names
	// extract correctly.
	dest := filepath.Join(t.TempDir(), "weburz", "terox-templates")
	zipPath := makeFakeZip(t, "weburz-terox-templates-abc123def", map[string]string{
		"npm-package/terox.json": `{"name":"npm-package"}`,
		"nuxt-module/terox.json": `{"name":"nuxt-module"}`,
		"README.md":              "monorepo readme\n",
	})

	if err := extractTemplate(zipPath, dest); err != nil {
		t.Fatalf("extractTemplate: %v", err)
	}

	for _, p := range []string{
		"README.md",
		filepath.Join("npm-package", "terox.json"),
		filepath.Join("nuxt-module", "terox.json"),
	} {
		if _, err := os.Stat(filepath.Join(dest, p)); err != nil {
			t.Errorf("expected %s to exist after extract: %v", p, err)
		}
	}
}
