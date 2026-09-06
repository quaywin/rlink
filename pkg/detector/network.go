package detector

import (
	"bytes"
	"net"
	"os/exec"
	"strings"
)

// NetworkSuggestion holds potential IPs for direct connection from Remote to Local.
type NetworkSuggestion struct {
	Type        string // "Tailscale", "LAN", "Hostname"
	Address     string
	Description string
}

// DetectNetworkSuggestions discovers Tailscale IPs and local LAN IPs.
func DetectNetworkSuggestions() []NetworkSuggestion {
	var suggestions []NetworkSuggestion

	// 1. Detect Tailscale IPv4
	if tsIP := getTailscaleIPv4(); tsIP != "" {
		suggestions = append(suggestions, NetworkSuggestion{
			Type:        "Tailscale",
			Address:     tsIP,
			Description: "Tailscale private mesh IP (Recommended if both machines are on Tailscale)",
		})
	}

	// 2. Detect Local LAN IPv4 addresses
	lanIPs := getLocalIPv4s()
	for _, ip := range lanIPs {
		suggestions = append(suggestions, NetworkSuggestion{
			Type:        "LAN",
			Address:     ip,
			Description: "Local Area Network IP (usable if on same Wi-Fi/VPC)",
		})
	}

	return suggestions
}

func getTailscaleIPv4() string {
	cmd := exec.Command("tailscale", "ip", "-4")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		ip := strings.TrimSpace(out.String())
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	return ""
}

func getLocalIPv4s() []string {
	var ips []string
	interfaces, err := net.Interfaces()
	if err != nil {
		return ips
	}

	for _, iface := range interfaces {
		// Skip down and loopback interfaces
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip4 := ip.To4()
			if ip4 != nil && !ip4.IsLinkLocalUnicast() {
				// Don't duplicate
				already := false
				for _, existing := range ips {
					if existing == ip4.String() {
						already = true
						break
					}
				}
				if !already {
					ips = append(ips, ip4.String())
				}
			}
		}
	}
	return ips
}
