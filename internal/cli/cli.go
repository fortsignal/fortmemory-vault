// Package cli wires fortmemory subcommands.
package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"log/slog"
	"os/signal"
	"syscall"

	"github.com/fortsignal/fortmemory-vault/internal/agent"
	"github.com/fortsignal/fortmemory-vault/internal/app"
	"github.com/fortsignal/fortmemory-vault/internal/config"
	"github.com/fortsignal/fortmemory-vault/internal/fortsignal"
	"github.com/fortsignal/fortmemory-vault/internal/index"
	"github.com/fortsignal/fortmemory-vault/internal/mcpserver"
	"github.com/fortsignal/fortmemory-vault/internal/memory"
	"github.com/fortsignal/fortmemory-vault/internal/policy"
	"github.com/fortsignal/fortmemory-vault/internal/receipts"
	"github.com/fortsignal/fortmemory-vault/internal/server"
	"github.com/fortsignal/fortmemory-vault/internal/vault"
	"github.com/fortsignal/fortmemory-vault/internal/watcher"
)

// Version is set at build time via -ldflags when releasing.
var Version = "0.0.0-dev"

// Execute runs the CLI with argv (excluding program name).
func Execute(args []string) error {
	if len(args) == 0 {
		printUsage()
		return fmt.Errorf("usage: fortmemory <command>")
	}

	switch args[0] {
	case "version", "-v", "--version":
		fmt.Println(Version)
		return nil
	case "help", "-h", "--help":
		printUsage()
		return nil
	case "init":
		return runInit(args[1:])
	case "write":
		return runWrite(args[1:])
	case "delete":
		return runDelete(args[1:])
	case "serve":
		return runServe(args[1:])
	case "reindex":
		return runReindex(args[1:])
	case "agent":
		return runAgent(args[1:])
	case "tunnel":
		return runTunnel(args[1:])
	case "cloudflare", "cf":
		return runCloudflare(args[1:])
	case "tailscale", "ts":
		return runTailscale(args[1:])
	case "mcp":
		return runMCP(args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printUsage() {
	fmt.Println(strings.TrimSpace(`
FortMemory — verifiable local agent memory

Usage:
  fortmemory version
  fortmemory init [vault-path] [--id personal] [--force]
  fortmemory write --path Scratch/note.md --body "..." [--config path] [--key agent-key.json]
  fortmemory write --path Scratch/note.md --file ./note.md [--mode overwrite]
  fortmemory delete --path Scratch/note.md [--config path] [--key agent-key.json]
  fortmemory serve [--config path]
  fortmemory reindex
  fortmemory agent add <agentId>
  fortmemory mcp --agent <id> [--config path] [--key path]
  fortmemory cloudflare install|check|config|quick|run
  fortmemory tailscale [check|print-serve]
  fortmemory tunnel cloudflare|tailscale … (aliases)

License: Apache-2.0 (open-core — see docs/OPEN-CORE.md)

Environment:
  FORTSIGNAL_API_KEY   FortSignal API key (required for write)
  FORTMEMORY_CONFIG    Path to .fortmemory/config.toml

Docs: docs/CLI.md  docs/CLOUDFLARE-TUNNEL.md  docs/OPEN-CORE.md
`))
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	id := fs.String("id", "personal", "vault_id")
	force := fs.Bool("force", false, "overwrite existing config")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	vaultPath := fs.Arg(0)
	if vaultPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		vaultPath = cwd
	}
	abs, err := filepath.Abs(vaultPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return err
	}
	// Helpful starter dirs (Obsidian-friendly)
	for _, d := range []string{"Scratch", "Inbox", "Private"} {
		_ = os.MkdirAll(filepath.Join(abs, d), 0o755)
	}
	fmDir := config.ConfigDir(abs)
	if err := os.MkdirAll(fmDir, 0o700); err != nil {
		return err
	}

	cfg := config.Default()
	cfg.VaultID = *id
	cfg.VaultPath = abs
	cfg.Agent.ID = ""
	cfg.Agent.KeyFile = ""

	cfgPath := config.ConfigFile(abs)
	if err := config.WriteDefault(cfgPath, cfg, *force); err != nil {
		return err
	}
	// Touch receipts log
	if f, err := os.OpenFile(config.ReceiptsFile(abs), os.O_CREATE|os.O_APPEND, 0o600); err == nil {
		_ = f.Close()
	}

	fmt.Printf("Initialized FortMemory vault\n")
	fmt.Printf("  vault:  %s\n", abs)
	fmt.Printf("  id:     %s\n", *id)
	fmt.Printf("  config: %s\n", cfgPath)
	fmt.Printf("\nNext:\n")
	fmt.Printf("  1. export FORTSIGNAL_API_KEY=fs_live_...\n")
	fmt.Printf("  2. Register agent + approve delegation in FortSignal dashboard\n")
	fmt.Printf("     Policy must allow action memory.write and recipients like %s/Scratch/*\n", *id)
	fmt.Printf("  3. fortmemory write --config %s --key /path/to/agent-key.json --path Scratch/hello.md --body \"# hi\"\n", cfgPath)
	return nil
}

func runWrite(args []string) error {
	fs := flag.NewFlagSet("write", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "path to config.toml")
	keyPath := fs.String("key", "", "path to agent key JSON (Deep Agents format)")
	relPath := fs.String("path", "", "relative path under vault (required)")
	body := fs.String("body", "", "file contents as string")
	file := fs.String("file", "", "read contents from file")
	mode := fs.String("mode", "overwrite", "create|append|overwrite")
	agentID := fs.String("agent", "", "agentId (default: from key file / config)")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *relPath == "" {
		return fmt.Errorf("write: --path is required")
	}
	if *body == "" && *file == "" {
		return fmt.Errorf("write: provide --body or --file")
	}
	if *body != "" && *file != "" {
		return fmt.Errorf("write: use only one of --body or --file")
	}

	var content []byte
	var err error
	if *file != "" {
		content, err = os.ReadFile(*file)
		if err != nil {
			return err
		}
	} else {
		content = []byte(*body)
	}

	discovered, err := config.Discover(*cfgPath)
	if err != nil {
		return err
	}
	cfg, err := config.Load(discovered)
	if err != nil {
		return err
	}

	keyFile := *keyPath
	if keyFile == "" {
		keyFile = cfg.Agent.KeyFile
	}
	if keyFile == "" {
		return fmt.Errorf("write: --key or config agent.key_file required")
	}
	signer, err := agent.LoadSigner(keyFile)
	if err != nil {
		return fmt.Errorf("load agent key: %w", err)
	}
	aid := *agentID
	if aid == "" {
		aid = cfg.Agent.ID
	}
	if aid == "" {
		aid = signer.AgentID()
	}

	apiKey, err := cfg.APIKey()
	if err != nil {
		return err
	}
	fsClient := fortsignal.New(apiKey, cfg.FortSignal.APIBase)

	store, err := vault.New(cfg.VaultPath)
	if err != nil {
		return err
	}
	recStore, err := receipts.OpenJSONL(config.ReceiptsFile(cfg.VaultPath))
	if err != nil {
		return err
	}
	defer recStore.Close()
	idx, err := index.Open(config.IndexFile(cfg.VaultPath))
	if err != nil {
		return err
	}
	defer idx.Close()

	svc := &memory.Service{
		Cfg:        cfg,
		Vault:      store,
		Index:      idx,
		Receipts:   recStore,
		Policy:     policy.New(cfg.Policy),
		FortSignal: fsClient,
		Signer:     signer,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	res, err := svc.Write(ctx, memory.WriteInput{
		AgentID: aid,
		Path:    *relPath,
		Content: content,
		Mode:    vault.WriteMode(*mode),
	})
	if err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(res); err != nil {
		return err
	}
	if res.Decision != "allow" {
		return fmt.Errorf("denied: %s", res.Reason)
	}
	return nil
}

func runDelete(args []string) error {
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "path to config.toml")
	keyPath := fs.String("key", "", "path to agent key JSON")
	relPath := fs.String("path", "", "relative path under vault (required)")
	agentID := fs.String("agent", "", "agentId")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *relPath == "" {
		return fmt.Errorf("delete: --path is required")
	}
	rt, err := app.Open(*cfgPath, true)
	if err != nil {
		return err
	}
	defer rt.Close()
	keyFile := *keyPath
	if keyFile == "" {
		keyFile = rt.Cfg.Agent.KeyFile
	}
	aid := *agentID
	if aid == "" {
		aid = rt.Cfg.Agent.ID
	}
	if keyFile != "" {
		if err := rt.EnsureSigner(aid, keyFile); err != nil {
			// EnsureSigner uses agentID for map key after load — load directly
			sig, err2 := agent.LoadSigner(keyFile)
			if err2 != nil {
				return err
			}
			rt.Memory.Signer = sig
			if rt.Memory.Signers == nil {
				rt.Memory.Signers = map[string]agent.Signer{}
			}
			rt.Memory.Signers[sig.AgentID()] = sig
			aid = sig.AgentID()
		} else if aid == "" {
			aid = rt.Memory.Signer.AgentID()
		}
	}
	if aid == "" && rt.Memory.Signer != nil {
		aid = rt.Memory.Signer.AgentID()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	res, err := rt.Memory.Delete(ctx, aid, *relPath)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(res)
	if res.Decision != "allow" {
		return fmt.Errorf("denied: %s", res.Reason)
	}
	return nil
}

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "path to config.toml")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}

	rt, err := app.Open(*cfgPath, true)
	if err != nil {
		return err
	}
	defer rt.Close()

	if len(rt.Signers) == 0 {
		fmt.Fprintln(os.Stderr, "warning: no agent signing keys loaded — HTTP write will fail until you run:")
		fmt.Fprintln(os.Stderr, "  fortmemory agent add <agentId> --key /path/to/agent-key.json")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	w := watcher.New(rt.Vault, rt.Index, slog.Default())
	if err := w.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "warning: vault watcher: %v\n", err)
	}

	srv := server.New(server.Deps{
		Config:   rt.Cfg,
		Version:  Version,
		Vault:    rt.Vault,
		Index:    rt.Index,
		Receipts: rt.Receipts,
		Agents:   rt.Agents,
		Memory:   rt.Memory,
	})

	fmt.Fprintf(os.Stderr, "FortMemory serving vault %q on http://%s:%d\n",
		rt.Cfg.VaultID, rt.Cfg.Bind, rt.Cfg.Port)
	fmt.Fprintf(os.Stderr, "Dashboard: http://%s:%d/\n", rt.Cfg.Bind, rt.Cfg.Port)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		return nil
	case err := <-errCh:
		return err
	}
}

