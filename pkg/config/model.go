package config

import (
	"fmt"
	"strings"
)

// HostConfig represents a parsed Host block from ~/.ssh/config.
type HostConfig struct {
	// Alias is the identifier specified in 'Host <alias>'
	Alias string `json:"alias"`

	// HostName is the actual server IP or domain
	HostName string `json:"hostname,omitempty"`

	// User is the SSH login username
	User string `json:"user,omitempty"`

	// Port is the SSH port (defaults to 22 if not specified)
	Port int `json:"port,omitempty"`

	// IdentityFiles lists private key paths associated with this host
	IdentityFiles []string `json:"identity_files,omitempty"`

	// RemoteForwards lists existing RemoteForward directives (e.g., "22222 localhost:22")
	RemoteForwards []string `json:"remote_forwards,omitempty"`

	// RawDirectives stores all directives in order for rewriting
	RawDirectives map[string]string `json:"-"`
}

// DisplayLabel formats a friendly label for interactive CLI selection.
func (h *HostConfig) DisplayLabel() string {
	parts := []string{h.Alias}
	details := []string{}

	if h.User != "" && h.HostName != "" {
		details = append(details, fmt.Sprintf("%s@%s", h.User, h.HostName))
	} else if h.HostName != "" {
		details = append(details, h.HostName)
	}

	if h.Port != 0 && h.Port != 22 {
		details = append(details, fmt.Sprintf("port %d", h.Port))
	}

	if len(details) > 0 {
		return fmt.Sprintf("%-16s (%s)", parts[0], strings.Join(details, ", "))
	}
	return h.Alias
}

// GetExistingTunnelPort extracts an existing RemoteForward port if one forwards to localhost:22.
func (h *HostConfig) GetExistingTunnelPort() int {
	for _, rf := range h.RemoteForwards {
		fields := strings.Fields(rf)
		if len(fields) >= 2 {
			// Check if destination is localhost:22 or 127.0.0.1:22
			dest := fields[1]
			if strings.HasSuffix(dest, ":22") || dest == "22" {
				var port int
				if _, err := fmt.Sscanf(fields[0], "%d", &port); err == nil && port > 0 {
					return port
				}
			}
		}
	}
	return 0
}

