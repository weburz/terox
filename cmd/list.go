/**
 * Package cmd - The "cmd" package contains the logic to handle various
 * commands passed to the CLI tool.
 *
 * The "list" file in particular contains the logic to list all locally
 * available template(s) on disk.
 */
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/weburz/terox/internal/template"
)

var listCmdShortHelp = "List all locally available templates."

var listCmdLongHelp = `
List all the locally available templates.

This command will list all the available templates on your local system. The
default directory, Terox will check of available templates is at
"${XDG_DATA_HOME}/terox" wherein XDG_DATA_HOME is usually set to
$HOME/.local/share.
`

var listExample = "terox list\nterox ls\nterox show"

// Logic to handle the "list" command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "show"},
	Short:   listCmdShortHelp,
	Example: listExample,
	Long:    listCmdLongHelp,
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		template.List()
	},
}

// Register the logic above with the CLI application
func init() {
	rootCmd.AddCommand(listCmd)
}
