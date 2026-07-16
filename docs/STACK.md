# Stack Decision — Go Local Memory Server

## Decision

**Implement the FortMemory local memory server in Go (Golang).**

Status: **Accepted** (see [DECISIONS.md](./DECISIONS.md) ADR-001).

## Recommendation summary

Go is the best default for a **security-first, local-first background daemon** that solo-founds well:

| Criterion | Go | TypeScript | Rust |
|-----------|----|------------|------|
| Single static-ish binary | Excellent | Needs packaging | Excellent |
| Daemon / CLI ergonomics | Excellent | Good | Good |
| fsnotify + HTTP + SQLite | Mature | Mature | Mature |
| mTLS / networking | Excellent | Good | Excellent |
| Solo iteration speed | High | Highest if already in TS | Lower if not fluent |
| “Feels like infrastructure” | Strong | Weaker for local security daemon | Strongest |
| FortSignal integration | HTTPS JSON (fine) | Official `@fortsignal/sdk` | HTTPS JSON (fine) |

**Why not TypeScript for the daemon?**  
TS is excellent for product UIs and agents. A long-lived **local security service** benefits from one binary, simpler distribution, lower footguns around native modules (`better-sqlite3`), and a clearer process boundary from agent runtimes.

**Why not Rust for v1?**  
Rust is fine long-term if Jeffrey is fluent. For solo founder velocity on a still-validating product, Go hits the performance/security/maintainability sweet spot without borrow-checker tax on every filesystem edge case.

**FortSignal SDK note:** Official SDK is TypeScript. FortMemory talks to FortSignal via **HTTP REST** (`/challenge/start`, `/challenge/verify`, etc.). No dependency on the TS SDK in-process. A thin Go client in `internal/fortsignal` is intentional and healthy.

## UI stack (aligned with hybrid strategy)

| Layer | Choice | Phase |
|-------|--------|-------|
| Daemon | Go binary | MVP |
| Dashboard | Static HTML/JS embedded (`embed.FS`) | MVP thin |
| Tray | Optional Go systray | Phase 2 |
| Desktop shell | Optional Tauri wrapping localhost | Phase 3 |
| Obsidian | Companion plugin | Add-on |

**Not** pure browser app. See [UI-HYBRID.md](./UI-HYBRID.md).

## Recommended libraries (MVP)

| Concern | Choice | Notes |
|---------|--------|-------|
| Module path | `github.com/fortsignal/fortmemory` | Adjust if org differs |
| CLI | `cobra` | Subcommands: init, serve, reindex, agent, version |
| HTTP | `go-chi/chi/v5` | Lightweight middleware |
| Config | TOML via `pelletier/go-toml/v2` | Vault-local `.fortmemory/config.toml` |
| FS watch | `fsnotify` | Debounce; ignore `.fortmemory/` |
| SQLite | `modernc.org/sqlite` (pure Go) | Easy cross-compile |
| FTS | SQLite FTS5 | Always-on keyword search |
| Vectors | blob table + cosine later | Optional post-MVP |
| Embeddings | HTTP → Ollama | Optional; FTS fallback |
| FortSignal | `internal/fortsignal` HTTP client | Mirrors SDK agent API |
| Ed25519 | `crypto/ed25519` | Local signer mode |
| Hash | `crypto/sha256` | contentHash |
| Logging | `log/slog` | Structured |
| Tests | std `testing` + `httptest` | Table-driven |

## Explicit non-choices (MVP)

| Avoid | Why |
|-------|-----|
| Tauri / Electron | Not required for daemon; UI is localhost web later |
| Kubernetes-style deps | Single-user local process |
| Graph DB (Neo4j, etc.) | Post-MVP if ever |
| gRPC | REST JSON is enough for agents and curl |
| Heavy DI frameworks | Keep packages small and explicit |
| Embedding models inside binary | Call Ollama / external; keep binary lean |

## Runtime topology

```
┌──────────────────────────────────────────────┐
│  fortmemory (single Go process)              │
│  ├── HTTP API :7432 (loopback default)       │
│  ├── vault watcher                           │
│  ├── SQLite index                            │
│  ├── FortSignal HTTPS client                 │
│  └── optional embed worker queue             │
└──────────────────────────────────────────────┘
         ▲                ▲
         │                │
    agents (any)     browser UI (optional, later)
```

**One process per vault instance** (or multi-vault in one process later).  
Agents never write the index directly.

## Embeddings strategy

1. **Default:** hybrid FTS5 + vectors when Ollama is reachable  
2. **Degraded:** FTS5 only (still shippable)  
3. **Never block writes** on embed completion — index async; mark `embed_pending`

## Cross-compile / distribution (target)

```bash
GOOS=linux  GOARCH=amd64 go build -o fortmemory ./cmd/fortmemory
GOOS=darwin GOARCH=arm64 go build -o fortmemory ./cmd/fortmemory
GOOS=windows GOARCH=amd64 go build -o fortmemory.exe ./cmd/fortmemory
```

Prefer pure-Go SQLite for this path.

## Cloudflare Tunnel vs Tailscale

| Tunnel | Role |
|--------|------|
| **Cloudflare Tunnel + mTLS** | **Primary** remote access (founder preference) |
| **Tailscale** | Supported alternative |

CLI helpers document both. Tunnel work is **post-MVP** unless needed for dogfood. See [FOUNDING-CONTEXT.md](./FOUNDING-CONTEXT.md).

## Stack principles for a solo founder

1. Stdlib first  
2. One way to do HTTP  
3. One database file  
4. No plugin system until a second real consumer needs it  
5. OpenAPI as contract, not as enterprise ceremony  
