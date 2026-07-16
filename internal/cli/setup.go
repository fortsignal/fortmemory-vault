package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fortsignal/fortmemory-vault/internal/config"
)

func isInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// resolveServeConfig finds a vault for start/serve.
// First run (interactive): default folder + ask for id (default personal).
func resolveServeConfig(explicit string) (string, error) {
	if path, err := config.Discover(explicit); err == nil {
		_ = config.SetActive(path)
		return path, nil
	} else if explicit != "" || os.Getenv("FORTMEMORY_CONFIG") != "" {
		return "", err
	}

	def, err := config.DefaultVaultPath()
	if err != nil {
		return "", err
	}
	const defaultID = "personal"

	// Non-interactive / CI: create defaults, no questions.
	if !isInteractive() || os.Getenv("FORTMEMORY_YES") != "" {
		cfgPath, created, err := config.EnsureVault(def, defaultID)
		if err != nil {
			return "", err
		}
		if created {
			fmt.Fprintf(os.Stderr, "Created vault %s (id=%s)\n", def, defaultID)
		}
		return cfgPath, nil
	}

	return runFirstTimeSetup(def, defaultID)
}

// runFirstTimeSetup: standard defaults, one simple id prompt.
func runFirstTimeSetup(vaultPath, defaultID string) (string, error) {
	fmt.Println()
	fmt.Println("No vault yet — quick setup.")
	fmt.Println()
	fmt.Printf("  Folder: %s\n", vaultPath)
	fmt.Println()
	fmt.Printf("Vault id [%s]: ", defaultID)
	fmt.Print("") // keep prompt clean

	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil && strings.TrimSpace(line) == "" {
		return "", err
	}
	id := sanitizeVaultID(strings.TrimSpace(line))
	if id == "" || id == "y" || id == "yes" {
		id = defaultID
	}

	fmt.Println()
	fmt.Printf("Creating vault id=%s …\n", id)

	cfgPath, created, err := config.EnsureVault(vaultPath, id)
	if err != nil {
		return "", err
	}
	if created {
		fmt.Printf("Ready. FortSignal write paths look like: %s/Scratch/*\n\n", id)
	}
	return cfgPath, nil
}

func expandHome(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", fmt.Errorf("empty path")
	}
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		p = filepath.Join(home, p[2:])
	} else if p == "~" {
		return os.UserHomeDir()
	}
	return filepath.Abs(p)
}

func sanitizeVaultID(id string) string {
	id = strings.TrimSpace(id)
	id = strings.ToLower(id)
	var b strings.Builder
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteByte('-')
		}
	}
	return b.String()
}
