# Implementation Status

## Done

| Step | Status | Notes |
|------|--------|-------|
| config.Load + init | **Done** | TOML, `fortmemory init`, discover config |
| vault path jail | **Done** | traversal reject, `.fortmemory` reserved, atomic write |
| fortsignal.Client | **Done** | start/verify, 403 deny as decision |
| memory.Service.Write | **Done** | local policy → FS challenge → sign → verify → write |
| CLI `write` | **Done** | gated write without HTTP server |
| agent tokens | **Done** | `fortmemory token`, `agent add/list`, `.fortmemory/agents.json` |
| HTTP `serve` | **Done** | health, read, write, delete, search, receipts on 127.0.0.1 |
| FTS index + reindex | **Done** | SQLite FTS5, `fortmemory reindex` |
| Vault watcher | **Done** | fsnotify → debounced reindex (Obsidian edits) |
| Thin dashboard | **Done** | `/` Home · Search · Activity · Agents · Settings |
| MCP stdio tools | **Done** | `fortmemory mcp` — search/read/write/delete |
| Frontmatter signalId | **Done** | post-allow `last_signal_id` annotation |
| Hybrid embeddings | **Done** | optional Ollama; FTS fallback |
| CLI delete | **Done** | FortSignal-gated |
| Public repo prep | **Done** | git init, PUBLISH.md, examples |
| Integration doctor | **Done** | `fortmemory doctor` + agent list + challenge probe |
| Integration guide | **Done** | docs/INTEGRATION.md |

## Strategy

See **[STRATEGY-LOCK.md](./STRATEGY-LOCK.md)** — freeze high-level decisions; **no more Cloudflare engineering** this cycle.

## Next code

1. Dogfood on a real vault + live FortSignal key  
2. Push public repo (`docs/PUBLISH.md`)  
3. Obsidian thin plugin (status + recent signals)  
4. Release binaries / `go install` path once remote exists

## Dogfood: gated write

```bash
cd ~/projects/fortmemory-vault
go build -o bin/fortmemory ./cmd/fortmemory

# 1. Start (first run: choose YOUR vault id)
fortmemory

# 2. FortSignal (dashboard) — use YOUR vault id in recipients
#    - Agent passport → download agent-key.json
#    - Policy: memory.write, max 65536, recipients {vaultId}/Scratch/*
#    - Passkey-approve delegation

export FORTSIGNAL_API_KEY=fs_live_...   # YOUR tenant key

# 3. Local dashboard token (not FortSignal)
fortmemory token

# 4. Wire agent signing key
fortmemory agent add <agentId> --key /path/to/agent-key.json

# 5. Prove integration
fortmemory doctor --key /path/to/agent-key.json --write-probe

# 6. API (TOK from fortmemory token)
export TOK=fm_…
curl -s http://127.0.0.1:7432/v1/health
curl -s -H "Authorization: Bearer $TOK" \
  "http://127.0.0.1:7432/v1/read?path=Scratch/hello.md"
```

## Tests

```bash
go test ./...
```
