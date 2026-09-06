package cmd

import (
	"fmt"
	"strings"

	"github.com/quaywin/rlink/pkg/config"
	"github.com/quaywin/rlink/pkg/remote"
	"github.com/spf13/cobra"
)

var flagCheckStatus bool

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"list", "ls"},
	Short:   "Display configured remote hosts, tunnels, and wrapper status",
	Long:    "Inspect ~/.ssh/config and show all hosts configured with rlink RemoteForward tunnels.",
	RunE:    runStatus,
}

func init() {
	statusCmd.Flags().BoolVarP(&flagCheckStatus, "check", "c", false, "Test live SSH connection and discover installed remote wrappers")
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	sshConfigPath, err := config.DefaultSSHConfigPath()
	if err != nil {
		return err
	}

	hosts, err := config.ParseSSHConfigFile(sshConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", sshConfigPath, err)
	}

	type ConfiguredHost struct {
		Host   *config.HostConfig
		Port   int
		Direct string
	}

	var configured []ConfiguredHost
	for _, h := range hosts {
		if port := h.GetExistingTunnelPort(); port > 0 {
			for _, rf := range h.RemoteForwards {
				if strings.Contains(rf, fmt.Sprintf("%d", port)) {
					configured = append(configured, ConfiguredHost{
						Host:   h,
						Port:   port,
						Direct: rf,
					})
					break
				}
			}
		}
	}

	fmt.Println()
	fmt.Println("🔗 rlink Configured Hosts & Tunnels")
	fmt.Println("=========================================================")

	if len(configured) == 0 {
		fmt.Println("No hosts currently configured with RemoteForward tunnels in ~/.ssh/config.")
		fmt.Println()
		fmt.Println("Run 'rlink setup' to configure a remote server.")
		fmt.Println()
		return nil
	}

	for _, item := range configured {
		h := item.Host
		target := h.HostName
		if target == "" {
			target = h.Alias
		}
		if h.User != "" {
			target = h.User + "@" + target
		}

		fmt.Printf("\n• Host: \033[1m%s\033[0m\n", h.Alias)
		fmt.Printf("  Target:         %s (port %d)\n", target, h.Port)
		fmt.Printf("  Reverse Tunnel: RemoteForward %s\n", item.Direct)

		if flagCheckStatus {
			fmt.Printf("  Testing live connection to '%s'...\n", h.Alias)
			client := remote.NewSSHClient(h.Alias)
			out, err := client.Run("ls ~/.local/bin 2>/dev/null || true")
			if err != nil {
				fmt.Printf("  Status:         \033[31m✗ Unreachable (%v)\033[0m\n", err)
			} else {
				fmt.Printf("  Status:         \033[32m✓ Connected\033[0m\n")
				// Check for known wrapper names
				files := strings.Fields(out)
				var foundWrappers []string
				known := []string{"rzed", "rcode", "rcursor", "rwindsurf", "rcode-insiders", "rsubl", "ropen", "rclip", "rpaste", "zr", "cr", "cur", "zed", "code", "cursor", "windsurf"}
				for _, f := range files {
					for _, k := range known {
						if f == k {
							foundWrappers = append(foundWrappers, f)
							break
						}
					}
				}
				if len(foundWrappers) > 0 {
					fmt.Printf("  Wrappers:       %s\n", strings.Join(foundWrappers, ", "))
				} else {
					fmt.Printf("  Wrappers:       None found in ~/.local/bin\n")
				}
			}
		}
	}

	fmt.Println()
	fmt.Println("Tip: Pass '-c' / '--check' to test live SSH connections and detect wrappers.")
	fmt.Println()

	return nil
}
