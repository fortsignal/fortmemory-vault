package policy

import (
	"testing"

	"github.com/fortsignal/fortmemory-vault/internal/config"
)

func TestAllowWriteList(t *testing.T) {
	e := New(config.LocalPolicyConfig{
		AllowWrite: []string{"Scratch/**", "Inbox/*"},
	})
	if !e.CheckWrite("Scratch/a.md").Allow {
		t.Fatal("expected allow Scratch")
	}
	if e.CheckWrite("Private/x.md").Allow {
		t.Fatal("expected deny Private")
	}
}

func TestDenyWriteFortMemory(t *testing.T) {
	e := New(config.LocalPolicyConfig{})
	if e.CheckWrite(".fortmemory/config.toml").Allow {
		t.Fatal("expected deny")
	}
}
