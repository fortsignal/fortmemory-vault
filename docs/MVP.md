# MVP Scope (Locked)

**Authority:** [STRATEGY-LOCK.md](./STRATEGY-LOCK.md)  
**Goal:** Lowest-risk ship that proves “governed local memory writes,” not a full platform.

## Definition of done

A solo developer can:

1. Install/run a single Go binary  
2. `fortmemory init` on an existing Markdown/Obsidian vault  
3. **Write** a note only when FortSignal returns `decision: allow` + `signalId`  
4. **Serve** a loopback HTTP API with agent bearer tokens  
5. **Read** and **search** (FTS) notes via that API  
6. Inspect **local receipts** for governed writes  
7. Open the vault in Obsidian with the daemon stopped — files still there  

Ugly UI is fine. No cloud required.

## In scope

| Area | MVP include |
|------|-------------|
| CLI | `init`, `write`, `serve`, `version` (`reindex` if cheap) |
| Config | `.fortmemory/config.toml` |
| Vault | Path jail, atomic write, read, no escape to `.fortmemory/` |
| FortSignal | Agent start/verify on **write** (delete if trivial) |
| HTTP | `127.0.0.1` only: health, read, write, search, receipts |
| Auth | Local bearer token → agentId |
| Index | SQLite FTS5 (embeddings optional, not required) |
| Receipts | Local JSONL or SQLite append |
| Docs | Curl + FortSignal agent/delegation setup |
| Remote | **Document** Cloudflare Tunnel primary — do not build more tunnel code for MVP |

## API surface (MVP)

```
GET  /v1/health
POST /v1/search
GET  /v1/read
POST /v1/write
POST /v1/delete      # optional if timeboxed
GET  /v1/receipts
```

## Out of scope (explicit)

- FortVault / R2 sync  
- Peer-to-peer sharing  
- Vault profile discovery network  
- Graph memory  
- Auto memory extraction from chats  
- Tauri / heavy desktop  
- Deeper Cloudflare Access automation  
- NL Policy Composer inside FortMemory (link to FortSignal)  
- Multi-master multi-device writers  

## Non-goals for MVP metrics

- Beating Mem0 on recall benchmarks  
- Pretty dashboard  
- Public multi-tenant SaaS memory  

## Success criteria

| Signal | Pass |
|--------|------|
| You dogfood write+serve | ≥14 days |
| External builders | ≥3 try it |
| Strategy interviews | See VALIDATION.md |

## Implementation honesty

Already done in repo (do not redo as “skeleton”):

- config + init  
- vault path jail  
- FortSignal client + CLI `write`  

**True remaining MVP code:**

1. `serve` + agent bearer tokens  
2. HTTP read/write/search/receipts wiring  
3. FTS index + reindex  
4. Minimal curl docs  

Cloudflare: strategy only for this cycle.

## After MVP

Phase 2: MCP tools, policy templates, thin UI  
Phase 3: remote dogfood using existing tunnel helpers if needed  
Phase 4: FortVault commercial  
