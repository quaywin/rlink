package remote

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var pidRegex = regexp.MustCompile(`pid=(\d+)`)

// LiveSessionResult describes active SSH sessions to the target host and whether a tunnel was attached.
type LiveSessionResult struct {
	HasMasterSession         bool
	MasterPID                int
	TunnelDynamicallyAttached bool
	RunningPIDs              []int
	Error                    error
}

// CheckMasterSession tests whether an OpenSSH multiplexing master socket is active for hostAlias.
func CheckMasterSession(hostAlias string) (bool, int) {
	cmd := exec.Command("ssh", "-O", "check", hostAlias)
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	err := cmd.Run()
	output := outBuf.String()

	// If exit code is 0 or output indicates master is running
	if err == nil || strings.Contains(output, "Master running") {
		pid := 0
		if matches := pidRegex.FindStringSubmatch(output); len(matches) > 1 {
			if p, convErr := strconv.Atoi(matches[1]); convErr == nil {
				pid = p
			}
		}
		return true, pid
	}

	return false, 0
}

// DynamicallyAttachTunnel asks the running OpenSSH master process to establish the remote forward immediately.
func DynamicallyAttachTunnel(hostAlias string, remotePort int, localPort int) error {
	forwardSpec := fmt.Sprintf("%d:localhost:%d", remotePort, localPort)
	cmd := exec.Command("ssh", "-O", "forward", "-R", forwardSpec, hostAlias)
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		errStr := strings.TrimSpace(errBuf.String())
		// If port forward already exists, treat as success
		if strings.Contains(errStr, "already requested") || strings.Contains(errStr, "already in use") {
			return nil
		}
		return fmt.Errorf("failed to attach tunnel dynamically: %s (%w)", errStr, err)
	}

	return nil
}

// FindRunningSSHProcesses scans the process table for active ssh clients connected to hostAlias.
func FindRunningSSHProcesses(hostAlias string) []int {
	cmd := exec.Command("ps", "-eo", "pid,command")
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf

	if err := cmd.Run(); err != nil {
		return nil
	}

	myPid := os.Getpid()
	lines := strings.Split(outBuf.String(), "\n")
	var pids []int

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		parts := strings.Fields(trimmed)
		if len(parts) < 2 {
			continue
		}

		pid, err := strconv.Atoi(parts[0])
		if err != nil || pid == myPid {
			continue
		}

		command := strings.Join(parts[1:], " ")

		// Look for 'ssh ' or '/ssh ' command containing hostAlias as a distinct argument
		isSSH := strings.HasPrefix(parts[1], "ssh") || strings.HasSuffix(parts[1], "/ssh")
		if !isSSH {
			continue
		}

		// Avoid matching control commands (e.g. ssh -O check / forward)
		if strings.Contains(command, "-O check") || strings.Contains(command, "-O forward") {
			continue
		}

		// Check if hostAlias is in command arguments
		hasHost := false
		for _, arg := range parts[2:] {
			if arg == hostAlias {
				hasHost = true
				break
			}
		}

		if hasHost {
			pids = append(pids, pid)
		}
	}

	return pids
}

// AttachOrDetectLiveSession inspects running SSH connections and attempts dynamic tunnel attachment if possible.
func AttachOrDetectLiveSession(hostAlias string, remotePort int, localPort int) LiveSessionResult {
	hasMaster, masterPid := CheckMasterSession(hostAlias)
	if hasMaster {
		err := DynamicallyAttachTunnel(hostAlias, remotePort, localPort)
		return LiveSessionResult{
			HasMasterSession:         true,
			MasterPID:                masterPid,
			TunnelDynamicallyAttached: err == nil,
			Error:                    err,
		}
	}

	// If no master session, check for standalone ssh processes
	pids := FindRunningSSHProcesses(hostAlias)
	return LiveSessionResult{
		HasMasterSession: false,
		RunningPIDs:      pids,
	}
}
