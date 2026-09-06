package template

import (
	"strings"
	"testing"
)

func TestGenerateWrapperZed(t *testing.T) {
	cfg := WrapperConfig{
		EditorName:     "Zed",
		CommandName:    "zr",
		HostAlias:      "dev-box",
		LocalUser:      "myuser",
		ConnectHost:    "127.0.0.1",
		ConnectPort:    22222,
		SSHKeyPath:     "~/.ssh/rlink_id_ed25519",
		ConnectionMode: ModeReverseTunnel,
		SyntaxPattern:  `zed "ssh://{{.Host}}{{.Path}}"`,
	}

	script, err := GenerateWrapper(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(script, "#!/bin/sh") {
		t.Errorf("script must start with #!/bin/sh")
	}

	if !strings.Contains(script, `HOST_ALIAS="dev-box"`) {
		t.Errorf("script missing HOST_ALIAS")
	}

	if !strings.Contains(script, `CONNECT_PORT="22222"`) {
		t.Errorf("script missing CONNECT_PORT")
	}

	if !strings.Contains(script, `zed "ssh://__RLINK_HOST____RLINK_PATH__"`) {
		t.Errorf("script missing syntax pattern")
	}
}
