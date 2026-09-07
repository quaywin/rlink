package remote

import (
	"testing"
)

func TestCheckMasterSessionNonExistent(t *testing.T) {
	hasMaster, pid := CheckMasterSession("non_existent_host_alias_99999")
	if hasMaster {
		t.Errorf("expected no master session for non-existent host, got true")
	}
	if pid != 0 {
		t.Errorf("expected PID 0, got %d", pid)
	}
}

func TestFindRunningSSHProcesses(t *testing.T) {
	// Should not crash and should return empty slice for non-existent host
	pids := FindRunningSSHProcesses("non_existent_host_xyz_12345")
	if len(pids) != 0 {
		t.Errorf("expected 0 running processes, got %v", pids)
	}
}

func TestAttachOrDetectLiveSessionNonExistent(t *testing.T) {
	result := AttachOrDetectLiveSession("non_existent_host_alias_88888", 22222, 22)
	if result.HasMasterSession {
		t.Errorf("expected HasMasterSession to be false")
	}
	if result.TunnelDynamicallyAttached {
		t.Errorf("expected TunnelDynamicallyAttached to be false")
	}
}

func TestPidRegexParsing(t *testing.T) {
	sampleOutput := "Master running (pid=12345)\n"
	matches := pidRegex.FindStringSubmatch(sampleOutput)
	if len(matches) < 2 {
		t.Fatalf("expected pid match, got %v", matches)
	}
	if matches[1] != "12345" {
		t.Errorf("expected '12345', got %s", matches[1])
	}
}
