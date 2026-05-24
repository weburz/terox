package template

import (
	"bytes"
	"crypto/md5"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderString_NoTemplateShortcut(t *testing.T) {
	out, err := renderString("plain text", nil)
	if err != nil {
		t.Fatalf("renderString: %v", err)
	}
	if out != "plain text" {
		t.Errorf("got %q want %q", out, "plain text")
	}
}

func TestRenderString_SimpleSubstitution(t *testing.T) {
	out, err := renderString("hello {{.name}}!", map[string]string{"name": "world"})
	if err != nil {
		t.Fatalf("renderString: %v", err)
	}
	if out != "hello world!" {
		t.Errorf("got %q want %q", out, "hello world!")
	}
}

func TestRenderString_MissingKeyRendersEmpty(t *testing.T) {
	out, err := renderString("hi {{.absent}}!", map[string]string{})
	if err != nil {
		t.Fatalf("renderString: %v", err)
	}
	if out != "hi !" {
		t.Errorf("missingkey=zero should produce empty, got %q", out)
	}
}

func TestRenderString_ParseError(t *testing.T) {
	_, err := renderString("{{.broken", map[string]string{})
	if err == nil {
		t.Fatal("expected parse error for unterminated action")
	}
}

func TestRenderPath_SubstitutesEachSegment(t *testing.T) {
	got, err := renderPath("{{.a}}/{{.b}}/file.txt", map[string]string{"a": "x", "b": "y"})
	if err != nil {
		t.Fatalf("renderPath: %v", err)
	}
	want := filepath.Join("x", "y", "file.txt")
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestRenderPath_EmptySegmentSkipsWholePath(t *testing.T) {
	got, err := renderPath(`{{if eq .x "yes"}}stuff{{end}}/file.txt`, map[string]string{"x": "no"})
	if err != nil {
		t.Fatalf("renderPath: %v", err)
	}
	if got != "" {
		t.Errorf("empty segment should yield empty path, got %q", got)
	}
}

func TestIsBinary(t *testing.T) {
	if isBinary([]byte("plain text only")) {
		t.Error("plain text should not be binary")
	}
	if !isBinary([]byte{'a', 'b', 0, 'c'}) {
		t.Error("buffer with null byte should be binary")
	}
	if isBinary([]byte{}) {
		t.Error("empty buffer should not be binary")
	}
}

// writeFile is a tiny helper for the end-to-end fixtures.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestRender_EndToEnd(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, ManifestFilename), `{"variables":[{"name":"name"}]}`)
	writeFile(t, filepath.Join(src, "{{.name}}", "README.md"), "# {{.name}}\n")
	writeFile(t, filepath.Join(src, "{{.name}}", "static.txt"), "no vars here\n")

	vars := map[string]string{"name": "cool"}
	if err := Render(src, dst, vars); err != nil {
		t.Fatalf("Render: %v", err)
	}

	// Manifest must NOT appear in output.
	if _, err := os.Stat(filepath.Join(dst, ManifestFilename)); !os.IsNotExist(err) {
		t.Errorf("manifest should be skipped from output")
	}

	got, err := os.ReadFile(filepath.Join(dst, "cool", "README.md"))
	if err != nil {
		t.Fatalf("read rendered README: %v", err)
	}
	if string(got) != "# cool\n" {
		t.Errorf("README contents: got %q want %q", got, "# cool\n")
	}

	got2, err := os.ReadFile(filepath.Join(dst, "cool", "static.txt"))
	if err != nil {
		t.Fatalf("read static.txt: %v", err)
	}
	if string(got2) != "no vars here\n" {
		t.Errorf("static.txt should pass through unchanged, got %q", got2)
	}
}

func TestRender_BinaryFilePassthrough(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, ManifestFilename), `{"variables":[{"name":"name"}]}`)

	binPath := filepath.Join(src, "{{.name}}", "image.bin")
	binBody := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d}
	// Append something that LOOKS like a template marker, to prove it stays raw.
	binBody = append(binBody, []byte("{{.name}}")...)
	if err := os.MkdirAll(filepath.Dir(binPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binPath, binBody, 0o644); err != nil {
		t.Fatal(err)
	}
	srcSum := md5.Sum(binBody)

	if err := Render(src, dst, map[string]string{"name": "out"}); err != nil {
		t.Fatalf("Render: %v", err)
	}

	gotBody, err := os.ReadFile(filepath.Join(dst, "out", "image.bin"))
	if err != nil {
		t.Fatalf("read rendered binary: %v", err)
	}
	dstSum := md5.Sum(gotBody)
	if !bytes.Equal(srcSum[:], dstSum[:]) {
		t.Errorf("binary md5 differs: src=%x dst=%x", srcSum, dstSum)
	}
	if !bytes.Contains(gotBody, []byte("{{.name}}")) {
		t.Error("binary marker {{.name}} should be preserved verbatim, not rendered")
	}
}

func TestCopy_VerbatimAndSkipsManifest(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, ManifestFilename), `{}`)
	writeFile(t, filepath.Join(src, "literal.txt"), "keep {{.unrendered}} as-is\n")
	writeFile(t, filepath.Join(src, "sub", "nested.txt"), "nested\n")

	if err := Copy(src, dst); err != nil {
		t.Fatalf("Copy: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dst, ManifestFilename)); !os.IsNotExist(err) {
		t.Errorf("manifest should be skipped in Copy too")
	}

	got, err := os.ReadFile(filepath.Join(dst, "literal.txt"))
	if err != nil {
		t.Fatalf("read literal: %v", err)
	}
	if !strings.Contains(string(got), "{{.unrendered}}") {
		t.Errorf("Copy should not render: got %q", got)
	}

	got2, err := os.ReadFile(filepath.Join(dst, "sub", "nested.txt"))
	if err != nil {
		t.Fatalf("read nested: %v", err)
	}
	if string(got2) != "nested\n" {
		t.Errorf("nested contents: got %q", got2)
	}
}
