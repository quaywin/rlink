package detector

import (
	"fmt"
	"net"
	"runtime"
	"time"
)

// SSHDaemonStatus represents the status of the local machine's SSH server.
type SSHDaemonStatus struct {
	IsListening bool
	Port        int
	Error       error
	HelpGuide   string
}

// CheckLocalSSHServer tests whether an SSH server is actively listening on local port.
func CheckLocalSSHServer(port int) SSHDaemonStatus {
	if port <= 0 {
		port = 22
	}

	address := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
	if err != nil {
		status := SSHDaemonStatus{
			IsListening: false,
			Port:        port,
			Error:       err,
		}

		if runtime.GOOS == "darwin" {
			status.HelpGuide = "To enable SSH on macOS: Open System Settings -> General -> Sharing -> Enable 'Remote Login'."
		} else {
			status.HelpGuide = "To enable SSH on Linux: Run 'sudo systemctl enable --now ssh' or 'sudo service ssh start'."
		}
		return status
	}
	_ = conn.Close()

	return SSHDaemonStatus{
		IsListening: true,
		Port:        port,
	}
}