func runMCP(args []string) error {
	fs := flag.NewFlagSet("mcp", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "path to config.toml")
	agentID := fs.String("agent", "", "agentId (must have signing key loaded)")
	keyPath := fs.String("key", "", "optional agent key JSON override")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *agentID == "" {
		return fmt.Errorf("mcp: --agent is required")
	}

	rt, err := app.Open(*cfgPath, true)
	if err != nil {
		return err
	}
	defer rt.Close()

	if *keyPath != "" {
		if err := rt.EnsureSigner(*agentID, *keyPath); err != nil {
			return err
		}
	}
	// Ensure agent id is in signers (key may register under file agentId)
	if _, ok := rt.Signers[*agentID]; !ok {
		// try loading from agents store
		if rec, err := rt.Agents.Get(context.Background(), *agentID); err == nil && rec.KeyPath != "" {
			if err := rt.EnsureSigner(*agentID, rec.KeyPath); err != nil {
				return err
			}
		}
	}
	if _, ok := rt.Signers[*agentID]; !ok {
		// maybe key file's agentId matches
		found := false
		for id := range rt.Signers {
			if id == *agentID {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("no signing key for agent %q (agent add --key or --key flag)", *agentID)
		}
	}

	fmt.Fprintf(os.Stderr, "fortmemory mcp: agent=%s vault=%s (stdio)\n", *agentID, rt.Cfg.VaultID)
	return mcpserver.RunStdio(context.Background(), mcpserver.Deps{
		Memory:  rt.Memory,
		AgentID: *agentID,
		Version: Version,
	})
}

func runReindex(args []string) error {
	fs := flag.NewFlagSet("reindex", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "path to config.toml")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	rt, err := app.Open(*cfgPath, false)
	if err != nil {
		return err
	}
	defer rt.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	n, err := rt.Memory.Reindex(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("reindexed %d markdown files\n", n)
	return nil
}

func runAgent(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: fortmemory agent add <agentId> [--key path] [--config path]")
	}
	switch args[0] {
	case "add":
		return runAgentAdd(args[1:])
	case "list":
		return runAgentList(args[1:])
	default:
		return fmt.Errorf("unknown agent subcommand %q (add|list)", args[0])
	}
}

func runAgentAdd(args []string) error {
	// Support both:
	//   fortmemory agent add research-01 --config … --key …
	//   fortmemory agent add --config … --key … research-01
	// (stdlib flag stops at first non-flag, so we peel the id out first.)
	agentID, flagArgs := peelPositionalID(args)
	fs := flag.NewFlagSet("agent add", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "path to config.toml")
	keyPath := fs.String("key", "", "path to FortSignal agent key JSON (for server writes)")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if agentID == "" {
		agentID = fs.Arg(0)
	}
	if agentID == "" {
		return fmt.Errorf("agent add: agentId required")
	}

	discovered, err := config.Discover(*cfgPath)
	if err != nil {
		return err
	}
	cfg, err := config.Load(discovered)
	if err != nil {
		return err
	}
	store, err := agent.OpenFileStore(config.AgentsFile(cfg.VaultPath))
	if err != nil {
		return err
	}
	key := *keyPath
	if key == "" {
		key = cfg.Agent.KeyFile
	}
	if key != "" {
		// validate key loads and agentId matches if possible
		sig, err := agent.LoadSigner(key)
		if err != nil {
			return fmt.Errorf("key: %w", err)
		}
		if sig.AgentID() != agentID {
			fmt.Fprintf(os.Stderr, "warning: key file agentId %q != %q (using key file id for signing)\n", sig.AgentID(), agentID)
		}
	}
	tok, err := store.Register(context.Background(), agentID, key)
	if err != nil {
		return err
	}
	fmt.Printf("Agent registered: %s\n", agentID)
	if key != "" {
		fmt.Printf("Signing key:     %s\n", key)
	}
	fmt.Printf("API token (save now — shown once):\n\n  %s\n\n", tok)
	fmt.Printf("Example:\n  curl -H \"Authorization: Bearer %s\" http://%s:%d/v1/health\n",
		tok, cfg.Bind, cfg.Port)
	fmt.Printf("  curl -H \"Authorization: Bearer %s\" \"http://%s:%d/v1/read?path=Scratch/hello.md\"\n",
		tok, cfg.Bind, cfg.Port)
	return nil
}

// peelPositionalID extracts the first non-flag token as an id; returns remaining args for flag.Parse.
func peelPositionalID(args []string) (id string, rest []string) {
	rest = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			rest = append(rest, args[i:]...)
			break
		}
		if strings.HasPrefix(a, "-") {
			rest = append(rest, a)
			// keep flag values that don't start with -
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") && !strings.Contains(a, "=") {
				// boolean flags have no value — our flags all take values except none here
				// all our flags take values
				i++
				if i < len(args) {
					rest = append(rest, args[i])
				}
			}
			continue
		}
		if id == "" {
			id = a
			continue
		}
		rest = append(rest, a)
	}
	return id, rest
}

func runAgentList(args []string) error {
	fs := flag.NewFlagSet("agent list", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "path to config.toml")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	discovered, err := config.Discover(*cfgPath)
	if err != nil {
		return err
	}
	cfg, err := config.Load(discovered)
	if err != nil {
		return err
	}
	store, err := agent.OpenFileStore(config.AgentsFile(cfg.VaultPath))
	if err != nil {
		return err
	}
	list, err := store.List(context.Background())
	if err != nil {
		return err
	}
	if len(list) == 0 {
		fmt.Println("(no agents)")
		return nil
	}
	for _, a := range list {
		key := a.KeyPath
		if key == "" {
			key = "-"
		}
		fmt.Printf("%s\tkey=%s\n", a.AgentID, key)
	}
	return nil
}
