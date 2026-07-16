# Product Vision — FortMemory

## One sentence

Persistent, human-readable agent memory that stays on the user’s machine by default — and nothing is written, read under policy, or shared without FortSignal cryptographic proof of intent.

## Problem

Session tokens prove login. Vector databases store embeddings. Neither proves:

- **what** an agent was allowed to remember or retrieve  
- **who/what** authorized a write  
- **whether** parameters were tampered after approval  

As agents gain long-term memory, memory becomes a high-stakes surface: secret exfil, poisoned notes, unauthorized overwrites, silent peer shares.

## Solution

**FortMemory** = Obsidian-compatible Markdown vault + local Go memory server + **FortSignal enforcement** on memory operations.

| Layer | Owner |
|-------|--------|
| State (files, index) | FortMemory |
| Decision (allow/deny + receipt) | FortSignal |
| Human edit path | Any Markdown editor (Obsidian, VS Code, etc.) |

## Naming

| Name | Use |
|------|-----|
| **FortMemory** | Product name (daemon, CLI, API, local UX) |
| **FortVault** | Optional cloud sync / team tier only |
| **SignalVault** | **Deprecated** — confuses FortSignal receipts with storage |

Hierarchy:

```
FortSignal (company + governance API)
 ├── FortSignal SDK / challenge-verify / agent passports
 ├── FortMemory (this product — local memory authority)
 └── FortVault (optional hybrid sync — later)
```

Tagline options:

1. **Memory you can prove.** (primary)  
2. Obsidian for agents — with receipts.  
3. Write once. Verify every time.  

## Positioning

**Not** “another Mem0.”  
**Is** “FortSignal applied to long-term memory.”

| Audience | Pitch |
|----------|--------|
| Indie / local LLM users | Private agent memory as real files; you can open it in Obsidian |
| Agent builders | API + MCP later; every write has a `signalId` |
| Enterprise / compliance | Memory mutations are authorized actions with audit receipts |
| Existing FortSignal users | Same passports, policies, Composer — new action class |

**Companion rule:** FortMemory upsells FortSignal. It never replaces or reimplements the governance plane.

## Core principles

1. **Local-first default** — data stays on disk unless user opts into FortVault  
2. **Files are truth** — SQLite index is derivative; Markdown is canonical  
3. **Single writer per vault** — many agent clients, one memory authority  
4. **Crypto then policy** — FortSignal Layer 1, then Layer 2, then local path/tag policy  
5. **Human always readable** — no proprietary binary vault format for MVP  
6. **Validate before cloud/P2P** — demand gate before multi-node complexity  
7. **Security-first, UI-minimal** — CLI and API before polished graphics  

## Relationship to FortSignal primitives

| FortSignal primitive | FortMemory use |
|----------------------|----------------|
| `/challenge/start` + `/challenge/verify` | Gate every mutate (and optional sensitive read) |
| Parameter binding | Bind vaultId, path, contentHash, mode, agentId |
| Passkeys (human) | Step-up for sensitive paths / exports / peer grants |
| Ed25519 agent passports | Autonomous memory agents |
| Passkey-approved delegation | Time-bound agent memory scope |
| NL Policy Composer | Memory policies in plain English |
| `signalId` + `/audit` | Immutable memory operation receipts |
| Signal Views | Scope peer/team receipt disclosure |
| mTLS / self-host | Enterprise access patterns |

## Non-goals (product level)

- Beating Mem0 on LOCOMO-style recall benchmarks as the lead metric  
- Multi-master CRDT sync in v1  
- Global public memory marketplace  
- Replacing FortSignal dashboard / Composer  
- Merging into a coding-agent monolith  

## Business sketch

| Tier | Includes |
|------|----------|
| Community (free) | Local server, 1+ vaults, OSS core |
| Pro | Multi-vault polish, support, advanced local features |
| FortVault Team | Encrypted R2 sync, device pairing, shared vaults |
| Enterprise | Self-host, mTLS, retention, joint FortSignal deal |

Every memory write can consume a FortSignal verification → natural metering / upsell to agent-capable FortSignal plans.

## Success definition (product)

A builder runs `fortmemory serve` against an existing vault, points an agent at the API, and can answer:

> Show me the `signalId` for every note this agent wrote last week.

If they cannot, the product has failed its thesis — regardless of retrieval quality.
