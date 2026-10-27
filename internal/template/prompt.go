package template

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
)

type PromptOptions struct {
	Preset         map[string]string
	NonInteractive bool
	// Now is the time used for built-in default variables such as .Year.
	// The zero value means time.Now().
	Now time.Time
}

// defaultBuiltins returns the values a manifest default can reference, e.g.
// "default": "{{ .Year }}". Defaults are rendered against these only, not
// against other answers, so interactive and non-interactive runs agree.
func defaultBuiltins(now time.Time) map[string]string {
	return map[string]string{
		"Year": strconv.Itoa(now.Year()),
	}
}

func Prompt(m *Manifest, opts PromptOptions) (map[string]string, error) {
	answers := make(map[string]string, len(m.Variables))
	for k, v := range opts.Preset {
		answers[k] = v
	}

	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	builtins := defaultBuiltins(now)

	stringPtrs := make(map[string]*string)
	boolPtrs := make(map[string]*bool)
	var fields []huh.Field

	for i := range m.Variables {
		v := m.Variables[i]
		if _, set := answers[v.Name]; set {
			continue
		}

		def, err := renderString(v.Default, builtins)
		if err != nil {
			return nil, fmt.Errorf("render default for %q: %w", v.Name, err)
		}
		v.Default = def

		if opts.NonInteractive {
			if v.Default == "" && v.Type != "bool" && len(v.Choices) == 0 {
				return nil, fmt.Errorf(
					"variable %q has no value (--set %s=...) and no default",
					v.Name, v.Name,
				)
			}
			answers[v.Name] = v.Default
			continue
		}

		label := v.Prompt
		if label == "" {
			label = v.Name
		}

		switch {
		case len(v.Choices) > 0:
			val := v.Default
			if val == "" {
				val = v.Choices[0]
			}
			stringPtrs[v.Name] = &val
			huhOpts := make([]huh.Option[string], 0, len(v.Choices))
			for _, c := range v.Choices {
				huhOpts = append(huhOpts, huh.NewOption(c, c))
			}
			fields = append(fields,
				huh.NewSelect[string]().
					Title(label).
					Options(huhOpts...).
					Value(stringPtrs[v.Name]),
			)
		case v.Type == "bool":
			def := strings.EqualFold(v.Default, "true")
			b := def
			boolPtrs[v.Name] = &b
			fields = append(fields,
				huh.NewConfirm().
					Title(label).
					Value(boolPtrs[v.Name]),
			)
		default:
			s := v.Default
			stringPtrs[v.Name] = &s
			fields = append(fields,
				huh.NewInput().
					Title(label).
					Value(stringPtrs[v.Name]),
			)
		}
	}

	if len(fields) > 0 {
		form := huh.NewForm(huh.NewGroup(fields...))
		if err := form.Run(); err != nil {
			return nil, fmt.Errorf("prompt cancelled: %w", err)
		}
	}

	for name, p := range stringPtrs {
		answers[name] = *p
	}
	for name, p := range boolPtrs {
		if *p {
			answers[name] = "true"
		} else {
			answers[name] = "false"
		}
	}

	return answers, nil
}
