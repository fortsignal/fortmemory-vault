// Package app wires FortMemory subsystems for CLI commands.
package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/fortsignal/fortmemory-vault/internal/agent"
	"github.com/fortsignal/fortmemory-vault/internal/config"
	"github.com/fortsignal/fortmemory-vault/internal/embed"
	"github.com/fortsignal/fortmemory-vault/internal/fortsignal"
	"github.com/fortsignal/fortmemory-vault/internal/index"
	"github.com/fortsignal/fortmemory-vault/internal/memory"
	"github.com/fortsignal/fortmemory-vault/internal/policy"
	"github.com/fortsignal/fortmemory-vault/internal/receipts"
	"github.com/fortsignal/fortmemory-vault/internal/vault"
)

// Runtime is a fully wired local instance.
type Runtime struct {
	Cfg      config.Config
	Vault    *vault.Store
	Index    *index.SQLite
	Receipts *receipts.JSONL
	Agents   *agent.FileStore
	Memory   *memory.Service
	Signers  map[string]agent.Signer
}

// Open loads config and opens stores. If requireAPIKey is true, FortSignal key must be set.
func Open(cfgPath string, requireAPIKey bool) (*Runtime, error) {
	path, err := config.Discover(cfgPath)
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}

	store, err := vault.New(cfg.VaultPath)
	if err != nil {
		return nil, err
	}
	idx, err := index.Open(config.IndexFile(cfg.VaultPath))
	if err != nil {
		return nil, fmt.Errorf("index: %w", err)
	}
	// Optional hybrid embeddings (Ollama). Failures degrade to FTS-only.
	if strings.EqualFold(cfg.Embeddings.Provider, "ollama") {
		ol := embed.NewOllama(cfg.Embeddings.OllamaURL, cfg.Embeddings.Model)
		if ol.Available(context.Background()) {
			idx.SetEmbedder(ol)
		}
	}
	recs, err := receipts.OpenJSONL(config.ReceiptsFile(cfg.VaultPath))
	if err != nil {
		_ = idx.Close()
		return nil, err
	}
	agents, err := agent.OpenFileStore(config.AgentsFile(cfg.VaultPath))
	if err != nil {
		_ = idx.Close()
		return nil, err
	}

	signers := map[string]agent.Signer{}
	// Default key from config
	if cfg.Agent.KeyFile != "" {
		sig, err := agent.LoadSigner(cfg.Agent.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("load config agent key: %w", err)
		}
		signers[sig.AgentID()] = sig
	}
	// Keys from registered agents
	list, _ := agents.List(context.Background())
	for _, rec := range list {
		if rec.KeyPath == "" {
			continue
		}
		if _, ok := signers[rec.AgentID]; ok {
			continue
		}
		sig, err := agent.LoadSigner(rec.KeyPath)
		if err != nil {
			// skip broken keys but warn via error only if that agent is used
			continue
		}
		if sig.AgentID() != rec.AgentID {
			// still register under file agentId if mismatch? prefer file identity
			signers[sig.AgentID()] = sig
			continue
		}
		signers[rec.AgentID] = sig
	}

	var fsClient fortsignal.Enforcer
	if requireAPIKey {
		key, err := cfg.APIKey()
		if err != nil {
			_ = idx.Close()
			return nil, err
		}
		fsClient = fortsignal.New(key, cfg.FortSignal.APIBase)
	} else if key, err := cfg.APIKey(); err == nil {
		fsClient = fortsignal.New(key, cfg.FortSignal.APIBase)
	}

	var defaultSigner agent.Signer
	if cfg.Agent.ID != "" {
		defaultSigner = signers[cfg.Agent.ID]
	}
	if defaultSigner == nil && len(signers) == 1 {
		for _, s := range signers {
			defaultSigner = s
		}
	}

	svc := &memory.Service{
		Cfg:        cfg,
		Vault:      store,
		Index:      idx,
		Receipts:   recs,
		Policy:     policy.New(cfg.Policy),
		FortSignal: fsClient,
		Signer:     defaultSigner,
		Signers:    signers,
	}

	return &Runtime{
		Cfg:      cfg,
		Vault:    store,
		Index:    idx,
		Receipts: recs,
		Agents:   agents,
		Memory:   svc,
		Signers:  signers,
	}, nil
}

// Close releases resources.
func (rt *Runtime) Close() error {
	if rt.Index != nil {
		return rt.Index.Close()
	}
	return nil
}

// EnsureSigner loads key for agentID if missing.
func (rt *Runtime) EnsureSigner(agentID, keyPath string) error {
	if _, ok := rt.Signers[agentID]; ok {
		return nil
	}
	if keyPath == "" {
		return fmt.Errorf("no key for agent %q", agentID)
	}
	sig, err := agent.LoadSigner(keyPath)
	if err != nil {
		return err
	}
	rt.Signers[sig.AgentID()] = sig
	if rt.Memory.Signers == nil {
		rt.Memory.Signers = map[string]agent.Signer{}
	}
	rt.Memory.Signers[sig.AgentID()] = sig
	if rt.Memory.Signer == nil {
		rt.Memory.Signer = sig
	}
	return nil
}
