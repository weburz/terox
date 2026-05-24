/**
 * Package cmd - The "cmd" package contains the logic to handle various
 * commands passed to the CLI application.
 *
 * The "root" file in particular which is part of the "cmd" package contains
 * the simple logic to handle to CLI application itself!
 */
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Core logic to handle the CLI application itself (known as "root" throughout
// the project's source code).
var rootCmd = &cobra.Command{
	Use:           "terox",
	Short:         "A project generator built in Golang!",
	SilenceUsage:  true,
	SilenceErrors: false,
}

// Expose a function to execute the functions of the root command (the CLI
// application itself).
func Execute() {
	err := rootCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}
