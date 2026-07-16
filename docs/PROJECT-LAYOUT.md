# Go Project Layout

Target module layout for implementation (not yet scaffolded in repo).

```
fortmemory-vault/
├── README.md
├── Makefile
├── go.mod
├── docs/                         # design + API alignment
├── cmd/
│   └── fortmemory/
│       └── main.go               # CLI entrypoint
├── internal/
│   ├── cli/                      # subcommands
│   ├── config/                   # load/validate config
│   ├── server/                   # HTTP API + embedded dashboard
│   ├── memory/                   # application service (orchestrates mutates)
│   ├── vault/                    # path jail, read/write/delete Markdown
│   ├── index/                    # SQLite FTS + vectors
│   ├── fortsignal/               # HTTP client (mirrors SDK agent API)
│   ├── policy/                   # local path/tag policy evaluation
│   ├── agent/                    # local tokens + optional Ed25519 signer
│   ├── receipts/                 # local receipt store
│   ├── embed/                    # Ollama client + queue
│   ├── auth/                     # bearer middleware
│   └── tunnel/                   # Phase 3 helpers
├── web/                          # static dashboard assets (embed later)
├── scripts/
│   └── demo-write.sh
└── bin/                          # build output (gitignored)
```

**Scaffold status:** packages exist as stubs with interfaces; `fortsignal.EncodeRecipient` + tests implemented.

## Package responsibilities

| Package | Responsibility | Must not |
|---------|----------------|----------|
| `cmd/fortmemory` | Wire CLI commands | Business logic |
| `server` | HTTP adapters | Talk SQLite directly if avoidable |
| `vault` | FS ops + path jail | Call FortSignal |
| `index` | Search structures | Enforce FortSignal |
| `fortsignal` | Remote enforce client | Write vault files |
| `policy` | Local globs/sensitivity | Crypto verify |
| `receipts` | Persist decisions | Mutate notes |
| `agent` | Local API credentials | Hold FortSignal private keys for all agents (agents sign themselves) |

## Signing architecture note

**Preferred:** FortMemory asks the **agent process** to sign (agent holds Ed25519 key), or FortMemory holds a per-agent key only if the agent is co-located and configured that way.

Two supported modes (document clearly):

1. **Sidecar sign (ideal):** API accepts pre-produced FortSignal verify payload from agent  
2. **Local signer (convenient dogfood):** FortMemory loads agent key from path in config and signs on behalf of local agents  

MVP may implement (2) for single-machine dogfood; design packages so (1) is clean.

## Suggested first implementation order

1. `config` + `cmd` init/serve skeleton  
2. `vault` path jail + read/write  
3. `index` FTS only  
4. `server` health/search/read  
5. `fortsignal` client + `write` gate  
6. `receipts`  
7. embeddings optional  
8. delete + agent tokens  

## Testing layout

```
internal/vault/vault_test.go
internal/index/index_test.go
internal/server/server_test.go
internal/fortsignal/client_test.go   # httptest
```

Integration tests behind build tag or env:

```bash
FORTSIGNAL_API_KEY=... go test ./... -tags=integration
```

## Coding standards (lightweight)

- Context on all I/O  
- No global mutable vault handles without mutex  
- Errors wrapped with `%w`  
- slog for logs  
- Table-driven tests  
