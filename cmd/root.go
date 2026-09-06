package cmd

import (
	"fmt"
	"os"

	"github.com/quaywin/rlink/pkg/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rlink",
	Short: "rlink bridges remote terminal sessions with local GUI editors (Zed, VS Code, Cursor)",
	Long: `rlink is a smart CLI wizard running on your local machine.
It configures and provisions lightweight wrapper scripts on your remote servers,
enabling you to type 'zr .' or 'cr .' in remote terminal sessions to instantly
open remote directories in your local Zed, VS Code, or Cursor editor.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	rootCmd.Version = version.GetVersionInfo()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(versionCmd)
}
