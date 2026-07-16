# System Architecture — FortSignal × FortMemory × Obsidian

**Audience:** Founder + engineers  
**Status:** Canonical full-stack picture (not Composer-only)  
**Date:** 2026-07-16  

This document is the **whole system**: who owns what, how every major FortSignal surface plugs into FortMemory, how Obsidian fits, and how agents actually run.  
If you only read one architecture doc for “does this hang together,” read this.

---

## 1. The one-sentence system

> **Humans author notes in Obsidian and author rules in FortSignal. Agents act through FortMemory. FortSignal decides allow/deny. Disk holds truth. Receipts prove it.**

Nothing else is allowed to blur those roles.

---

## 2. Three products, three jobs

```
┌──────────────────────────────────────────────────────────────────────────┐
│                         HUMAN WORLD                                       │
│  Obsidian (edit notes)     FortSignal Dashboard (policy, passports,      │
│                            Composer, delegations, audit)                  │
└───────────────┬──────────────────────────────┬───────────────────────────┘
                │ files on disk                │ rules + identity (cloud
                │                              │ or self-host FS API)
                ▼                              ▼
┌───────────────────────────┐    ┌─────────────────────────────────────────┐
│  FORTMEMORY (local)       │───►│  FORTSIGNAL (enforcement plane)         │
│  Memory state plane       │    │  Intent bind · policy · passports       │
│  Vault · index · HTTP API │◄───│  challenge/start · verify · signalId    │
└─────────────┬─────────────┘    └─────────────────────────────────────────┘
              │
              │ localhost / tunnel
              ▼
┌───────────────────────────┐
│  AGENTS (customer lane)   │
│  LLM · tools · MCP · CLI  │
│  Hold Ed25519 private key │
└───────────────────────────┘
```

| Product | Owns | Does **not** own |
|---------|------|------------------|
| **Obsidian** | Human UX for notes, links, graph | Agent auth, policy, multi-agent API |
| **FortMemory** | Vault I/O, search index, local agent API, local receipts mirror | Cryptographic verification, NL policy compilation, global identity |
| **FortSignal** | Signatures, policy, delegation, `signalId`, Composer, audit API | Note content as source of truth, RAG, Obsidian plugins |

**FortVault (later)** = optional commercial sync of vault objects (R2). Still not the governance plane.

---

## 3. Map onto FortSignal’s own “three lanes”

From FortSignal agentic architecture:

| Lane | FortSignal name | In this system |
|------|-----------------|----------------|
| **1** | Customer agent (their model/logic) | Coding agent, research agent, MCP client, `fortmemory write` |
| **2** | Integration middleware | **FortMemory server** (and later optional DeepAgents-style hooks that call FortMemory, not raw FS) |
| **3** | Enforcement Worker | **FortSignal API** unchanged |

Critical alignment:

- FortSignal **does not run agents**.  
- FortMemory **does not invent allow/deny**.  
- Agents **must not** write the vault with raw `open()` in the supported path.

```
Lane 1: Agent decides “remember X”
Lane 2: FortMemory packs intent + signs (or requires signature) + writes file only on allow
Lane 3: FortSignal verifies signature + policy + returns signalId
```

DeepAgents today intercepts **tools** (shell, files in a repo).  
FortMemory is the parallel lane for **long-term knowledge vault** tools (`memory.write`, `memory.search`).  
Same company story: **every high-stakes effect goes through FortSignal.**

---

## 4. Full stack layers (end-to-end)

```
L5  EXPERIENCE
    Obsidian app · FortSignal dashboard/Composer · optional FM thin UI · optional Obsidian plugin

L4  ACCESS
    FortMemory loopback HTTP · agent bearer tokens
    Cloudflare Tunnel + Access/mTLS (primary remote) · Tailscale (alt)
    (Later) peer share channels

L3  GOVERNANCE  ← 100% FortSignal product surface
    WebAuthn humans · Ed25519 agents · passkey-approved delegations
    NL Policy Composer → PolicyProfile in KV
    Layer 1 param bind · Layer 2 assertPolicy · Signal Views · /audit · signalId

L2  MEMORY ENGINE  ← FortMemory
    Path jail · local path policy · write planner · FTS/embeddings · receipts log
    FortSignal HTTP client · agent key registry · reindex/watcher

L1  STORAGE  ← user machine (and later FortVault)
    Markdown vault (canonical) · .fortmemory/* derivative state
```

