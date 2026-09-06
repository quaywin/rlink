package detector

import (
	"net"
	"testing"
)

func TestCheckLocalSSHServer(t *testing.T) {
	// Test on an ephemeral listening port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	status := CheckLocalSSHServer(port)
	if !status.IsListening {
		t.Errorf("expected port %d to be listening", port)
	}

	// Test on an unlikely open port
	closedStatus := CheckLocalSSHServer(59999)
	if closedStatus.IsListening {
		t.Errorf("expected port 59999 to NOT be listening")
	}
	if closedStatus.HelpGuide == "" {
		t.Errorf("expected help guide to be non-empty when port is closed")
	}
}
