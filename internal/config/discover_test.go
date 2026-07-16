package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureVaultAndActive(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	// Avoid inheriting caller's FORTMEMORY_CONFIG
	t.Setenv("FORTMEMORY_CONFIG", "")

	vault := filepath.Join(home, "Vaults", "FortMemory")
	cfgPath, created, err := EnsureVault(vault, "personal")
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("expected created")
	}
	if _, err := os.Stat(cfgPath); err != nil {
		t.Fatal(err)
	}
	active, ok := Active()
	if !ok || active != cfgPath {
		t.Fatalf("active=%q ok=%v want %q", active, ok, cfgPath)
	}

	// Second ensure is no-op
	_, created2, err := EnsureVault(vault, "personal")
	if err != nil || created2 {
		t.Fatalf("created2=%v err=%v", created2, err)
	}
}

func TestDiscoverOrCreateUsesActive(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("FORTMEMORY_CONFIG", "")

	// cwd with no vault
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	path, created, err := DiscoverOrCreate("")
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("expected auto-create default vault")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}

	// Again: should find active, not re-create
	path2, created2, err := DiscoverOrCreate("")
	if err != nil {
		t.Fatal(err)
	}
	if created2 {
		t.Fatal("should not create again")
	}
	if path2 != path {
		t.Fatalf("%q vs %q", path2, path)
	}
}
