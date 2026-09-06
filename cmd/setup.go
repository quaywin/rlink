package cmd

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/quaywin/rlink/pkg/auth"
	"github.com/quaywin/rlink/pkg/config"
	"github.com/quaywin/rlink/pkg/detector"
	"github.com/quaywin/rlink/pkg/remote"
	"github.com/quaywin/rlink/pkg/template"
	"github.com/spf13/cobra"
)

var (
	flagTool   string
	flagEditor string
	flagName   string
	flagHost   string
	flagPort   int
	flagYes    bool
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive wizard to configure local tools and provision remote wrappers",
	Long: `rlink setup configures local tools (GUI editors, system opener, clipboard, custom commands)
and provisions wrapper scripts on your remote server via Reverse SSH Tunnel.

You can run this multiple times to set up multiple different tools on the same server with custom names
(for example: 'rzed' for Zed, 'rcode' for VS Code, 'ropen' for Web/File Opener, 'rclip' for Clipboard).`,
	RunE: runSetupWizard,
}

func init() {
	setupCmd.Flags().StringVarP(&flagTool, "tool", "t", "", "Target tool or command: zed, code, cursor, windsurf, open, clip, paste, or custom command")
	setupCmd.Flags().StringVarP(&flagEditor, "editor", "e", "", "Target GUI editor (alias for --tool)")
	setupCmd.Flags().StringVarP(&flagName, "name", "n", "", "Custom remote wrapper command name (default: rzed, rcode, ropen, rclip, etc.)")
	setupCmd.Flags().StringVarP(&flagHost, "host", "H", "", "Remote SSH host alias from ~/.ssh/config or user@hostname")
	setupCmd.Flags().IntVarP(&flagPort, "port", "p", 0, "Remote Forward port (default: 22222)")
	setupCmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "Automatically deploy without interactive confirmation prompt")
}