Governance is not a “library inside FortMemory.” It is a **network dependency** with a fail-closed contract on mutates.

---

## 5. FortSignal surface area → FortMemory usage

This is the **whole** FortSignal product map, not just Composer.

### 5.1 Identity & registration

| FortSignal capability | How FortMemory uses it |
|----------------------|-------------------------|
| Human passkeys (`/register/*`) | Owner step-up later (export, sensitive path, peer grant); MVP agent-first |
| `userId` model | Optional: vault owner as `userId` for human ceremonies |
| Agent passport (`/agent/register`, dashboard keygen) | Each memory agent has Ed25519 identity |
| Public key only on server | Private key on disk next to FortMemory or in agent process |
| Key rotate / revoke | Dashboard revoke → next `memory.write` denied |

### 5.2 Delegation (the human grant)

| Capability | Role in system |
|------------|----------------|
| Passkey-approved delegation | Human sets memory boundary once |
| Policy + expiry on agent | Binds which memory actions/paths/sizes |
| Instant revoke | Compromised agent stops writing **without** redeploying FortMemory |
| Dashboard-only issue/revoke | API key cannot mint god-mode agents (security feature) |

**Architecture rule:** FortMemory never creates delegations. It only **consumes** them at challenge time.

### 5.3 Policy engine (Layer 2)

| Capability | Memory mapping |
|------------|----------------|
| `allowedActions` | `memory.write`, `memory.read`, `memory.delete`, `memory.search`, later `memory.share_*` |
| `maxAmountPerAction` / per-action caps | **Byte length** of body (or 0 for non-byte ops) |
| `allowedRecipients` + `/*` wildcards | `vaultId/path` e.g. `personal/Scratch/*` |
| `allowedFromSources` | Optional: agentId, `opencode`, `mcp`, hostname |
| `requiredMetadata` | Prefer fixed constants only (`vaultId=personal`); **never** require dynamic `contentHash` |
| `requireBiometric` / thresholds | Human step-up flows; agents use delegation-time biometric |
| `expiresAt` | Delegation/policy window |
| Empty recipient list | No FS-side recipient constraint in FS — **FortMemory local deny still applies** for `Private/**` |

### 5.4 NL Policy Composer (part of governance UX, not FortMemory core)

| Composer piece | System role |
|----------------|-------------|
| Public `/composer` + `/policy/translate` | Humans describe memory rules in English |
| Passkey save to policy | Rules become real Layer 2 constraints |
| `nlText` stored on profile | Audit “what the human meant” |
| Dashboard edit | Same policies attach to agent delegations |

**Architecture rule:** Composer is a **FortSignal frontend to PolicyProfile**.  
FortMemory supplies the **domain vocabulary** (action names, path scheme, templates) and **deep links**. It does not run an LLM policy translator.

### 5.5 Challenge pipeline (the spine)

```
Every FortMemory mutate (and later gated read/share):

  pack(action, amount, recipient, source, metadata)
       → POST /challenge/start   { agentId, ... }
       → (403 deny possible here — agent fast-fail)
       → Ed25519 sign(challenge bytes)
       → POST /challenge/verify  { agentId, challenge, signature }
       → allow + signalId  OR  deny + reason
       → execute local effect only on allow
```

Same pipeline as transfers/tools. **Memory is another action class**, not a fork of the protocol.

### 5.6 Receipts, audit, Signal Views

| Capability | System role |
|------------|-------------|
| `signalId` on allow | Written into local receipt log + optional note frontmatter |
| `GET /signal/:id` | Optional reconciliation / incident |
| `GET /audit` | Tenant-wide history (cloud 90d); FortMemory keeps **durable local** mirror |
| Signal Views | Peer/team disclosure scopes later (`audit` / `proof` receipts to partners) |
| Deny reasons | Returned to agents so they can escalate (aligns with proactive policy later) |

### 5.7 DeepAgents & other FortSignal agent tooling

| Piece | Relationship |
|-------|----------------|
| `fortsignal-deepagents` | Gates **coding/tool** actions in-process |
| FortMemory | Gates **vault memory** actions as a service |
| Ideal agent stack | DeepAgents for tools **+** FortMemory client for long-term notes |
| Anti-pattern | DeepAgents writes vault files directly with FS tools |

