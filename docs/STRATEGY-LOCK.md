# Strategy Lock — FortMemory

**Status:** LOCKED for current planning cycle  
**Owner:** Jeffrey Walters (FortSignal)  
**Date:** 2026-07-16  

This document is the high-leverage freeze. Do not expand Cloudflare, peer mesh, or FortVault implementation until Phase 0 validation and MVP code path are intentional.

---

## 1. Product one-liner

**FortMemory** = local-first, Obsidian-compatible agent memory with **FortSignal-verifiable** writes (and later shares).  
**FortVault** = optional paid cloud sync (R2).  
**FortSignal** = enforcement plane (existing product).

> Memory you can prove.

---

## 2. Principles (non-negotiable)

| # | Principle |
|---|-----------|
| 1 | **Local-first** — Markdown vault on disk is source of truth |
| 2 | **FortSignal on mutates** — signatures, policies, passports, `signalId` |
| 3 | **Hybrid optional** — FortVault R2 later, not MVP |
| 4 | **Remote access** — Cloudflare Tunnel **primary** (mTLS/Access); Tailscale supported |
| 5 | **Go** local server |
| 6 | **UI** — daemon + thin localhost dashboard + optional tray later; not pure browser app |
| 7 | **Validate demand** before heavy multi-node / cloud build |

---

## 3. Open-core boundaries (locked)

**License (core):** Apache-2.0  

### Open source (this product direction)

- Local Memory Server (Go CLI + API)  
- Vault path jail, watcher, basic search  
- FortSignal **client** integration  
- Local policy (path globs)  
- Local receipts  
- Tunnel **helpers** (Cloudflare + Tailscale) — docs/light CLI only; **no further tunnel engineering now**  
- Minimal localhost dashboard  

### Proprietary / paid

- FortVault cloud sync (R2, multi-device, team)  
- Enterprise (SSO, compliance export polish, managed hosting)  
- Premium support  
- FortSignal SaaS itself (Composer, passports UI, metering) — separate product  

### Monetization twist

OSS local core → adoption → **FortSignal verification volume** + later **FortVault Team** seats/storage.  
Do **not** license-key the local binary or phone-home by default.

Full detail: [OPEN-CORE.md](./OPEN-CORE.md) · [COMMERCIAL.md](./COMMERCIAL.md)

---

## 4. Remote access (local-first; tunnels optional)

| Decision | Value |
|----------|--------|
| Default | **Localhost only** — no domain/tunnel required |
| Primary for hostname/Access | Cloudflare Tunnel + Access/mTLS |
| Supported mesh | **Tailscale** (`fortmemory tailscale`) |
| OSS | Light helpers/docs for both |
| Heavy tunnel product work | Not prioritized beyond thin CLI |

Canonical: [REMOTE-ACCESS.md](./REMOTE-ACCESS.md) · [TAILSCALE.md](./TAILSCALE.md) · [CLOUDFLARE-TUNNEL.md](./CLOUDFLARE-TUNNEL.md)

---

## 5. Exact MVP scope (locked)

### MVP definition of done

A builder can:

1. Run one Go binary (`fortmemory`)  
2. `init` against a real Markdown/Obsidian folder  
3. **Write** a note only when FortSignal returns `allow` + `signalId`  
4. **Read/search** notes via local HTTP (loopback) with agent token  
5. See local **receipts** for governed writes  
6. Keep data as plain files if the daemon dies  

### In MVP

| Area | Include |
|------|---------|
| CLI | `init`, `write`, `serve`, `version` (+ `reindex` if cheap) |
| Vault | Path jail, atomic write, read |
| FortSignal | Agent challenge start/verify on **write** (and **delete** if trivial) |
| HTTP | `127.0.0.1` only: health, read, write, search (FTS), receipts |
| Auth | Local bearer tokens mapped to agentId |
| Index | SQLite FTS5 minimum (embeddings optional) |
| Config | `.fortmemory/config.toml` |
| Docs | Curl examples + FortSignal setup |

### Out of MVP (explicit)

