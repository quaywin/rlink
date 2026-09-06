package cmd

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/quaywin/rlink/pkg/config"
	"github.com/quaywin/rlink/pkg/detector"
	"github.com/quaywin/rlink/pkg/remote"
	"github.com/quaywin/rlink/pkg/template"
	"github.com/spf13/cobra"
)

var (
	flagEditor string
	flagName   string
	flagHost   string
	flagPort   int
	flagMode   string
	flagYes    bool
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive wizard to configure local editor and provision remote wrapper",
	Long: `rlink setup configures your local editor mapping and provisions a wrapper script on your remote server.
You can run this multiple times to set up multiple different editors on the same server with custom names
(for example: 'zr' or 'zed' for Zed, 'cr' or 'code' for VS Code, 'cur' or 'cursor' for Cursor).`,
	RunE: runSetupWizard,
}

func init() {
	setupCmd.Flags().StringVarP(&flagEditor, "editor", "e", "", "Target GUI editor: zed, code, or cursor")
	setupCmd.Flags().StringVarP(&flagName, "name", "n", "", "Custom remote wrapper command name (e.g. zr, zed, cr, code, cur, cursor)")
	setupCmd.Flags().StringVarP(&flagHost, "host", "H", "", "Remote SSH host alias from ~/.ssh/config or user@hostname")
	setupCmd.Flags().IntVarP(&flagPort, "port", "p", 0, "Forward/connect port (default: 22222 for tunnel, 22 for direct)")
	setupCmd.Flags().StringVarP(&flagMode, "mode", "m", "", "Connection mode: 'tunnel' (reverse SSH) or 'direct' (Tailscale/LAN)")
	setupCmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "Automatically deploy without interactive confirmation prompt")
}

