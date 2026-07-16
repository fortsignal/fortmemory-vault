// Package config loads vault-local and process configuration for FortMemory.
//
// Primary file: <vault>/.fortmemory/config.toml
// Secrets: prefer environment variables (FORTSIGNAL_API_KEY) over plaintext files.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// DefaultBind and DefaultPort are the safe local-first listen defaults.
const (
	DefaultBind       = "127.0.0.1"
	DefaultPort       = 7432
	ConfigDirName     = ".fortmemory"
	ConfigFileName    = "config.toml"
	ReceiptsFileName  = "receipts.jsonl"
	DefaultAPIBase    = "https://api.fortsignal.com"
	DefaultAPIKeyEnv  = "FORTSIGNAL_API_KEY"
)

// Config is the process configuration for one vault instance.
type Config struct {
	VaultID   string `toml:"vault_id"`
	VaultPath string `toml:"vault_path"`
	Bind      string `toml:"bind"`
	Port      int    `toml:"port"`

	FortSignal FortSignalConfig  `toml:"fortsignal"`
	Agent      AgentConfig       `toml:"agent"`
	Embeddings EmbeddingsConfig  `toml:"embeddings"`
	Policy     LocalPolicyConfig `toml:"policy"`
	Security   SecurityConfig    `toml:"security"`

	// ConfigPath is the absolute path of the loaded config file (not serialized).
	ConfigPath string `toml:"-"`
}

// FortSignalConfig controls the enforcement plane client.
type FortSignalConfig struct {
	APIBase   string `toml:"api_base"`
	APIKeyEnv string `toml:"api_key_env"`
}

// AgentConfig is the default local-signer agent for CLI writes.
type AgentConfig struct {
	ID      string `toml:"id"`
	KeyFile string `toml:"key_file"` // Deep Agents style JSON key file
}

// EmbeddingsConfig selects optional local embedding provider.
type EmbeddingsConfig struct {
	Provider  string `toml:"provider"` // ollama | none
	OllamaURL string `toml:"ollama_url"`
	Model     string `toml:"model"`
}

// LocalPolicyConfig is FortMemory-side path policy (ADR-010).
type LocalPolicyConfig struct {
	AllowWrite []string `toml:"allow_write"`
	DenyRead   []string `toml:"deny_read"`
	DenyWrite  []string `toml:"deny_write"`
}

// SecurityConfig hardens runtime behaviour.
type SecurityConfig struct {
	FailClosedOnFortSignal bool `toml:"fail_closed_on_fortsignal"`
	AllowUngatedReads      bool `toml:"allow_ungated_reads"`
}

// Default returns a safe starter config (paths still required).
func Default() Config {
	return Config{
		Bind: DefaultBind,
		Port: DefaultPort,
		FortSignal: FortSignalConfig{
			APIBase:   DefaultAPIBase,
			APIKeyEnv: DefaultAPIKeyEnv,
		},
		Embeddings: EmbeddingsConfig{
			Provider:  "none",
			OllamaURL: "http://127.0.0.1:11434",
			Model:     "nomic-embed-text",
		},
		Policy: LocalPolicyConfig{
			// Empty allow_write = allow all (except .fortmemory and deny_write).
			DenyWrite: []string{".fortmemory/**", ".fortmemory/*"},
			DenyRead:  []string{"Private/**"},
		},
		Security: SecurityConfig{
			FailClosedOnFortSignal: true,
			AllowUngatedReads:      true,
		},
	}
}

// ConfigDir returns <vault>/.fortmemory.
func ConfigDir(vaultPath string) string {
	return filepath.Join(vaultPath, ConfigDirName)
}

// ConfigFile returns <vault>/.fortmemory/config.toml.
func ConfigFile(vaultPath string) string {
	return filepath.Join(ConfigDir(vaultPath), ConfigFileName)
}

// ReceiptsFile returns <vault>/.fortmemory/receipts.jsonl.
func ReceiptsFile(vaultPath string) string {
	return filepath.Join(ConfigDir(vaultPath), ReceiptsFileName)
}

// AgentsFile returns <vault>/.fortmemory/agents.json.
func AgentsFile(vaultPath string) string {
	return filepath.Join(ConfigDir(vaultPath), "agents.json")
}

// IndexFile returns <vault>/.fortmemory/index.sqlite.
func IndexFile(vaultPath string) string {
	return filepath.Join(ConfigDir(vaultPath), "index.sqlite")
}

// APIKey resolves the FortSignal API key from the environment.
func (c Config) APIKey() (string, error) {
	env := c.FortSignal.APIKeyEnv
	if env == "" {
		env = DefaultAPIKeyEnv
	}
	key := strings.TrimSpace(os.Getenv(env))
	if key == "" {
		return "", fmt.Errorf("missing %s (FortSignal API key)", env)
	}
	return key, nil
}

// Validate checks required fields after load/init.
func (c Config) Validate() error {
	if strings.TrimSpace(c.VaultID) == "" {
		return fmt.Errorf("vault_id is required")
	}
	if strings.TrimSpace(c.VaultPath) == "" {
		return fmt.Errorf("vault_path is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be 1-65535")
	}
	if c.Bind == "" {
		return fmt.Errorf("bind is required")
	}
	return nil
}

