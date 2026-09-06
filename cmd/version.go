package cmd

import (
	"fmt"

	"github.com/quaywin/rlink/pkg/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print rlink version and build architecture",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("rlink %s\n", version.GetVersionInfo())
	},
}
