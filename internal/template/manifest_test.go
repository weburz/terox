package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeManifest(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ManifestFilename), []byte(content), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}

func TestLoadManifest_MissingFileReturnsNilNil(t *testing.T) {
	m, err := LoadManifest(t.TempDir())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if m != nil {
		t.Fatalf("expected nil manifest, got %+v", m)
	}
}

func TestLoadManifest_Valid(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, `{
		"name": "Example",
		"description": "An example",
		"variables": [
			{"name": "project_name", "prompt": "Name", "default": "x"},
			{"name": "license", "choices": ["MIT", "Apache-2.0"], "default": "MIT"}
		]
	}`)

	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if m == nil {
		t.Fatal("expected non-nil manifest")
	}
	if m.Name != "Example" {
		t.Errorf("Name: got %q want %q", m.Name, "Example")
	}
	if len(m.Variables) != 2 {
		t.Fatalf("Variables len: got %d want 2", len(m.Variables))
	}
	if m.Variables[0].Name != "project_name" || m.Variables[0].Default != "x" {
		t.Errorf("first variable wrong: %+v", m.Variables[0])
	}
	if len(m.Variables[1].Choices) != 2 || m.Variables[1].Choices[0] != "MIT" {
		t.Errorf("choices wrong: %+v", m.Variables[1])
	}
}

func TestLoadManifest_MalformedJSON(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, `{ not valid json`)
	_, err := LoadManifest(dir)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("expected parse error, got: %v", err)
	}
}

func TestLoadManifest_VariableMissingName(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, `{"variables":[{"prompt":"x","default":"y"}]}`)
	_, err := LoadManifest(dir)
	if err == nil {
		t.Fatal("expected error for variable missing name")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error should mention missing name, got: %v", err)
	}
}

func TestLoadManifest_EmptyVariables(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, `{"name":"x"}`)
	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if len(m.Variables) != 0 {
		t.Errorf("expected 0 variables, got %d", len(m.Variables))
	}
}
