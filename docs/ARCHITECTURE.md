# Architecture — FortMemory

## Overview

FortMemory is a **local memory authority**: it owns vault files and the search index, and it refuses mutating work unless FortSignal returns `decision: allow` with a `signalId`.

```
┌─────────────────────────────────────────────────────────────────────────┐
│ L4  ACCESS                                                              │
│  Loopback HTTP · CF Tunnel + mTLS (primary) · Tailscale · peers later   │
└───────────────────────────────┬─────────────────────────────────────────┘
                                │
┌───────────────────────────────▼─────────────────────────────────────────┐
│ L3  FORTSIGNAL GOVERNANCE                                               │
│  challenge/start · sign · challenge/verify · policy · receipts          │
└───────────────────────────────┬─────────────────────────────────────────┘
                                │ allow only
┌───────────────────────────────▼─────────────────────────────────────────┐
│ L2  MEMORY ENGINE                                                       │
│  search · write planner · chunk/embed · summarize (later) · peer router │
└───────────────────────────────┬─────────────────────────────────────────┘
                                │
┌───────────────────────────────▼─────────────────────────────────────────┐
│ L1  STORAGE                                                             │
│  Markdown vault (canonical) · .fortmemory/index.sqlite · receipts.sqlite│
└─────────────────────────────────────────────────────────────────────────┘
```

## L1 — Storage

### Canonical vault layout (example)

```
~/Vaults/Personal/                 # user-owned Obsidian vault
  README.md
  Agents/
  Inbox/
  Scratch/
  Private/
  .fortmemory/
    config.toml
    index.sqlite                   # FTS + vectors + file mtimes
    receipts.sqlite                # local signalId mirror
    agent-tokens/                  # local API credentials (not FortSignal keys)
    cache/
```

### Principles

- **Markdown files are source of truth.** Delete the daemon; vault still opens in Obsidian.  
- **Index is rebuildable** via `fortmemory reindex`.  
- **Receipts are append-only locally** (mirror of FortSignal allows + local denials).  
- **No proprietary binary document format** for notes.

### Suggested frontmatter (optional, progressive)

```yaml
---
id: mem_7f3a
type: fact | preference | procedure | episode
tags: [payments, vendors]
sensitivity: public | internal | confidential | restricted
updated: 2026-07-16T12:00:00Z
content_hash: sha256:…
last_signal_id: 5564c849-294c-4703-8f9f-0b10403de1d4
---
```

Human edits without frontmatter still work; engine fills metadata on governed writes.

## L2 — Memory engine

### Responsibilities

| Component | Job |
|-----------|-----|
| Vault watcher | Detect external human edits; reindex |
| Write planner | Resolve path, mode (create/append/overwrite), conflict rules |
| Chunker | Split Markdown for embedding |
| Indexer | FTS5 + optional vectors |
| Search | Hybrid BM25/FTS + vector fusion |
| Receipt store | Persist decision + params + path |
| (Later) Compactor | Summarization as governed action |
| (Later) Peer router | Share request protocol |

### Mutate pipeline (core path)

```
request
  → authenticate local agent/API token
  → compute contentHash = SHA-256(body)
  → FortSignal challenge.start(action, recipient=path, metadata{vaultId,contentHash,mode})
  → if deny: return reason (agent fast-fail)
  → agent/human signs challenge
  → FortSignal challenge.verify
  → if deny: return reason
  → write file under vault root (path jail)
  → update index (or enqueue embed)
  → append local receipt with signalId
  → return { decision: allow, signalId, path }
```

### Read / search pipeline (MVP default)

```
request
  → authenticate
  → optional local policy check (path/tag)
  → query index
  → return citations { path, score, excerpt, last_signal_id? }
```

**Product intent:** FortSignal governance covers memory ops including sensitive reads and all shares ([FOUNDING-CONTEXT.md](./FOUNDING-CONTEXT.md)).  

**MVP sequencing default:** routine read/search may be local-auth only for latency; mutates and shares always FortSignal-gated. Sensitivity labels and config can force FortSignal on reads.

### Path jail

All paths resolve under vault root. Reject `..`, absolute escapes, and writes into `.fortmemory/` except by the engine itself.

## L3 — FortSignal governance

FortSignal remains **stateless enforcement**. FortMemory owns state and only mutates on allow.

### Action taxonomy

| Action | When |
|--------|------|
| `memory.write` | create / append / overwrite |
| `memory.delete` | delete file |
| `memory.read` | optional gated read |
| `memory.search` | optional gated search |
| `memory.export` | later |
| `memory.share_request` | later |
| `memory.share_grant` | later |
| `memory.summarize` | later |

### Parameter binding mapping

| FortSignal field | Memory meaning |
|------------------|----------------|
| `action` | e.g. `memory.write` |
| `amount` | byte length (or 0) |
| `recipient` | `vault:{vaultId}/{relative/path.md}` |
| `source` | agentId or `ui` |
| `metadata.vaultId` | vault id |
| `metadata.contentHash` | `sha256:hex` of exact body |
| `metadata.mode` | `create` \| `append` \| `overwrite` |
| `metadata.bytes` | size |

Tampering body after challenge start → hash mismatch → deny `parameters_tampered`.

### Two policy layers

1. **FortSignal Layer 2** — allowed actions, amount/byte caps, recipient allowlists, expiry, biometric flags, delegation  
2. **FortMemory local policy** — path globs, tags, sensitivity (if FortSignal recipient matching is too coarse)

Order: **signature → FortSignal policy → local path/tag policy → execute**.

## L4 — Access

| Mode | MVP | Notes |
|------|-----|-------|
| `127.0.0.1` only | Yes | Default |
| Token auth per agent | Yes | Local bearer |
| mTLS | Later | Enterprise / peer |
| Cloudflare Tunnel + mTLS | Docs + helper later | **Primary remote** |
| Tailscale | Docs + helper later | Supported remote |
| FortVault control plane | Later | Cloud pairing |

## Multi-agent / multi-model scaling

**Pattern: one memory authority, many clients.**

```
        FortMemory (single writer)
              ▲
   ┌──────────┼──────────┐
agent-A   agent-B    UI/dashboard
(Ollama)  (coder)    (browser)
```

Rules:

1. Agents do not co-write SQLite  
2. File locks / single process serialize vault mutations  
3. Embed jobs async in-process queue  
4. Multi-machine: one shared FortMemory over tunnel **or** per-device vaults + peer share — not two writers on one bare folder  

Prefer language: **multi-agent clients on one vault**, not “swarms.”

## FortVault (future hybrid layer)

Not in MVP. When added:

- Local remains primary  
- Encrypted objects to R2  
- Manifest with content hashes + last signalIds  
- No multi-master CRDT in first cloud ship — prefer single active writer or explicit merge folder  

## Failure modes

| Failure | Behavior |
|---------|----------|
| FortSignal unreachable | Fail closed on mutates (MVP); optional offline queue is post-MVP |
| Ollama down | Search degrades to FTS |
| External human edit | Watcher reindexes; no signalId on ungoverned human edit (honest labeling) |
| Partial write | Temp file + rename; receipt only after durable write |

## Trust boundaries

```
[Untrusted] Agent process, peer node, browser UI
[Trusted computing base] fortmemory binary + vault FS permissions + FortSignal verify path
[Out of TCB] LLM outputs, embedding quality, Obsidian plugins
```

LLMs propose; FortMemory + FortSignal dispose.