func runSetupWizard(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println("🔗 rlink — Bridge Remote Terminal to Local Actions")
	fmt.Println("=========================================================")
	fmt.Println()

	// Handle alias --editor -> --tool
	if flagTool == "" && flagEditor != "" {
		flagTool = flagEditor
	}

	// -------------------------------------------------------------
	// 1. Local Tool Detection & Selection
	// -------------------------------------------------------------
	allDetected := detector.DetectAllTools()
	if len(allDetected) == 0 {
		return fmt.Errorf("no supported tools found on local system")
	}

	var chosenTool *detector.DetectedTool
	var wrapperCmdName string = flagName

	if flagTool != "" {
		// 1. Try matching against predefined tools
		for i := range allDetected {
			if strings.EqualFold(string(allDetected[i].Type), flagTool) ||
				strings.EqualFold(allDetected[i].DefaultWrapperName, flagTool) ||
				strings.EqualFold(allDetected[i].Name, flagTool) {
				chosenTool = &allDetected[i]
				break
			}
		}

		// 2. If not matched, attempt custom command lookup
		if chosenTool == nil {
			custom, err := detector.ValidateCustomCommand(flagTool)
			if err != nil {
				return fmt.Errorf("tool %q not recognized and not found as local executable: %w", flagTool, err)
			}
			chosenTool = custom
		}
	} else {
		var toolChoices []huh.Option[string]
		defaultToolType := string(allDetected[0].Type)

		for _, t := range allDetected {
			label := t.DisplayLabel()
			toolChoices = append(toolChoices, huh.NewOption(label, string(t.Type)))
			if t.IsInstalled && defaultToolType == string(allDetected[0].Type) && t.Type != allDetected[0].Type {
				defaultToolType = string(t.Type)
			}
		}
		toolChoices = append(toolChoices, huh.NewOption("⚙️  [+] Custom Local Command...", "custom_manual_command"))

		var selectedToolType string = defaultToolType

		formTool := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Step 1: Select Tool or Action").
					Description("Choose an editor, system tool, or custom command to bridge to remote terminal").
					Options(toolChoices...).
					Value(&selectedToolType),
			),
		)

		if err := formTool.Run(); err != nil {
			return err
		}

		if selectedToolType == "custom_manual_command" {
			customCmdInput := ""
			formCustom := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Enter Local Command").
						Description("Executable name or path available on this local machine").
						Placeholder("e.g. mpv, subl, gimp, git-gui").
						Value(&customCmdInput),
				),
			)
			if err := formCustom.Run(); err != nil {
				return err
			}
			customTool, err := detector.ValidateCustomCommand(customCmdInput)
			if err != nil {
				return fmt.Errorf("invalid custom command: %w", err)
			}
			chosenTool = customTool
		} else {
			for i := range allDetected {
				if string(allDetected[i].Type) == selectedToolType {
					chosenTool = &allDetected[i]
					break
				}
			}
		}
	}

	if chosenTool == nil {
		return fmt.Errorf("no tool selected")
	}

	// Wrapper command name prompt (if not specified via --name flag)
	if wrapperCmdName == "" {
		wrapperCmdName = chosenTool.DefaultWrapperName

		descriptionText := fmt.Sprintf(
			"Command you will type on the remote terminal to trigger %s.\n"+
				"Recommended convention: '%s' (with 'r' prefix for remote).\n"+
				"Tip: You can customize any name, and wrap multiple tools on the same server!",
			chosenTool.Name,
			chosenTool.DefaultWrapperName,
		)

		formWrapperName := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Remote Command Name (Wrapper Name)").
					Description(descriptionText).
					Placeholder(fmt.Sprintf("e.g. %s", chosenTool.DefaultWrapperName)).
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
	// 3. Reverse SSH Tunnel Setup
	// -------------------------------------------------------------
	sshStatus := detector.CheckLocalSSHServer(22)
	if !sshStatus.IsListening {
		fmt.Println()
		fmt.Printf("⚠️  Notice: Local SSH daemon is not currently listening on port 22.\n")
		fmt.Printf("   %s\n\n", sshStatus.HelpGuide)
	}

	connectPort := 22222
	injectConfig := false

	existingTunnelPort := 0
	if selectedHostConfig != nil {
		existingTunnelPort = selectedHostConfig.GetExistingTunnelPort()
	}

	if flagPort > 0 {
		connectPort = flagPort
		injectConfig = true
	} else if existingTunnelPort > 0 {
		// Reuse existing tunnel automatically without prompt
		connectPort = existingTunnelPort
		injectConfig = false
		fmt.Printf("✓ Found existing RemoteForward tunnel on port %d for '%s'. Reusing tunnel.\n", existingTunnelPort, selectedHost)
	} else {
		// New host configuration
		injectConfig = true
		if !flagYes {
			portStr := "22222"
			formTunnel := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Step 3: Remote Forward Port").
						Description("Port on the remote server forwarded back to your local machine (default: 22222)").
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
			if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
				connectPort = p
			}
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
	// Ensure dedicated Ed25519 keypair for seamless passwordless remote triggering
	keypair, keyErr := auth.EnsureLocalSSHKeyPair()
	sshKeyPathOnRemote := ""
	if keyErr == nil {
		if err := auth.AuthorizeLocalKey(keypair.PublicKeyContent); err == nil {
			sshKeyPathOnRemote = "$HOME/.ssh/rlink_id_ed25519"
		}
	}

	wrapperConfig := template.WrapperConfig{
		ToolType:      string(chosenTool.Type),
		ToolCategory:  string(chosenTool.Category),
		ToolName:      chosenTool.Name,
		CommandName:   wrapperCmdName,
		HostAlias:     selectedHost,
		LocalUser:     localUsername,
		ConnectPort:   connectPort,
		SSHKeyPath:    sshKeyPathOnRemote,
		SyntaxPattern: chosenTool.SyntaxTemplate,
		LocalBinary:   chosenTool.BinaryPath,
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
					Description(fmt.Sprintf("Will install %s wrapper at remote bin directory", chosenTool.Name)).
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

	// Deploy dedicated private key to remote server for passwordless reverse authentication
	if keypair != nil {
		remoteKeyPath := "~/.ssh/rlink_id_ed25519"
		if err := sshClient.DeployPrivateKey(remoteKeyPath, keypair.PrivateKeyContent); err == nil {
			fmt.Println("✓ Configured dedicated Ed25519 key on remote server for passwordless auth.")
		}
	}

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

	// If user agreed, update local ~/.ssh/config with RemoteForward
	if injectConfig {
		fmt.Printf("🔧 Injecting 'RemoteForward %d localhost:22' into %s (Host %s)...\n", connectPort, sshConfigPath, selectedHost)
		if err := config.InjectRemoteForward(sshConfigPath, selectedHost, connectPort, 22); err != nil {
			fmt.Printf("Warning: Failed to update %s automatically: %v\n", sshConfigPath, err)
			fmt.Printf("Please manually add to %s under Host %s:\n    RemoteForward %d localhost:22\n", sshConfigPath, selectedHost, connectPort)
		} else {
			fmt.Println("✓ Updated ~/.ssh/config successfully.")
		}
	}

	// Print final congratulations & usage examples tailored to chosen tool
	fmt.Println("\n=========================================================")
	fmt.Println("🎉 Setup Complete!")
	fmt.Println("=========================================================")
	fmt.Printf("On remote server '%s', you can now run:\n", selectedHost)
	switch chosenTool.Type {
	case detector.ToolClipCopy:
		fmt.Printf("   $ cat file.txt | %s\n", wrapperCmdName)
		fmt.Printf("   $ %s \"text to copy\"\n", wrapperCmdName)
	case detector.ToolClipPaste:
		fmt.Printf("   $ %s > file.txt\n", wrapperCmdName)
		fmt.Printf("   $ %s | grep something\n", wrapperCmdName)
	case detector.ToolOpener:
		fmt.Printf("   $ %s https://github.com\n", wrapperCmdName)
		fmt.Printf("   $ %s plot.png\n", wrapperCmdName)
	default:
		fmt.Printf("   $ %s .\n", wrapperCmdName)
	}
	fmt.Printf("to trigger %s on this machine!\n\n", chosenTool.Name)
	fmt.Println("💡 Tip: Want to set up another tool (e.g. 'ropen', 'rclip', or another editor)?")
	fmt.Println("   Simply run 'rlink setup' again!")
	fmt.Println()

	return nil
}
