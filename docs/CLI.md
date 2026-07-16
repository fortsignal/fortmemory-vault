# CLI Surface

Binary name: `fortmemory`

## Commands (target)

### `fortmemory version`

Print version and commit if available.

### `fortmemory init [path]`

Initialize a vault for FortMemory.

```bash
fortmemory init ~/Vaults/Personal
fortmemory init ~/Vaults/Personal --id personal --force
```

Behavior:

- Create path if missing  
- Create `Scratch/`, `Inbox/`, `Private/`  
- Create `.fortmemory/config.toml` defaults  
- Create empty `receipts.jsonl`  
- Refuse to overwrite existing config without `--force`  

### `fortmemory write` (implemented)

Governed write via FortSignal (no HTTP server required).

```bash
export FORTSIGNAL_API_KEY=fs_live_...
fortmemory write \
  --config ~/Vaults/Personal/.fortmemory/config.toml \
  --key ~/agent-key.json \
  --path Scratch/note.md \
  --body "# hello" \
  --mode overwrite
```

Flags: `--config`, `--key`, `--path`, `--body` | `--file`, `--mode`, `--agent`  
Stdout: JSON `MutateResult` (`decision`, `signalId`, …). Non-zero exit on deny.  

### `fortmemory serve`

Run the local memory server.

```bash
fortmemory serve
fortmemory serve --config ~/Vaults/Personal/.fortmemory/config.toml
fortmemory serve --port 7432 --bind 127.0.0.1
```

Flags:

| Flag | Default | Notes |
|------|---------|-------|
| `--config` | discover from cwd / env | Path to config |
| `--bind` | `127.0.0.1` | Dangerous if `0.0.0.0` |
| `--port` | `7432` | |
| `--vault` | from config | Override vault root |

### `fortmemory reindex`

Rebuild FTS/vector index from vault files.

```bash
fortmemory reindex
fortmemory reindex --full
```

### `fortmemory agent add <agentId>` (Phase 1.5)

Issue a local API token mapped to FortSignal `agentId`.

```bash
fortmemory agent add research-01
# prints token once
```

### `fortmemory cloudflare` — Cloudflare Tunnel plugin (primary)

```bash
fortmemory cloudflare install
fortmemory cloudflare check
fortmemory cloudflare config --hostname memory.example.com
fortmemory cloudflare quick
fortmemory cloudflare run --name fortmemory --cf-config ~/.cloudflared/config-fortmemory.yml
```

Alias: `fortmemory tunnel cloudflare …`  
Docs: [CLOUDFLARE-TUNNEL.md](./CLOUDFLARE-TUNNEL.md)

### `fortmemory tailscale` — supported remote (mesh)

```bash
fortmemory tailscale                 # status + guide
fortmemory tailscale check
fortmemory tailscale print-serve     # → tailscale serve --bg http://127.0.0.1:7432
```

Docs: [TAILSCALE.md](./TAILSCALE.md) · [REMOTE-ACCESS.md](./REMOTE-ACCESS.md)

## Config sketch (`.fortmemory/config.toml`)

```toml
vault_id = "personal"
vault_path = "/home/user/Vaults/Personal"
bind = "127.0.0.1"
port = 7432

[fortsignal]
api_base = "https://api.fortsignal.com"
# api_key from env FORTSIGNAL_API_KEY preferred
api_key_env = "FORTSIGNAL_API_KEY"

[embeddings]
provider = "ollama"          # ollama | none
ollama_url = "http://127.0.0.1:11434"
model = "nomic-embed-text"

[policy]
# optional local globs; FortSignal still authoritative for crypto
# allow_write = ["Scratch/**", "Inbox/**"]
# deny_read = ["Private/**"]

[security]
fail_closed_on_fortsignal = true
allow_ungated_reads = true
```

## Environment variables

| Var | Purpose |
|-----|---------|
| `FORTSIGNAL_API_KEY` | Tenant API key |
| `FORTMEMORY_CONFIG` | Config path override |
| `FORTMEMORY_TOKEN` | Optional default client token for tooling |

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Generic failure |
| 2 | Config / usage error |
| 3 | FortSignal / dependency failure |

## Operator loop (dogfood)

```bash
export FORTSIGNAL_API_KEY=fs_live_...
fortmemory init ~/Vaults/Personal
# configure agent passport + delegation in FortSignal dashboard
fortmemory agent add research-01
fortmemory serve
# other terminal:
curl -s localhost:7432/v1/health | jq
```
