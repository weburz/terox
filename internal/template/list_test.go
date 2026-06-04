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
			name: "drops common non-template dirs",
			in:   []string{"docs", "node_modules", ".vscode", "nuxt-module"},
			want: []string{"nuxt-module"},
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

func TestPrintCatalogue(t *testing.T) {
	var buf bytes.Buffer
	printCatalogue(&buf, "weburz/terox-templates", []catalogueEntry{
		{Name: "npm-package", Description: "TS lib starter"},
		{Name: "nuxt-module", Description: "Nuxt 4 module starter"},
		{Name: "no-desc", Description: ""},
	})

	out := buf.String()

	if !strings.HasPrefix(out, "weburz/terox-templates:\n") {
		t.Errorf("missing header in output:\n%s", out)
	}
	for _, want := range []string{"npm-package", "TS lib starter", "nuxt-module", "Nuxt 4 module starter", "no-desc"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}

	// Verify the name column is padded so descriptions align: the
	// substring between "npm-package" and "TS lib" must be the same
	// width as the substring between "nuxt-module" and "Nuxt 4".
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

func TestListRemote_RejectsSubpath(t *testing.T) {
	err := ListRemote("weburz/terox-templates/npm-package")
	if err == nil {
		t.Fatal("expected error when ref includes a subpath")
	}
	if !strings.Contains(err.Error(), "subpath") {
		t.Errorf("error should mention subpath, got %q", err.Error())
	}
}

func TestListRemote_RejectsInvalidRef(t *testing.T) {
	err := ListRemote("nofslash")
	if err == nil {
		t.Fatal("expected error for malformed ref")
	}
}