Future: a DeepAgents/MCP tool adapter whose implementation is `HTTP POST fortmemory/v1/write` (governed), not `fs.writeFile`.

### 5.8 Enterprise FortSignal (hosted + mTLS, self-host)

| Mode | FortMemory impact |
|------|-------------------|
| SaaS `api.fortsignal.com` | Default `api_base` |
| Self-host FortSignal | Point `fortsignal.api_base` at internal URL; same client |
| mTLS to FortSignal | Enterprise: client certs on FortMemory process |
| Daily limits / cooldowns in FS | Cumulative counters still **also** local if needed (FS is per-action) |

### 5.9 What FortSignal deliberately does **not** do here

- Store Markdown bodies as system of record  
- Run RAG  
- Watch the filesystem  
- Replace Obsidian  
- Issue agent private keys to the cloud  

---

## 6. FortMemory internal architecture (state plane)

```
                    ┌──────────── CLI / HTTP / MCP ────────────┐
                    │  init serve write agent reindex cloudflare│
                    └──────────────────┬───────────────────────┘
                                       │
                    ┌──────────────────▼───────────────────────┐
                    │              memory.Service                │
                    │  resolveSigner · local policy · orchestrate│
                    └─┬──────────┬──────────┬──────────┬───────┘
                      │          │          │          │
              ┌───────▼──┐ ┌─────▼────┐ ┌───▼────┐ ┌───▼────────┐
              │  vault   │ │  index   │ │receipts│ │ fortsignal │
              │ path jail│ │ FTS5/RAG │ │ JSONL  │ │ HTTP client│
              └────┬─────┘ └────▲─────┘ └────────┘ └─────▲──────┘
                   │            │  reindex                 │
                   │            │                          │
              ┌────▼────────────┴─────┐                    │
              │  Markdown vault disk   │                    │
              │  + .fortmemory/        │                    │
              └───────────▲────────────┘                    │
                          │ human edits                      │
                    ┌─────┴─────┐                    ┌───────┴────────┐
                    │  Obsidian │                    │ FortSignal API │
                    └───────────┘                    └────────────────┘
```

### Dual policy (ordered)

```
1. Cryptography     FortSignal Layer 1 (signature ↔ params)
2. Remote policy    FortSignal Layer 2 (delegation + PolicyProfile)
3. Local policy     FortMemory globs (Private/**, .fortmemory/**)
4. Execute          vault write / delete
```

Local policy is **defense in depth**, not a second product.

### Data that lives where

| Data | Location |
|------|----------|
| Note content | Vault `.md` files |
| Search index | `.fortmemory/index.sqlite` (rebuildable) |
| Local API tokens | `.fortmemory/agents.json` (hashed) |
| Agent private keys | User-managed path (never FortSignal, never git) |
| FortSignal API key | Env `FORTSIGNAL_API_KEY` |
| Governance decisions | FortSignal cloud audit + local receipts.jsonl |
| Policies / delegations | FortSignal KV only |

---

## 7. Obsidian in the full architecture

### 7.1 What Obsidian is here

- **Human IDE for the vault**  
- Not the agent runtime  
- Not required for FortMemory to run  
- Benefits automatically from plain Markdown + folder layout  

### 7.2 Compatibility contract

| Rule | Why |
|------|-----|
| Canonical store = vault folder Obsidian opens | One truth |
| Engine state in `.fortmemory/` (hidden) | Clean library |
| Agent writes default to `Inbox/`, `Scratch/` | Social convention humans understand |
| `Private/` local-deny by default | Safe defaults without Composer |
| Frontmatter `last_signal_id` on governed writes | Visible proof inside the note app |
| Human edits without signalId are OK | Humans are root authority (matches FortSignal identity model) |

### 7.3 Integration tiers (system view)

| Tier | Component | Depends on Obsidian process? |
|------|-----------|------------------------------|
| 0 | FortMemory ↔ folder | **No** |
| 1 | Thin plugin (status, recent signals, open Composer) | Yes, optional |
| 2 | MCP → FortMemory | No |
| ✗ | Core requires Local REST API plugin | Rejected as required path |