func runSetupWizard(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println("🔗 rlink — Bridge Remote Terminal to Local GUI Editor")
	fmt.Println("=========================================================")
	fmt.Println()

	// -------------------------------------------------------------
	// 1. Local Editor Detection & Selection
	// -------------------------------------------------------------
	detected := detector.DetectInstalledEditors()
	if len(detected) == 0 {
		return fmt.Errorf("no supported editor definitions found")
	}

	var chosenEditor *detector.DetectedEditor
	var wrapperCmdName string = flagName

	if flagEditor != "" {
		for i := range detected {
			if strings.EqualFold(string(detected[i].Type), flagEditor) {
				chosenEditor = &detected[i]
				break
			}
		}
		if chosenEditor == nil {
			return fmt.Errorf("unknown editor %q. Supported editors: zed, code, cursor", flagEditor)
		}
	} else {
		var editorChoices []huh.Option[string]
		defaultEditorType := string(detected[0].Type)

		for _, ed := range detected {
			label := ed.DisplayLabel()
			editorChoices = append(editorChoices, huh.NewOption(label, string(ed.Type)))
			if ed.IsInstalled && defaultEditorType == string(detected[0].Type) {
				defaultEditorType = string(ed.Type)
			}
		}

		var selectedEditorType string = defaultEditorType

		formEditor := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Step 1: Select Local GUI Editor").
					Description("Choose which editor to open when triggered from remote terminal").
					Options(editorChoices...).
					Value(&selectedEditorType),
			),
		)

		if err := formEditor.Run(); err != nil {
			return err
		}

		for i := range detected {
			if string(detected[i].Type) == selectedEditorType {
				chosenEditor = &detected[i]
				break
			}
		}
	}

	// Wrapper command name prompt (if not specified via --name flag)
	if wrapperCmdName == "" {
		wrapperCmdName = chosenEditor.DefaultWrapperName

		descriptionText := fmt.Sprintf(
			"Command you will type on the remote terminal to open %s.\n"+
				"Tip: You can use any custom name (e.g. '%s', '%s', '%s').\n"+
				"Run setup again anytime to wrap other editors on the same server!",
			chosenEditor.Name,
			chosenEditor.DefaultWrapperName,
			strings.ToLower(chosenEditor.Name),
			string(strings.ToLower(chosenEditor.Name)[0]),
		)

		formWrapperName := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Remote Command Name (Wrapper Name)").
					Description(descriptionText).
					Placeholder("e.g. zr, zed, cr, code, cursor").
					Value(&wrapperCmdName),
			),
		)

		if err := formWrapperName.Run(); err != nil {
			return err
		}
	}

	wrapperCmdName = strings.TrimSpace(wrapperCmdName)
	if wrapperCmdName == "" {
		return fmt.Errorf("remote command name cannot be empty")
	}

	// -------------------------------------------------------------
	// 2. SSH Config Parsing & Remote Host Selection
	// -------------------------------------------------------------
	sshConfigPath, _ := config.DefaultSSHConfigPath()
	hosts, err := config.ParseSSHConfigFile(sshConfigPath)
	if err != nil {
		fmt.Printf("Warning: Failed to read %s: %v\n", sshConfigPath, err)
	}

	var selectedHost string = flagHost
	var selectedHostConfig *config.HostConfig

	if selectedHost == "" {
		var hostChoices []huh.Option[string]
		for _, h := range hosts {
			hostChoices = append(hostChoices, huh.NewOption(h.DisplayLabel(), h.Alias))
		}
		hostChoices = append(hostChoices, huh.NewOption("[+] Enter remote host manually...", "custom_manual_host"))

		if len(hosts) > 0 {
			selectedHost = hosts[0].Alias
		} else {
			selectedHost = "custom_manual_host"
		}

		formHost := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Step 2: Select Remote Server").
					Description("Target server where the wrapper script will be installed").
					Options(hostChoices...).
					Value(&selectedHost),
			),
		)

		if err := formHost.Run(); err != nil {
			return err
		}

		if selectedHost == "custom_manual_host" {
			manualHostInput := ""
			formManualHost := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Enter Remote Host").
						Description("SSH host alias or user@hostname:port").
						Placeholder("e.g. dev-box or ubuntu@192.168.1.50").
						Value(&manualHostInput),
				),
			)
			if err := formManualHost.Run(); err != nil {
				return err
			}
			selectedHost = strings.TrimSpace(manualHostInput)
		}
	}

	if selectedHost == "" {
		return fmt.Errorf("remote host cannot be empty")
	}

	// Find if host exists in ssh config to check for existing tunnels
	for _, h := range hosts {
		if h.Alias == selectedHost {
			selectedHostConfig = h
			break
		}
	}

	// -------------------------------------------------------------
	// 3. Connection Strategy (Reverse Tunnel vs Direct / Tailscale)
	// -------------------------------------------------------------
	strategyChoice := flagMode
	existingTunnelPort := 0
	if selectedHostConfig != nil {
		existingTunnelPort = selectedHostConfig.GetExistingTunnelPort()
	}

	if strategyChoice == "" {
		strategyChoice = "tunnel"
		var strategyOptions []huh.Option[string]

		if existingTunnelPort > 0 {
			strategyOptions = append(strategyOptions, huh.NewOption(
				fmt.Sprintf("Reuse Existing SSH Tunnel (Port %d already forwarded for %s)", existingTunnelPort, selectedHost),
				"tunnel",
			))
		} else {
			strategyOptions = append(strategyOptions, huh.NewOption(
				"Reverse SSH Tunnel (Recommended: Works behind NAT, Wi-Fi, Firewalls)",
				"tunnel",
			))
		}

		strategyOptions = append(strategyOptions, huh.NewOption(
			"Direct / Tailscale Network (Uses Tailscale private mesh or LAN IP)",
			"direct",
		))

		formStrategy := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Step 3: Connection Strategy").
					Description("How the remote server communicates back to your local machine").
					Options(strategyOptions...).
					Value(&strategyChoice),
			),
		)

		if err := formStrategy.Run(); err != nil {
			return err
		}
	}

	connectHost := "127.0.0.1"
	connectPort := 22222
	if flagPort > 0 {
		connectPort = flagPort
	} else if existingTunnelPort > 0 {
		connectPort = existingTunnelPort
	}

	injectConfig := false

	if strategyChoice == "tunnel" {
		if existingTunnelPort > 0 && flagPort == 0 {
			// Reuse existing tunnel automatically
			fmt.Printf("✓ Detected existing RemoteForward on port %d for '%s'. Reusing tunnel.\n", existingTunnelPort, selectedHost)
			connectPort = existingTunnelPort
			injectConfig = false
		} else {
			portStr := strconv.Itoa(connectPort)
			injectConfig = true

			formTunnel := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Remote Forward Port").
						Description("Port on remote server that forwards to local SSH server (port 22)").
						Value(&portStr),
					huh.NewConfirm().
						Title("Update ~/.ssh/config automatically?").
						Description(fmt.Sprintf("Inject 'RemoteForward %s localhost:22' into Host %s", portStr, selectedHost)).
						Value(&injectConfig),
				),
			)
			if err := formTunnel.Run(); err != nil {
				return err
			}
			if p, err := strconv.Atoi(portStr); err == nil {
				connectPort = p
			}
		}
	} else {
		// Direct mode: Suggest IPs
		netSuggestions := detector.DetectNetworkSuggestions()
		var ipChoices []huh.Option[string]
		for _, s := range netSuggestions {
			ipChoices = append(ipChoices, huh.NewOption(fmt.Sprintf("%s: %s (%s)", s.Type, s.Address, s.Description), s.Address))
		}
		ipChoices = append(ipChoices, huh.NewOption("[+] Enter IP/hostname manually", "manual_ip"))

		var selectedIP string
		if len(ipChoices) > 0 {
			selectedIP = ipChoices[0].Value
		} else {
			selectedIP = "manual_ip"
		}

		formDirectIP := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Local Machine IP Address").
					Description("Choose the IP address of this local machine reachable by the remote server").
					Options(ipChoices...).
					Value(&selectedIP),
			),
		)
		if err := formDirectIP.Run(); err != nil {
			return err
		}

		if selectedIP == "manual_ip" {
			manualIPInput := ""
			formManualIP := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Enter Local IP/Hostname").
						Placeholder("e.g. 100.x.y.z or my-macbook.local").
						Value(&manualIPInput),
				),
			)
			if err := formManualIP.Run(); err != nil {
				return err
			}
			connectHost = strings.TrimSpace(manualIPInput)
		} else {
			connectHost = selectedIP
		}

		portStr := "22"
		if flagPort > 0 {
			portStr = strconv.Itoa(flagPort)
		}

		formDirectPort := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Local SSH Port").
					Description("Port where the SSH daemon is listening on this machine").
					Value(&portStr),
			),
		)
		if err := formDirectPort.Run(); err != nil {
			return err
		}
		if p, err := strconv.Atoi(portStr); err == nil {
			connectPort = p
		}
	}

	// Local username determination
	localUsername := os.Getenv("USER")
	if localUsername == "" {
		if u, err := user.Current(); err == nil {
			localUsername = u.Username
		}
	}

	if !flagYes {
		formLocalUser := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Local Machine Username").
					Description("Your username on this machine (used by remote wrapper to SSH back)").
					Value(&localUsername),
			),
		)
		if err := formLocalUser.Run(); err != nil {
			return err
		}
	}

	// -------------------------------------------------------------
	// 4. Script Generation & Remote Provisioning
	// -------------------------------------------------------------
	connMode := template.ModeReverseTunnel
	if strategyChoice == "direct" {
		connMode = template.ModeDirectIP
	}

	wrapperConfig := template.WrapperConfig{
		EditorName:     chosenEditor.Name,
		CommandName:    wrapperCmdName,
		HostAlias:      selectedHost,
		LocalUser:      localUsername,
		ConnectHost:    connectHost,
		ConnectPort:    connectPort,
		ConnectionMode: connMode,
		SyntaxPattern:  chosenEditor.SyntaxTemplate,
	}

	scriptContent, err := template.GenerateWrapper(wrapperConfig)
	if err != nil {
		return fmt.Errorf("failed to generate wrapper script: %w", err)
	}

	shouldDeploy := true
	if !flagYes {
		formDeploy := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("Deploy '%s' to remote server '%s' now?", wrapperCmdName, selectedHost)).
					Description(fmt.Sprintf("Will install %s wrapper at remote bin directory", chosenEditor.Name)).
					Value(&shouldDeploy),
			),
		)
		if err := formDeploy.Run(); err != nil {
			return err
		}
	}

	if !shouldDeploy {
		fmt.Println("\n📄 Generated wrapper script (dry-run output):")
		fmt.Println("----------------------------------------------")
		fmt.Println(scriptContent)
		return nil
	}

	// Execute remote provisioning
	fmt.Printf("\n🚀 Connecting to '%s' to deploy '%s' wrapper...\n", selectedHost, wrapperCmdName)
	sshClient := remote.NewSSHClient(selectedHost)

	targetDir, err := sshClient.DetectRemoteBinDir()
	if err != nil {
		fmt.Printf("Notice: Could not automatically detect remote bin directory: %v\n", err)
		targetDir = "~/.local/bin"
	}

	targetFilePath := fmt.Sprintf("%s/%s", strings.TrimRight(targetDir, "/"), wrapperCmdName)
	fmt.Printf("📦 Target remote path: %s\n", targetFilePath)

	if err := sshClient.UploadScript(targetFilePath, scriptContent); err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}
	fmt.Printf("✓ Successfully uploaded '%s' and granted executable permissions (chmod +x).\n", wrapperCmdName)

	// Check if directory is in remote PATH
	isInPath := sshClient.IsDirInRemotePath(targetDir)
	if !isInPath {
		fmt.Printf("\n⚠️  Note: '%s' might not be in your remote $PATH yet.\n", targetDir)
		fmt.Println("   Add the following line to your remote ~/.bashrc or ~/.zshrc if needed:")
		fmt.Printf("   export PATH=\"%s:$PATH\"\n", targetDir)
	}

	// If tunnel mode and user agreed, update local ~/.ssh/config
	if strategyChoice == "tunnel" && injectConfig {
		fmt.Printf("🔧 Injecting 'RemoteForward %d localhost:22' into %s (Host %s)...\n", connectPort, sshConfigPath, selectedHost)
		if err := config.InjectRemoteForward(sshConfigPath, selectedHost, connectPort, 22); err != nil {
			fmt.Printf("Warning: Failed to update %s automatically: %v\n", sshConfigPath, err)
			fmt.Printf("Please manually add to %s under Host %s:\n    RemoteForward %d localhost:22\n", sshConfigPath, selectedHost, connectPort)
		} else {
			fmt.Println("✓ Updated ~/.ssh/config successfully.")
		}
	}

	// Print final congratulations
	fmt.Println("\n=========================================================")
	fmt.Println("🎉 Setup Complete!")
	fmt.Println("=========================================================")
	fmt.Printf("On remote server '%s', you can now simply run:\n", selectedHost)
	fmt.Printf("   $ %s .\n", wrapperCmdName)
	fmt.Printf("to instantly open the current folder in %s on this machine!\n\n", chosenEditor.Name)
	fmt.Println("💡 Tip: Want to wrap another editor on the same server (e.g. Zed alongside VS Code)?")
	fmt.Println("   Simply run 'rlink setup' again and pick a different command name!")
	fmt.Println()

	return nil
}