- FortVault / R2  
- Peer-to-peer share protocol  
- Vault profile discovery network  
- Graph memory  
- Mem0-style auto-extraction  
- Tauri / heavy desktop  
- Deeper Cloudflare Zero Trust automation  
- NL Composer fork (deep-link FortSignal only)  
- mTLS mesh between agents  

### MVP success metric

> 3 external builders dogfood `write` + `serve` for a week without you operating their machines.

---

## 5b. Product surface / UI (locked direction)

**Canonical:** [PRODUCT-SURFACE.md](./PRODUCT-SURFACE.md) · [UI.md](./UI.md)

| Rule | Detail |
|------|--------|
| FortMemory is the **memory product** | Search, vault, agents, activity |
| FortSignal is **governance infrastructure** | Composer/delegations deep-linked from Settings |
| Home ≠ Composer | Avoid bounce-to-policy as primary UX |
| Value + proof | Demo: recall → write → deny Private → signalId |
| Agent UI | API/MCP structure first; operator dashboard second |
| Aesthetic | Minimal ops console, not heavy graphics |

---

## 6. Public landing outline (short)

**URL candidates:** fortsignal.com/memory · fortmemory.dev  

| Block | Content |
|-------|---------|
| **Hero** | “Memory you can prove.” · Local Markdown + FortSignal receipts · CTA: Join waitlist |
| **Problem** | Agents remember without authorization · Opaque DBs · Cloud-first privacy friction |
| **How it works** | 3 steps: vault on disk → agent API → FortSignal allow/deny + signalId |
| **Open-core** | Free local forever · FortSignal for governance · FortVault later for teams |
| **Who** | Local LLM / agent builders · FortSignal users · compliance-curious teams |
| **Not** | Not Mem0 clone · Not pure browser app · Not multi-master cloud day one |
| **CTA** | Waitlist email + “I run local agents” checkbox · Optional 15-min call |
| **Footer** | Apache-2.0 core · FortSignal · Contact |

Full copy bank: [LANDING.md](./LANDING.md)

---

## 7. Phased roadmap (after strategy)

| Phase | Focus | Stop condition |
|-------|--------|----------------|
| **0 Validation** | Landing + 10 interviews | Kill/proceed gate |
| **1 MVP code** | serve + tokens + search (write path largely exists) | Dogfood 14 days |
| **2 Agents** | MCP tools, policy templates, thin UI | 3 external users |
| **3 Remote** | Cloudflare/Tailscale **guides** only unless demand | Real remote dogfood |
| **4 FortVault** | R2 team sync | Paying design partner |

---

## 8. Code status vs “start skeleton” advice

Honest inventory (do not re-scaffold blindly):

| Item | Status |
|------|--------|
| Go module + CLI | **Exists** |
| config + `init` | **Done** |
| vault path jail | **Done** |
| FortSignal client + gated `write` | **Done** |
| HTTP `serve` + agent tokens | **Not done** — true next code |
| FTS search + watcher | **Not done** |
| Cloudflare plugin depth | **Stop** — enough for strategy |

When leaving strategy mode, next engineering order:

1. `serve` + bearer tokens  
2. `read` + FTS `search`  
3. watcher / reindex  
4. Thin dashboard  
5. Tunnel polish only if users need remote  

---

## 9. Decision log (this freeze)

- [x] Cloudflare Tunnel = primary remote (strategy)  
- [x] Open-core Apache-2.0 local; FortVault commercial  
- [x] MVP = local vault + FortSignal writes + loopback API  
- [x] No more Cloudflare engineering this cycle  
- [x] Landing outline ready for waitlist page  
- [x] Validate before FortVault / P2P  

---

## 10. Reply template (for other agents / future you)

> Strategy locked. Cloudflare Tunnel is primary remote in open-core; no more tunnel code now.  
> Open-core boundaries and exact MVP are in `docs/STRATEGY-LOCK.md`.  
> Landing outline is ready.  
> **Next code (when unfrozen):** HTTP `serve` + agent bearer tokens + read/search — not another skeleton, not deeper Cloudflare.