### 7.4 Concurrency model

```
Obsidian write  → disk change → (watcher) reindex → searchable, verifiedBy: human_external
FortMemory write → FortSignal allow → disk change → reindex → signalId receipt
Conflict same path → last writer wins on disk; MVP no CRDT
```

Professional honesty: multi-writer CRDT is a different product (Obsidian Sync class). Don’t pretend otherwise.

---

## 8. End-to-end journeys (whole system)

### Journey A — First-time setup (human)

```
1. Install fortmemory · init vault path (same folder as Obsidian vault)
2. FortSignal signup · API key
3. Composer: memory policy NL → passkey save PolicyProfile
4. Dashboard: Agent Passport · download key · passkey delegate policy
5. fortmemory agent add <id> --key <file>  → local bearer token
6. fortmemory serve
7. Open Obsidian on same folder
```

### Journey B — Agent remembers something

```
Agent → POST /v1/write { path, content }
FortMemory → local policy → challenge/start → sign → verify
allow → write Markdown → upsert FTS → append receipt (signalId)
Human → opens note in Obsidian · sees content (+ optional frontmatter signal)
```

### Journey C — Agent overreaches

```
Agent → write Private/taxes.md
Local deny OR FortSignal recipient_not_allowed
No file change · deny reason to agent · receipt of deny
Human → still trusts Private/
```

### Journey D — Human revokes agent

```
Dashboard → revoke delegation
Next memory.write → delegation_invalid
FortMemory binary unchanged · vault intact
```

### Journey E — Human edits in Obsidian

```
Edit note in UI → save
Watcher reindexes
No FortSignal call (human root path)
Search finds new text
```

### Journey F — Tool agent + memory agent (portfolio)

```
DeepAgents gates: shell, apply_patch, deploy
FortMemory gates: long-term notes, prefs, runbooks
Both produce signalId under same FortSignal tenant
One audit story for regulators: “actions and memories”
```

### Journey G — Remote laptop (later)

```
FortMemory on home machine 127.0.0.1
cloudflared + Access/mTLS
Agent on VPS → HTTPS → tunnel → FortMemory → FortSignal
Still no public unauthenticated vault API
```

---

## 9. Action & data contracts (system-wide)

### Memory actions (FortSignal `action` strings)

| Action | amount | recipient | When |
|--------|--------|-----------|------|
| `memory.write` | bytes | `{vaultId}/{relPath}` | Create/append/overwrite |
| `memory.delete` | 0 | path | Delete file |
| `memory.read` | 0 or maxChars | path | Optional gated read |
| `memory.search` | topK | vaultId or prefix | Optional gated search |
| `memory.share_request` | 0 | peer vault | Phase 3 |
| `memory.share_grant` | TTL/bytes | peer+scope | Phase 3 |
| `memory.export` | 0 | destination | Later |
| `memory.summarize` | 0 | scope path | Later |

### Intent metadata (bound, small)

```json
{
  "vaultId": "personal",
  "contentHash": "sha256:…",
  "mode": "overwrite",
  "path": "Scratch/note.md"
}
```

Full body never sent to FortSignal (2048 metadata limit + privacy).

### Local vs remote auth

| Credential | Scope |
|------------|--------|
| `FORTSIGNAL_API_KEY` | Tenant → FortSignal API (FortMemory process only) |
| Agent Ed25519 private key | Signs challenges (agent or local-signer mode) |
| `fm_*` bearer | AuthN to FortMemory HTTP only |

Three different secrets. Confusing them is the #1 integration bug.

---

## 10. Trust & threat model (system)

| Trust zone | Trusted for |
|------------|-------------|
| User OS account | Can always read vault files (honest limit) |
| FortMemory process | Path jail, calling FortSignal, writing on allow |
| FortSignal | Signature + policy decision |
| Agent process | Untrusted for integrity — may be prompt-injected |
| Obsidian plugins | Untrusted — can edit files; treated as human_external |
| Cloudflare edge | Transport; Access/mTLS for remote |

**Product promise (precise):**  
We stop **unauthorized agents** from mutating memory under policy — not a malicious root user with disk access.

---

## 11. Alignment with FortSignal design principles

