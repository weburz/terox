package template

import (
	"strings"
	"testing"
	"time"
)

func TestPrompt_NonInteractive_UsesDefaults(t *testing.T) {
	m := &Manifest{Variables: []Variable{
		{Name: "name", Default: "alice"},
		{Name: "license", Default: "MIT"},
	}}
	answers, err := Prompt(m, PromptOptions{NonInteractive: true})
	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}
	if answers["name"] != "alice" || answers["license"] != "MIT" {
		t.Errorf("defaults not applied: %+v", answers)
	}
}

func TestPrompt_NonInteractive_PresetOverridesDefault(t *testing.T) {
	m := &Manifest{Variables: []Variable{
		{Name: "name", Default: "alice"},
	}}
	answers, err := Prompt(m, PromptOptions{
		NonInteractive: true,
		Preset:         map[string]string{"name": "bob"},
	})
	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}
	if answers["name"] != "bob" {
		t.Errorf("preset should override default, got %q", answers["name"])
	}
}

func TestPrompt_NonInteractive_MissingRequiredErrors(t *testing.T) {
	m := &Manifest{Variables: []Variable{
		{Name: "required_thing"}, // no default, no preset
	}}
	_, err := Prompt(m, PromptOptions{NonInteractive: true})
	if err == nil {
		t.Fatal("expected error for missing required variable")
	}
	if !strings.Contains(err.Error(), "required_thing") {
		t.Errorf("error should name the missing variable, got: %v", err)
	}
}

func TestPrompt_NonInteractive_BoolWithoutDefault(t *testing.T) {
	// Booleans should resolve to empty (rendered "" by current code path)
	// even when no default is provided, because Type=="bool" is short-circuited
	// past the "needs a value" check.
	m := &Manifest{Variables: []Variable{
		{Name: "flag", Type: "bool"},
	}}
	answers, err := Prompt(m, PromptOptions{NonInteractive: true})
	if err != nil {
		t.Fatalf("Prompt should accept bool with no default: %v", err)
	}
	if _, ok := answers["flag"]; !ok {
		t.Errorf("flag should be set, got %+v", answers)
	}
}

func TestPrompt_NonInteractive_ChoicesWithoutDefault(t *testing.T) {
	// Choice variables with no default should not error in non-interactive;
	// the empty default is acceptable (callers can --set to override).
	m := &Manifest{Variables: []Variable{
		{Name: "license", Choices: []string{"MIT", "Apache-2.0"}},
	}}
	answers, err := Prompt(m, PromptOptions{NonInteractive: true})
	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}
	if _, ok := answers["license"]; !ok {
		t.Errorf("license should be set (even if empty), got %+v", answers)
	}
}

func TestPrompt_NoVariables(t *testing.T) {
	m := &Manifest{}
	answers, err := Prompt(m, PromptOptions{NonInteractive: true})
	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}
	if len(answers) != 0 {
		t.Errorf("expected empty answers, got %+v", answers)
	}
}

func TestPrompt_NonInteractive_DefaultRendersBuiltins(t *testing.T) {
	m := &Manifest{Variables: []Variable{
		{Name: "LicenseYear", Default: "{{ .Year }}"},
		{Name: "Copyright", Default: "(c) {{ .Year }} Weburz"},
	}}
	answers, err := Prompt(m, PromptOptions{
		NonInteractive: true,
		Now:            time.Date(2031, time.March, 4, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}
	if answers["LicenseYear"] != "2031" {
		t.Errorf("LicenseYear = %q, want %q", answers["LicenseYear"], "2031")
	}
	if answers["Copyright"] != "(c) 2031 Weburz" {
		t.Errorf("Copyright = %q, want %q", answers["Copyright"], "(c) 2031 Weburz")
	}
}

func TestPrompt_NonInteractive_PresetIsNotRendered(t *testing.T) {
	// Values passed via --set are taken literally, even if they look like
	// template syntax; only manifest defaults are rendered.
	m := &Manifest{Variables: []Variable{
		{Name: "LicenseYear", Default: "{{ .Year }}"},
	}}
	answers, err := Prompt(m, PromptOptions{
		NonInteractive: true,
		Preset:         map[string]string{"LicenseYear": "{{ .Year }}"},
	})
	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}
	if answers["LicenseYear"] != "{{ .Year }}" {
		t.Errorf("preset should be kept verbatim, got %q", answers["LicenseYear"])
	}
}

func TestPrompt_InvalidDefaultTemplateErrors(t *testing.T) {
	m := &Manifest{Variables: []Variable{
		{Name: "LicenseYear", Default: "{{ .Year"},
	}}
	_, err := Prompt(m, PromptOptions{NonInteractive: true})
	if err == nil {
		t.Fatal("expected error for unparsable default")
	}
	if !strings.Contains(err.Error(), "LicenseYear") {
		t.Errorf("error should name the variable, got: %v", err)
	}
}
