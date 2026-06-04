/**
 * Package cmd - The "cmd" package contains the logic to handle various
 * commands passed to the CLI tool.
 *
 * The "list" file in particular contains the logic to list templates,
 * either from the local cache or from a remote GitHub repository.
 */
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/weburz/terox/internal/template"
)

var listCmdShortHelp = "List available templates (locally cached or in a remote repo)."

var listCmdLongHelp = `
List available templates.

With no arguments, prints the templates already downloaded into the local
cache at "${XDG_DATA_HOME}/terox" (typically $HOME/.local/share/terox).

Given a GitHub repository ref of the form "owner/repo", prints the
templates available at the root of that repository. A top-level
directory is treated as a template iff it contains a terox.json
manifest. The manifest's "description" field, if set, is shown next to
the name.

By default, descriptions are truncated to fit the terminal width. Pass
--long (-l) to render each template's full description, wrapped under
its name. When stdout is not a terminal (piped or redirected), the
compact layout is preserved but truncation is skipped so the output
stays parseable.
`

var listExample = `  terox list
  terox list weburz/terox-templates
  terox list -l weburz/terox-templates
  terox ls weburz/terox-templates`

var listLong bool

var listCmd = &cobra.Command{
	Use:     "list [owner/repo]",
	Aliases: []string{"ls", "show"},
	Short:   listCmdShortHelp,
	Example: listExample,
	Long:    listCmdLongHelp,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return template.List()
		}
		return template.ListRemote(args[0], listLong)
	},
}

func init() {
	listCmd.Flags().BoolVarP(&listLong, "long", "l", false, "Show full descriptions wrapped under each template name.")
	rootCmd.AddCommand(listCmd)
}
