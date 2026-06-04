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
templates available at the root of that repository. Each top-level
directory is treated as a candidate template; if a directory contains a
terox.json with a "description" field, the description is shown next to
the name.
`

var listExample = `  terox list
  terox list weburz/terox-templates
  terox ls weburz/terox-templates`

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
		return template.ListRemote(args[0])
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
