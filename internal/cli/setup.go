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
// First run (interactive): one standard [Y/n] confirm with defaults.
// Custom path/id: use `fortmemory init`.
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

	return runFirstTimeConfirm(def, defaultID)
}

// runFirstTimeConfirm is the normal CLI first-run: show defaults, Y/n once.
func runFirstTimeConfirm(vaultPath, vaultID string) (string, error) {
	fmt.Println()
	fmt.Println("No vault configured.")
	fmt.Println()
	fmt.Printf("  Folder: %s\n", vaultPath)
	fmt.Printf("  Id:     %s\n", vaultID)
	fmt.Println()
	fmt.Println("  (Custom path/id: fortmemory init ~/path --id myid)")
	fmt.Println()
	fmt.Print("Create and start? [Y/n] ")

	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil && strings.TrimSpace(line) == "" {
		return "", err
	}
	ans := strings.ToLower(strings.TrimSpace(line))
	// Standard: empty, y, yes → accept. n/no → abort.
	if ans == "n" || ans == "no" {
		return "", fmt.Errorf("aborted — run: fortmemory init ~/Vaults/MyVault --id personal")
	}
	// anything else including y/yes/empty → proceed

	cfgPath, created, err := config.EnsureVault(vaultPath, vaultID)
	if err != nil {
		return "", err
	}
	if created {
		fmt.Printf("Created vault %s (id=%s)\n\n", vaultPath, vaultID)
	}
	return cfgPath, nil
}

// expandHome is used by init helpers if needed later.
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
