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

// Discover finds config: --config flag path, FORTMEMORY_CONFIG, or <cwd vault guess>.
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
		// Also if cwd is vault root
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("no config found (run fortmemory init or set FORTMEMORY_CONFIG)")
}
