package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// DefaultSSHConfigPath returns the absolute path to ~/.ssh/config.
func DefaultSSHConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine user home directory: %w", err)
	}
	return filepath.Join(home, ".ssh", "config"), nil
}

// ParseSSHConfigFile reads an OpenSSH client config file and returns parsed Host entries.
// It ignores wildcard host patterns (such as Host *) for remote selection.
func ParseSSHConfigFile(path string) ([]*HostConfig, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return []*HostConfig{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open ssh config at %s: %w", path, err)
	}
	defer file.Close()

	var hosts []*HostConfig
	var current *HostConfig

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.ToLower(fields[0])
		val := strings.Join(fields[1:], " ")

		if key == "host" {
			// A line can declare multiple hosts, e.g. "Host srv1 srv2"
			for _, alias := range fields[1:] {
				// We create a host block for the primary alias
				current = &HostConfig{
					Alias:         alias,
					Port:          22,
					IdentityFiles: []string{},
					RemoteForwards: []string{},
					RawDirectives: make(map[string]string),
				}
				hosts = append(hosts, current)
				break // For simplicity in mapping, track the first pattern
			}
			continue
		}

		if current == nil {
			// Directives before any Host block apply globally
			continue
		}

		current.RawDirectives[key] = val

		switch key {
		case "hostname":
			current.HostName = val
		case "user":
			current.User = val
		case "port":
			if p, err := strconv.Atoi(val); err == nil {
				current.Port = p
			}
		case "identityfile":
			current.IdentityFiles = append(current.IdentityFiles, val)
		case "remoteforward":
			current.RemoteForwards = append(current.RemoteForwards, val)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading ssh config: %w", err)
	}

	// Filter out wildcard and negative match patterns (e.g. *, !host, etc.)
	var filtered []*HostConfig
	for _, h := range hosts {
		if strings.ContainsAny(h.Alias, "*?!") {
			continue
		}
		filtered = append(filtered, h)
	}

	return filtered, nil
}

// InjectRemoteForward ensures that the Host block for hostAlias contains a RemoteForward directive.
// Example: RemoteForward <remotePort> localhost:<localPort>
func InjectRemoteForward(configPath string, hostAlias string, remotePort int, localPort int) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create ssh directory %s: %w", dir, err)
	}

	directive := fmt.Sprintf("RemoteForward %d localhost:%d", remotePort, localPort)

	// Check if file exists
	contentBytes, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		// Create new config file with host block
		newContent := fmt.Sprintf("Host %s\n    %s\n", hostAlias, directive)
		return os.WriteFile(configPath, []byte(newContent), 0600)
	} else if err != nil {
		return fmt.Errorf("failed to read ssh config: %w", err)
	}

	lines := strings.Split(string(contentBytes), "\n")
	var newLines []string

	inTargetHost := false
	alreadyHasForward := false
	hostFound := false
	inserted := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if entering a new Host or Match block
		fields := strings.Fields(trimmed)
		if len(fields) >= 2 && strings.EqualFold(fields[0], "host") {
			// If we were in target host and didn't insert yet, insert before next Host block
			if inTargetHost && !alreadyHasForward && !inserted {
				newLines = append(newLines, fmt.Sprintf("    %s", directive))
				inserted = true
			}

			// Check if this is our target host
			isTarget := false
			for _, pattern := range fields[1:] {
				if pattern == hostAlias {
					isTarget = true
					break
				}
			}

			inTargetHost = isTarget
			if isTarget {
				hostFound = true
			}
		} else if inTargetHost && len(fields) >= 1 && strings.EqualFold(fields[0], "match") {
			if !alreadyHasForward && !inserted {
				newLines = append(newLines, fmt.Sprintf("    %s", directive))
				inserted = true
			}
			inTargetHost = false
		}

		// Check if directive already exists inside the target block
		if inTargetHost && len(fields) >= 2 && strings.EqualFold(fields[0], "remoteforward") {
			// Check if same remote port is forwarded
			if fields[1] == strconv.Itoa(remotePort) {
				alreadyHasForward = true
			}
		}

		newLines = append(newLines, line)

		// If this is the last line of the file and we are in target host
		if i == len(lines)-1 && inTargetHost && !alreadyHasForward && !inserted {
			newLines = append(newLines, fmt.Sprintf("    %s", directive))
			inserted = true
		}
	}

	// If host was never found in the file, append a new Host block
	if !hostFound {
		if len(newLines) > 0 && strings.TrimSpace(newLines[len(newLines)-1]) != "" {
			newLines = append(newLines, "")
		}
		newLines = append(newLines, fmt.Sprintf("Host %s", hostAlias))
		newLines = append(newLines, fmt.Sprintf("    %s", directive))
	}

	// Make a backup copy before writing
	bakPath := configPath + ".rlink.bak"
	_ = os.WriteFile(bakPath, contentBytes, 0600)

	result := strings.Join(newLines, "\n")
	return os.WriteFile(configPath, []byte(result), 0600)
}
