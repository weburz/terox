package template

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestFilterListableDirs(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "passes plain template dirs",
			in:   []string{"npm-package", "nuxt-module", "simple-website"},
			want: []string{"npm-package", "nuxt-module", "simple-website"},
		},
		{
			name: "drops dot-prefixed and underscore-prefixed",
			in:   []string{"a", ".github", "_meta", "b"},
			want: []string{"a", "b"},
		},
		{
			name: "empty input gives empty output",
			in:   nil,
			want: []string{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := filterListableDirs(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestPrintCatalogueCompact_AlignsAndPreservesDescriptions(t *testing.T) {
	var buf bytes.Buffer
	printCatalogueCompact(&buf, "weburz/terox-templates", []catalogueEntry{
		{Name: "npm-package", Description: "TS lib starter"},
		{Name: "nuxt-module", Description: "Nuxt 4 module starter"},
		{Name: "no-desc", Description: ""},
	}, 0) // 0 disables truncation

	out := buf.String()

	if !strings.HasPrefix(out, "weburz/terox-templates:\n") {
		t.Errorf("missing header in output:\n%s", out)
	}
	for _, want := range []string{"npm-package", "TS lib starter", "nuxt-module", "Nuxt 4 module starter", "no-desc"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}

	// Verify the name column is padded so descriptions align.
	padAfter := func(name string) int {
		i := strings.Index(out, name)
		if i < 0 {
			t.Fatalf("name %q not found in output", name)
		}
		rest := out[i+len(name):]
		end := strings.IndexFunc(rest, func(r rune) bool { return r != ' ' })
		return end
	}
	if padAfter("npm-package") != padAfter("nuxt-module") {
		t.Errorf("name column not aligned between entries:\n%s", out)
	}
}

func TestPrintCatalogueCompact_TruncatesWhenWidthSet(t *testing.T) {
	var buf bytes.Buffer
	printCatalogueCompact(&buf, "weburz/terox-templates", []catalogueEntry{
		{Name: "docs-site", Description: "Weburz-flavored Nuxt 4 documentation site with content schema, ESLint, deploy workflow."},
	}, 40)

	out := buf.String()
	if !strings.Contains(out, "…") {
		t.Errorf("expected ellipsis in truncated output, got:\n%s", out)
	}
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if n := len([]rune(line)); n > 40 {
			t.Errorf("line of width %d exceeds max 40: %q", n, line)
		}
	}
}

func TestPrintCatalogueCompact_NoTruncationWhenWidthZero(t *testing.T) {
	long := strings.Repeat("x ", 60) + "end"
	var buf bytes.Buffer
	printCatalogueCompact(&buf, "ref", []catalogueEntry{
		{Name: "n", Description: long},
	}, 0)

	if !strings.Contains(buf.String(), "end") {
		t.Errorf("description was truncated when maxWidth==0:\n%s", buf.String())
	}
	if strings.Contains(buf.String(), "…") {
		t.Errorf("unexpected ellipsis when maxWidth==0:\n%s", buf.String())
	}
}

func TestPrintCatalogueLong_WrapsDescriptionUnderName(t *testing.T) {
	var buf bytes.Buffer
	printCatalogueLong(&buf, "weburz/terox-templates", []catalogueEntry{
		{Name: "docs-site", Description: "Weburz-flavored Nuxt 4 documentation site with content schema and deploy workflow."},
		{Name: "bare", Description: ""},
	}, 40)

	out := buf.String()

	if !strings.HasPrefix(out, "weburz/terox-templates:\n") {
		t.Errorf("missing header in output:\n%s", out)
	}
	if !strings.Contains(out, "  docs-site\n") {
		t.Errorf("name should be on its own line:\n%s", out)
	}
	if !strings.Contains(out, "    Weburz-flavored") {
		t.Errorf("description should be indented under the name:\n%s", out)
	}

	// Body lines must respect the wrap width (40 - 4 indent = 36).
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "    ") {
			body := strings.TrimPrefix(line, "    ")
			if n := len([]rune(body)); n > 36 {
				t.Errorf("wrapped line exceeds 36 cols (%d): %q", n, body)
			}
		}
	}

	// An entry with no description should still print its name.
	if !strings.Contains(out, "  bare\n") {
		t.Errorf("empty-description entry should still print its name:\n%s", out)
	}
}

func TestTruncateToWidth(t *testing.T) {
	tests := []struct {
		in   string
		max  int
		want string
	}{
		{"hello world", 11, "hello world"},
		{"hello world", 5, "hell…"},
		{"hello", 0, "hello"},
		{"hello", 1, "…"},
		// Trailing space before ellipsis should be trimmed.
		{"hello world", 7, "hello…"},
		{"hello there friend", 7, "hello…"},
	}
	for _, tc := range tests {
		got := truncateToWidth(tc.in, tc.max)
		if got != tc.want {
			t.Errorf("truncateToWidth(%q, %d) = %q, want %q", tc.in, tc.max, got, tc.want)
		}
	}
}

func TestWordWrap(t *testing.T) {
	tests := []struct {
		in    string
		width int
		want  []string
	}{
		{"one two three", 20, []string{"one two three"}},
		{"one two three four", 7, []string{"one two", "three", "four"}},
		{"", 10, nil},
		{"supercalifragilistic short", 10, []string{"supercalifragilistic", "short"}},
		{"a b c d e", 1, []string{"a", "b", "c", "d", "e"}},
	}
	for _, tc := range tests {
		got := wordWrap(tc.in, tc.width)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("wordWrap(%q, %d) = %q, want %q", tc.in, tc.width, got, tc.want)
		}
	}
}

func TestListRemote_RejectsSubpath(t *testing.T) {
	err := ListRemote("weburz/terox-templates/npm-package", false)
	if err == nil {
		t.Fatal("expected error when ref includes a subpath")
	}
	if !strings.Contains(err.Error(), "subpath") {
		t.Errorf("error should mention subpath, got %q", err.Error())
	}
}

func TestListRemote_RejectsInvalidRef(t *testing.T) {
	err := ListRemote("nofslash", false)
	if err == nil {
		t.Fatal("expected error for malformed ref")
	}
}
