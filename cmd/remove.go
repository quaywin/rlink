package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/quaywin/rlink/pkg/config"
	"github.com/quaywin/rlink/pkg/remote"
	"github.com/spf13/cobra"
)

var (
	flagRemoveWrapper string
	flagKeepTunnel    bool
	flagRemoveYes     bool
)

var removeCmd = &cobra.Command{
	Use:     "remove [host]",
	Aliases: []string{"rm", "uninstall"},
	Short:   "Remove rlink configuration and remote wrappers",
	Long: `Remove rlink configurations for a target host.
This can revert RemoteForward directives in ~/.ssh/config and delete wrapper scripts from the remote server.`,
	RunE: runRemove,
}

func init() {
	removeCmd.Flags().StringVarP(&flagRemoveWrapper, "wrapper", "w", "", "Specific wrapper name to remove on remote (e.g. 'zr', 'cr', or 'all')")
	removeCmd.Flags().BoolVar(&flagKeepTunnel, "keep-tunnel", false, "Do not remove RemoteForward directive from ~/.ssh/config")
	removeCmd.Flags().BoolVarP(&flagRemoveYes, "yes", "y", false, "Confirm removal without interactive prompts")
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	sshConfigPath, err := config.DefaultSSHConfigPath()
	if err != nil {
		return err
	}

	hosts, err := config.ParseSSHConfigFile(sshConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", sshConfigPath, err)
	}

	var targetHost string
	if len(args) > 0 {
		targetHost = args[0]
	}

	if targetHost == "" {
		// Discover hosts with tunnels
		var hostChoices []huh.Option[string]
		for _, h := range hosts {
			if h.GetExistingTunnelPort() > 0 {
				hostChoices = append(hostChoices, huh.NewOption(h.DisplayLabel(), h.Alias))
			}
		}

		if len(hostChoices) == 0 {
			fmt.Println("\nNo hosts with rlink RemoteForward tunnels found in ~/.ssh/config.")
			return nil
		}

		formSelect := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select Host to Remove").
					Description("Choose host to remove rlink configuration from").
					Options(hostChoices...).
					Value(&targetHost),
			),
		)
		if err := formSelect.Run(); err != nil {
			return err
		}
	}

	if targetHost == "" {
		return fmt.Errorf("target host cannot be empty")
	}

	confirm := true
	if !flagRemoveYes {
		formConfirm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("Remove rlink setup for host '%s'?", targetHost)).
					Description("This will clean up configuration and optionally remove remote wrappers").
					Value(&confirm),
			),
		)
		if err := formConfirm.Run(); err != nil {
			return err
		}
	}

	if !confirm {
		fmt.Println("Aborted.")
		return nil
	}

	// 1. Remove remote wrapper scripts if requested
	removeRemote := true
	wrapperName := flagRemoveWrapper
	if wrapperName == "" && !flagRemoveYes {
		wrapperName = "zr"
		formWrapper := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Remote wrapper script to remove").
					Description("Name of wrapper in ~/.local/bin/ (or 'all' for zr, cr, cur)").
					Value(&wrapperName),
			),
		)
		if err := formWrapper.Run(); err != nil {
			return err
		}
	}

	if removeRemote && wrapperName != "" {
		fmt.Printf("🗑️  Connecting to '%s' to remove wrapper script(s)...\n", targetHost)
		client := remote.NewSSHClient(targetHost)

		var deleteCmd string
		if wrapperName == "all" {
			deleteCmd = "rm -f ~/.local/bin/zr ~/.local/bin/cr ~/.local/bin/cur ~/.local/bin/zed ~/.local/bin/code ~/.local/bin/cursor"
		} else {
			deleteCmd = fmt.Sprintf("rm -f ~/.local/bin/%s", wrapperName)
		}

		if out, err := client.Run(deleteCmd); err != nil {
			fmt.Printf("Warning: Failed to delete remote wrapper: %s (%v)\n", out, err)
		} else {
			fmt.Println("✓ Remote wrapper script removed.")
		}
	}

	// 2. Remove RemoteForward from ~/.ssh/config
	if !flagKeepTunnel {
		fmt.Printf("🔧 Removing RemoteForward from %s (Host %s)...\n", sshConfigPath, targetHost)
		if err := config.RemoveRemoteForward(sshConfigPath, targetHost); err != nil {
			fmt.Printf("Warning: Failed to update %s: %v\n", sshConfigPath, err)
		} else {
			fmt.Println("✓ Reverted ~/.ssh/config successfully.")
		}
	}

	fmt.Printf("\n✓ rlink cleanup for '%s' completed successfully.\n\n", targetHost)
	return nil
}
