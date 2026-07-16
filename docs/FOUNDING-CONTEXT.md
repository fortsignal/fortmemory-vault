# Founding Context (Canonical)

**Owner:** Jeffrey Walters, founder of FortSignal ([fortsignal.com](https://fortsignal.com))  
**Product:** FortMemory (working names also heard: SignalVault / FortVault)  
**Updated:** 2026-07-16  

This file is the short, non-negotiable product context. Prefer it when docs disagree on preferences.

## What FortMemory is

A **verifiable, local-first agentic crypto memory layer**:

- Real **Obsidian-compatible Markdown vaults** on the user machine  
- **FortSignal** governance on memory operations (signatures, NL policies, agent passports, parameter binding, `signalId` receipts)  
- Optional hybrid **FortVault** cloud sync (Cloudflare R2)  
- Secure remote access via **tunnels**  
- Built for agents, local models, and eventual multi-node / peer sharing  

## Key principles

1. **Local-first by default** — data stays on disk unless the user opts into FortVault.  
2. **Strong FortSignal integration** — verifiable intent applied to memory ops (reads, writes, shares as product intent; see pragmatism note below).  
3. **Hybrid optional** — FortVault (R2) for teams / multi-device when needed.  
4. **Tunnels:** **Cloudflare Tunnel primary (with mTLS)**; Tailscale supported.  
5. **Language:** **Go** for the local memory server.  
6. **UI:** Minimalist — local server + browser dashboard + optional desktop tray. **Not** a heavy pure-browser product.  
7. **Validate demand before heavy building** — low-risk, solo-founder sequencing.  

## Main features (roadmap scope, not all MVP)

| Feature | Intent |
|---------|--------|
| Local Memory Server | Watch Obsidian vault; API for agents |
| Semantic search / RAG | Local models (e.g. Ollama) |
| Peer memory sharing | Scoped requests + policy + receipts |
| Vault profiles | Opt-in discovery (description, tags, access level) |
| NL policy composer | Memory ops + sharing (via FortSignal Composer, not a fork) |
| Multi-node agents | Local models connecting across nodes under governance |
| FortSignal on ops | Writes/shares always; reads under product policy |

## Current preferences (founder)

| Topic | Preference |
|-------|------------|
| Tunnel vendor | **Cloudflare Tunnel primary + mTLS**; Tailscale supported |
| Cloud sync | FortVault on **Cloudflare R2** (later / hybrid) |
| Server language | **Go** |
| UI | Local server + thin browser dashboard + optional tray |
| Risk posture | Pragmatic, validate before heavy build |
| Focus | Verifiable intent + local-first + Obsidian compatibility |

## Naming

| Name | Role |
|------|------|
| **FortMemory** | Primary product name (local server + product) |
| **FortVault** | Optional cloud sync tier (R2) |
| **SignalVault** | Working/alt name only — do not lead marketing with it |

## Pragmatism note (solo founder)

**Product intent:** FortSignal governance applies to memory operations broadly (including sensitive reads and all shares).  

**MVP engineering default (until demand proves otherwise):**

- Mutates (`write` / `delete`) and shares: **always** FortSignal-gated  
- Routine search/read: may be **locally authenticated** first for latency, with path/sensitivity policy; escalate to FortSignal for confidential/restricted or when config requires  

This is an implementation sequencing choice, not a retreat from verifiable intent. Flip `allow_ungated_reads` off when dogfood or customers require full gating.

## Business model

**Open-core (Apache-2.0 local core).**  
See [OPEN-CORE.md](./OPEN-CORE.md) and [COMMERCIAL.md](./COMMERCIAL.md).

- OSS: local server, vault, search, FortSignal client, tunnel helpers  
- Paid: FortVault R2/team, enterprise, support, managed hosting  
- FortSignal remains separately monetized governance  

## Out of band

Do not assume features, pricing, or partners beyond this file and the rest of `docs/` without founder confirmation.
