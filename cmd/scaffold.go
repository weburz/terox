package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/weburz/terox/internal/template"
)

var (
	scaffoldOutput         string
	scaffoldSets           []string
	scaffoldNonInteractive bool
	scaffoldRefresh        bool
)

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold <owner/repo|path>",
	Short: "Scaffold a project from a template",
	Long: `Scaffold a project from a template.

The template can be a GitHub repository (owner/repo) or a local directory path.
If the template contains a terox.json manifest, you will be prompted for the
declared variables and the template will be rendered into --output. If no
manifest is present, the template files are copied as-is.`,
	Example: `  terox scaffold Weburz/simple-website-template --output ./my-site
  terox scaffold ./my-local-template --output ./my-site
  terox scaffold Weburz/foo --set project_name=bar --set author=me --non-interactive`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]

		srcDir, err := resolveTemplate(ref, scaffoldRefresh)
		if err != nil {
			return err
		}

		manifest, err := template.LoadManifest(srcDir)
		if err != nil {
			return err
		}

		preset, err := parseSets(scaffoldSets)
		if err != nil {
			return err
		}

		if manifest == nil {
			fmt.Println("No terox.json found; copying template as-is.")
			if err := template.Copy(srcDir, scaffoldOutput); err != nil {
				return err
			}
		} else {
			vars, err := template.Prompt(manifest, template.PromptOptions{
				Preset:         preset,
				NonInteractive: scaffoldNonInteractive,
			})
			if err != nil {
				return err
			}
			if err := template.Render(srcDir, scaffoldOutput, vars); err != nil {
				return err
			}
		}

		fmt.Printf("Scaffolded into %s\n", scaffoldOutput)
		return nil
	},
}

func resolveTemplate(ref string, refresh bool) (string, error) {
	if info, err := os.Stat(ref); err == nil && info.IsDir() {
		if refresh {
			fmt.Println("--refresh has no effect for local templates; using folder as-is.")
		}
		return ref, nil
	}
	return template.Cache(ref, refresh)
}

func parseSets(sets []string) (map[string]string, error) {
	out := make(map[string]string, len(sets))
	for _, s := range sets {
		k, v, ok := strings.Cut(s, "=")
		if !ok || k == "" {
			return nil, fmt.Errorf("--set must be key=value, got %q", s)
		}
		out[k] = v
	}
	return out, nil
}

func init() {
	scaffoldCmd.Flags().StringVarP(
		&scaffoldOutput, "output", "o", ".",
		"directory to write the rendered project into",
	)
	scaffoldCmd.Flags().StringSliceVar(
		&scaffoldSets, "set", nil,
		"preset variable value (repeatable): --set name=value",
	)
	scaffoldCmd.Flags().BoolVar(
		&scaffoldNonInteractive, "non-interactive", false,
		"do not prompt; require all unset variables to have defaults",
	)
	scaffoldCmd.Flags().BoolVar(
		&scaffoldRefresh, "refresh", false,
		"re-download a remote template, ignoring the local cache",
	)
	rootCmd.AddCommand(scaffoldCmd)
}
