# Architecture Decision Records

## ADR-001: Go for local memory server

**Status:** Accepted  
**Date:** 2026-07-16  

### Context

Need a local-first memory daemon: CLI, HTTP API, filesystem watch, SQLite index, crypto hashes, optional tunnels, FortSignal HTTPS client. Solo founder; security-first; validate before heavy build.

### Decision

Build FortMemory server and CLI in **Go**.

### Consequences

- Single binary distribution  
- Clean process boundary from agent runtimes (TS/Python)  
- FortSignal via REST, not `@fortsignal/sdk` in-process  
- UI (if any) is separate localhost web app later  

### Alternatives considered

- **TypeScript:** faster if only stack is TS; weaker “infra daemon” packaging  
- **Rust:** strong but slower solo iteration unless already fluent  

---

## ADR-002: Markdown vault as source of truth

**Status:** Accepted  

### Decision

Canonical memory store is an Obsidian-compatible folder of Markdown files. SQLite under `.fortmemory/` is a rebuildable index.

### Consequences

- Human edit always possible  
- Git-friendly  
- Index can be deleted and rebuilt  
- Must path-jail and handle concurrent human + agent edits  

---

## ADR-003: FortSignal mandatory on mutating ops

**Status:** Accepted  

### Decision

`memory.write` and `memory.delete` always require FortSignal `decision: allow` before durable change. Reads/search ungated by default; may be gated by sensitivity later.

### Consequences

- Product thesis stays sharp  
- Latency on writes acceptable; reads stay fast  
- Requires FortSignal API key + agent passport/delegation for autonomous agents  
- Offline mutate is fail-closed in MVP  

---

## ADR-004: Single writer per vault

**Status:** Accepted  

### Decision

One FortMemory process is the mutation authority for a vault. Agents are API clients.

### Consequences

- Avoids split-brain index  
- Multi-machine requires tunnel to one server or separate vaults  
- Delays multi-master CRDT complexity  

---

## ADR-005: Name FortMemory / FortVault; retire SignalVault

**Status:** Accepted  

### Decision

Product = FortMemory. Cloud tier = FortVault. Do not brand as SignalVault.

---

## ADR-006: No Tauri for v1

**Status:** Accepted  

### Decision

Ship CLI + HTTP first. Optional thin web UI against localhost later. No desktop shell required for MVP.

---

## ADR-007: Defer FortVault and P2P until validation

**Status:** Accepted  

### Decision

Phase 0 validation and Phase 1 local core before R2 sync, peer protocol, or discovery network.

### Consequences

- Lower risk of overbuilding  
- Tunnel helpers optional for dogfood only  

---

## ADR-008: Hybrid search, not graph memory, for MVP

**Status:** Accepted  

### Decision

SQLite FTS5 + optional local embeddings. Temporal knowledge graphs out of MVP.

---

## ADR-009: Cloudflare Tunnel primary (mTLS); Tailscale supported

**Status:** Accepted (updated 2026-07-16 — founder preference)  

### Decision

- **Primary remote path:** Cloudflare Tunnel **with mTLS**  
- **Supported alternative:** Tailscale  
- Neither required for pure localhost MVP  

Aligns with Cloudflare R2 (FortVault) and existing Cloudflare familiarity.

---

## ADR-010: Local path/tag policy may extend FortSignal

**Status:** Accepted  

### Decision

Use FortSignal for action/amount/recipient/delegation/biometric. FortMemory may enforce additional path globs and sensitivity labels when recipient allowlists are insufficient.

### Order

Crypto → FortSignal policy → local policy → execute.

---

## ADR-011: Hybrid local-first UI — not pure browser product

**Status:** Accepted  
**Date:** 2026-07-16  

### Context

Temptation to ship a browser-only app for velocity. Conflicts with vault FS access, agent localhost API, key custody, and local-first promise.

### Decision

- **Primary:** Go daemon + CLI  
- **Secondary:** Embedded localhost web dashboard  
- **Later:** tray / optional Tauri / Obsidian plugin  
- **Not primary:** pure browser SaaS memory app  

See [UI-HYBRID.md](./UI-HYBRID.md).

---

## ADR-012: FortSignal HTTP client in Go; mirror SDK agent API

**Status:** Accepted  

### Decision

Implement `internal/fortsignal` against production REST shapes from fortsignal-api/sdk (agent start/verify). Do not embed the TypeScript SDK.

### Constraints carried into design

- recipient ≤ 256, metadata ≤ 2048  
- agent fast-fail may return HTTP 403 + `decision: deny`  
- path encoding scheme for policy wildcards  
- dashboard-only delegation approval  

See [FORTSIGNAL-INTEGRATION.md](./FORTSIGNAL-INTEGRATION.md).

---

## ADR-013: Open-core (Apache-2.0 local core)

**Status:** Accepted  
**Date:** 2026-07-16  

### Context

Need adoption and auditability for a security-sensitive local memory server, while monetizing as a solo founder.

### Decision

- **Open source (Apache 2.0):** local FortMemory server, vault, search, FortSignal *client*, local policy, tunnel helpers  
- **Commercial:** FortVault (R2/team), enterprise packs, managed hosting, support  
- **FortSignal SaaS** remains the metered governance plane (not “free unlimited” via FortMemory)  
- No license key required to run local core; no default phone-home  

See [OPEN-CORE.md](./OPEN-CORE.md) and [COMMERCIAL.md](./COMMERCIAL.md).

---

## ADR-014: Strategy freeze — no more Cloudflare engineering this cycle

**Status:** Accepted  
**Date:** 2026-07-16  

### Decision

Cloudflare Tunnel remains **primary remote strategy**. Existing light helpers are enough.  
**Do not** expand tunnel/Zero Trust automation until MVP `serve` + validation demand it.

Canonical freeze: [STRATEGY-LOCK.md](./STRATEGY-LOCK.md).

---

## ADR-015: Memory-first product surface; Composer is not the lobby

**Status:** Accepted  
**Date:** 2026-07-16  

### Context

Risk of FortMemory UX collapsing into “open FortSignal Policy Composer,” which under-delivers memory value and feels like a bounce.

### Decision

- FortMemory UI IA: **Home / Search / Activity / Agents / Vault / Settings**  
- Daily value: search, recall, activity, agent ops  
- FortSignal Composer + dashboard: **Settings deep links** + optional setup  
- Default path policies so day-one works without Composer  
- Agent interface prioritizes stable API/MCP + deny reasons  

Canonical: [PRODUCT-SURFACE.md](./PRODUCT-SURFACE.md).
