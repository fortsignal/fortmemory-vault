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

# 1. Init vault
./bin/fortmemory init ~/Vaults/Personal --id personal

# 2. FortSignal setup (dashboard)
#    - Register agent passport, download agent-key.json
#    - Policy allowedActions includes memory.write
#    - allowedRecipients e.g. personal/Scratch/*
#    - Approve delegation (passkey)

export FORTSIGNAL_API_KEY=fs_live_...

# 3. Local dashboard token (not FortSignal)
fortmemory token
# paste fm_… into dashboard Bearer field

# 4. Optional: FortSignal agent for governed writes
fortmemory agent add research-01 \
  --key ~/path/to/agent-key.json

# 5. Start (or already running)
fortmemory

# 6. API
export TOK=fm_…   # from fortmemory token
curl -s http://127.0.0.1:7432/v1/health
curl -s -H "Authorization: Bearer $TOK" \
  "http://127.0.0.1:7432/v1/read?path=Scratch/hello.md"
curl -s -H "Authorization: Bearer $TOK" -H 'Content-Type: application/json' \
  -d '{"q":"hello","topK":5}' http://127.0.0.1:7432/v1/search

# 6. Governed write (CLI or HTTP POST /v1/write)
./bin/fortmemory write \
  --config ~/Vaults/Personal/.fortmemory/config.toml \
  --key ~/path/to/agent-key.json \
  --path Scratch/hello.md \
  --body $'# Hello from FortMemory\n'
```

## Tests

```bash
go test ./...
```
