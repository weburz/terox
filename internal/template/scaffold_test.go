package template

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestSplitRepo(t *testing.T) {
	tests := []struct {
		in        string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{"foo/bar", "foo", "bar", false},
		{"Weburz/simple-website-template", "Weburz", "simple-website-template", false},
		{"foo/bar/baz", "foo", "bar/baz", false}, // SplitN keeps trailing slashes in repo
		{"", "", "", true},
		{"nofslash", "", "", true},
		{"/bar", "", "", true},
		{"foo/", "", "", true},
		{"/", "", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			owner, repo, err := splitRepo(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err: got %v wantErr=%v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if owner != tc.wantOwner || repo != tc.wantRepo {
				t.Errorf("got (%q, %q) want (%q, %q)", owner, repo, tc.wantOwner, tc.wantRepo)
			}
		})
	}
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

func TestExtractTemplate_LowercasesOwnerRepo(t *testing.T) {
	dest := t.TempDir()
	// GitHub's zipballs name the top folder like "weburz-crisp-abc123def".
	zipPath := makeFakeZip(t, "weburz-crisp-abc123def", map[string]string{
		"README.md":      "hello\n",
		"sub/nested.txt": "nested body\n",
	})

	finalDest, err := extractTemplate(zipPath, dest)
	if err != nil {
		t.Fatalf("extractTemplate: %v", err)
	}

	wantDest := filepath.Join(dest, "weburz", "crisp")
	if finalDest != wantDest {
		t.Errorf("finalDest: got %q want %q", finalDest, wantDest)
	}

	readme, err := os.ReadFile(filepath.Join(finalDest, "README.md"))
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	if string(readme) != "hello\n" {
		t.Errorf("README content: got %q", readme)
	}

	nested, err := os.ReadFile(filepath.Join(finalDest, "sub", "nested.txt"))
	if err != nil {
		t.Fatalf("read nested: %v", err)
	}
	if string(nested) != "nested body\n" {
		t.Errorf("nested content: got %q", nested)
	}
}

func TestExtractTemplate_RejectsBadTopFolder(t *testing.T) {
	dest := t.TempDir()
	// Top folder without the expected "owner-repo-sha" structure.
	zipPath := makeFakeZip(t, "weirdfolder", map[string]string{
		"file.txt": "hi\n",
	})
	if _, err := extractTemplate(zipPath, dest); err == nil {
		t.Fatal("expected error for malformed top folder")
	}
}
