package agent

import (
	"context"
	"path/filepath"
	"testing"
)

func TestLookupSeesTokenFromOtherProcess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agents.json")

	// Process A: server store
	server, err := OpenFileStore(path)
	if err != nil {
		t.Fatal(err)
	}

	// Process B: CLI mints token (separate store handle, same file)
	cli, err := OpenFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	tok, err := cli.Register(context.Background(), "dashboard", "")
	if err != nil {
		t.Fatal(err)
	}

	// Server must accept without restart
	rec, err := server.LookupByToken(context.Background(), tok)
	if err != nil {
		t.Fatalf("server should see CLI token after reload: %v", err)
	}
	if rec.AgentID != "dashboard" {
		t.Fatalf("agentId %q", rec.AgentID)
	}
}