| FortSignal principle | System expression |
|---------------------|-------------------|
| Human = root authority | Passkey delegations; Obsidian free edit; revoke anytime |
| Agent = zero inherent authority | No write without delegation |
| Same verify path | memory.* uses challenge/verify |
| Deterministic policy | assertPolicy, not LLM at enforce time |
| NL is authoring UX only | Composer translates; enforce is structured |
| Stateless enforcement | FortMemory holds files; FS holds rules |
| signalId audit | Local + cloud |
| DeepAgents-style intercept | FortMemory is intercept for vault tools |

---

## 12. What “good architecture” looks like for users (the click)

Users should feel **one company system**, not three random tools:

1. **I already use FortSignal** for agent actions → memory is “the same gate for notes.”  
2. **I already use Obsidian** → agents write *my* vault, not a black-box DB.  
3. **Deny is visible** → Private stays private.  
4. **Receipt is visible** → `signalId` in activity + optional frontmatter.  
5. **Revoke is instant** → dashboard, not redeploy.

If any layer forces a second identity system or a second policy language, architecture failed.

---

## 13. Non-goals (architecture-level)

| Non-goal | Reason |
|----------|--------|
| FortMemory reimplements Composer | Split brain, security risk |
| FortSignal stores vaults | Wrong plane; privacy regression |
| Obsidian required for agents | Server/headless agents must work |
| Multi-master CRDT MVP | Different company problem |
| Agent raw FS as supported path | Bypasses product |
| Phone-home OSS core | Breaks local-first trust |

---

## 14. Build sequence that respects the architecture

| Order | Why this order |
|-------|----------------|
| 1. Vault + path jail + FS client write | Spine of lanes 2–3 |
| 2. Serve + tokens + read/search | Agent-usable without Obsidian |
| 3. Watcher + frontmatter signalId | Obsidian coexistence |
| 4. Composer templates + deep links | Governance UX click |
| 5. Thin Obsidian plugin | Ambient trust, not core |
| 6. MCP → FortMemory | Agent ecosystem |
| 7. Tunnel hardening | Remote only when needed |
| 8. FortVault | Commercial sync last |

Skipping to plugin or cloud before (1–3) is how you get a demo that doesn’t survive real vaults.

---

## 15. One diagram to remember

```
                 ┌──────────────┐
                 │   Human      │
                 └──────┬───────┘
           notes│       │rules + biometric grant
                ▼       ▼
         ┌──────────┐  ┌─────────────────┐
         │ Obsidian │  │ FortSignal      │
         │  (UX)    │  │ Composer        │
         └────┬─────┘  │ Passports       │
              │        │ Delegations     │
              │ disk   │ Policy · Verify │
              ▼        └────────▲────────┘
         ┌──────────────────────┴──────────┐
         │         FortMemory              │
         │  API · jail · index · receipts  │
         └────────────────▲────────────────┘
                          │ memory.* tools only
                          │
                   ┌──────┴──────┐
                   │   Agents    │
                   └─────────────┘
```

---

## 16. Related docs

| Doc | Slice |
|-----|--------|
| [ARCHITECTURE.md](./ARCHITECTURE.md) | FortMemory-internal layers |
| [FORTSIGNAL-INTEGRATION.md](./FORTSIGNAL-INTEGRATION.md) | Wire-level API contract |
| [SECURITY.md](./SECURITY.md) | Threat model detail |
| [MVP.md](./MVP.md) | Scope cut |
| [STRATEGY-LOCK.md](./STRATEGY-LOCK.md) | Business/product freeze |
| FortSignal `12-agentic-layer.md` | Lanes + DeepAgents |
| FortSignal `14-policy-org-layer.md` | Layer 1/2, human/agent identity |
| FortSignal `17-phase-a-nl-policy.md` | Composer implementation |

---

## 17. Bottom line

This is **one architecture** with three planes:

1. **Experience** — Obsidian + FortSignal dashboard  
2. **State** — FortMemory + Markdown vault  
3. **Governance** — FortSignal (identity, policy, bind, verify, audit)  

Composer is **one UX entry** into plane 3 — not the architecture.  
Obsidian is **one UX entry** into plane 2 — not the architecture.  
Agents are **clients** of plane 2 that are **judged** by plane 3.

That’s the whole system. Everything we build should preserve those planes.
