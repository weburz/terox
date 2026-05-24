/**
 * Package cmd - The "cmd" package contains the logic to handle the commands
 * passed to the CLI tool.
 *
 * The "version" file of the "cmd" package contains the simple to print out
 * relevant version information of the CLI application.
 */
package cmd

import (
	"github.com/weburz/terox/internal/version"

	"github.com/spf13/cobra"
)

var shortUsageHelp = "Print the version information"
var LongUsageHelp = `Prints the detailed version information of the application,
including version, commit hash, build date, and Go version.`

// Handle the logic for the "version" command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: shortUsageHelp,
	Long:  LongUsageHelp,
	Run: func(cmd *cobra.Command, args []string) {
		// Fetch the version information of the tool
		version := version.Get()

		rootCmd.Println("### Terox Build Information ###")
		rootCmd.Printf("Version: \t%s\n", version.Version)
		rootCmd.Printf("Git Version: \t%s\n", version.GitVersion)
		rootCmd.Printf("Git Commit: \t%s\n", version.GitCommit)
		rootCmd.Printf("Build Date: \t%s\n", version.BuildDate)
		rootCmd.Printf("Go Version: \t%s\n", version.GoVersion)
		rootCmd.Printf("Compiler: \t%s\n", version.Compiler)
		rootCmd.Printf("Platform: \t%s\n", version.Platform)
	},
}

// Register the "version" command for the CLI tool
func init() {
	rootCmd.AddCommand(versionCmd)
}