// Load reads and validates config from path.
func Load(path string) (Config, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return Config{}, err
	}
	cfg := Default()
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	cfg.ConfigPath = abs
	if cfg.FortSignal.APIBase == "" {
		cfg.FortSignal.APIBase = DefaultAPIBase
	}
	if cfg.FortSignal.APIKeyEnv == "" {
		cfg.FortSignal.APIKeyEnv = DefaultAPIKeyEnv
	}
	if cfg.Bind == "" {
		cfg.Bind = DefaultBind
	}
	if cfg.Port == 0 {
		cfg.Port = DefaultPort
	}
	// Resolve vault_path relative to config file directory's parent if not absolute.
	if cfg.VaultPath != "" && !filepath.IsAbs(cfg.VaultPath) {
		cfg.VaultPath = filepath.Join(filepath.Dir(filepath.Dir(abs)), cfg.VaultPath)
	}
	cfg.VaultPath, err = filepath.Abs(cfg.VaultPath)
	if err != nil {
		return Config{}, err
	}
	if cfg.Agent.KeyFile != "" && !filepath.IsAbs(cfg.Agent.KeyFile) {
		cfg.Agent.KeyFile = filepath.Join(filepath.Dir(abs), cfg.Agent.KeyFile)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// WriteDefault writes config.toml (does not overwrite unless force).
func WriteDefault(path string, cfg Config, force bool) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if !force {
		if _, err := os.Stat(abs); err == nil {
			return fmt.Errorf("config already exists: %s (use --force)", abs)
		}
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o700); err != nil {
		return err
	}
	// Serialize without ConfigPath.
	out := cfg
	out.ConfigPath = ""
	data, err := toml.Marshal(out)
	if err != nil {
		return err
	}
	return os.WriteFile(abs, data, 0o600)
}

// DefaultVaultPath is the zero-config vault: ~/Vaults/FortMemory
func DefaultVaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Vaults", "FortMemory"), nil
}

// userStateDir holds the "last used vault" pointer so bare `fortmemory` works.
func userStateDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "fortmemory"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "fortmemory"), nil
}

// ActiveConfigFile is ~/.config/fortmemory/active (one line: path to config.toml).
func ActiveConfigFile() (string, error) {
	dir, err := userStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "active"), nil
}

// SetActive remembers which vault to use when the user types bare `fortmemory`.
func SetActive(configPath string) error {
	abs, err := filepath.Abs(configPath)
	if err != nil {
		return err
	}
	af, err := ActiveConfigFile()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(af), 0o700); err != nil {
		return err
	}
	return os.WriteFile(af, []byte(abs+"\n"), 0o600)
}

// Active returns the last active config path, if the file exists.
func Active() (string, bool) {
	af, err := ActiveConfigFile()
	if err != nil {
		return "", false
	}
	raw, err := os.ReadFile(af)
	if err != nil {
		return "", false
	}
	p := strings.TrimSpace(string(raw))
	if p == "" {
		return "", false
	}
	if st, err := os.Stat(p); err != nil || st.IsDir() {
		return "", false
	}
	return p, true
}

// EnsureVault creates a vault at vaultPath if missing and returns its config.toml path.
func EnsureVault(vaultPath, vaultID string) (configPath string, created bool, err error) {
	if vaultID == "" {
		vaultID = "personal"
	}
	abs, err := filepath.Abs(vaultPath)
	if err != nil {
		return "", false, err
	}
	cfgPath := ConfigFile(abs)
	if st, err := os.Stat(cfgPath); err == nil && !st.IsDir() {
		_ = SetActive(cfgPath)
		return cfgPath, false, nil
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return "", false, err
	}
	for _, d := range []string{"Scratch", "Inbox", "Private"} {
		_ = os.MkdirAll(filepath.Join(abs, d), 0o755)
	}
	if err := os.MkdirAll(ConfigDir(abs), 0o700); err != nil {
		return "", false, err
	}
	cfg := Default()
	cfg.VaultID = vaultID
	cfg.VaultPath = abs
	if err := WriteDefault(cfgPath, cfg, false); err != nil {
		// race: already exists
		if st, e2 := os.Stat(cfgPath); e2 == nil && !st.IsDir() {
			_ = SetActive(cfgPath)
			return cfgPath, false, nil
		}
		return "", false, err
	}
	if f, err := os.OpenFile(ReceiptsFile(abs), os.O_CREATE|os.O_APPEND, 0o600); err == nil {
		_ = f.Close()
	}
	if err := SetActive(cfgPath); err != nil {
		return cfgPath, true, err
	}
	return cfgPath, true, nil
}

// Discover finds config in this order:
//  1. explicit --config path
//  2. FORTMEMORY_CONFIG
//  3. walk up from cwd for .fortmemory/config.toml
//  4. last active vault (~/.config/fortmemory/active)
//  5. default vault ~/Vaults/FortMemory if already initialized
func Discover(explicit string) (string, error) {
	if explicit != "" {
		return filepath.Abs(explicit)
	}
	if env := os.Getenv("FORTMEMORY_CONFIG"); env != "" {
		return filepath.Abs(env)
	}
	// Walk up from cwd looking for .fortmemory/config.toml
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := ConfigFile(dir)
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if p, ok := Active(); ok {
		return p, nil
	}
	if def, err := DefaultVaultPath(); err == nil {
		candidate := ConfigFile(def)
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("no config found (run fortmemory init or just fortmemory to create a default vault)")
}

// DiscoverOrCreate is for start/serve: always resolves a config, creating the
// default vault (~/Vaults/FortMemory) if nothing is configured yet.
func DiscoverOrCreate(explicit string) (path string, created bool, err error) {
	path, err = Discover(explicit)
	if err == nil {
		return path, false, nil
	}
	if explicit != "" || os.Getenv("FORTMEMORY_CONFIG") != "" {
		// User pointed at something specific — don't invent a different vault.
		return "", false, err
	}
	def, err := DefaultVaultPath()
	if err != nil {
		return "", false, err
	}
	return EnsureVault(def, "personal")
}
