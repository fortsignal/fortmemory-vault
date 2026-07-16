# Product Surface & UI Structure (Locked Context)

**Status:** Planning authority for UX / IA  
**Date:** 2026-07-16  
**Why this exists:** So later work does not collapse FortMemory into “bounce to Policy Composer” or security-only chrome.

Related: [SYSTEM-ARCHITECTURE.md](./SYSTEM-ARCHITECTURE.md) · [UI.md](./UI.md) · [UI-HYBRID.md](./UI-HYBRID.md) · [MVP.md](./MVP.md)

---

## Product identity (remember this)

| Layer | What users experience | Owner |
|-------|----------------------|--------|
| **Memory product** | Vault, search, write/recall, agents list, activity | **FortMemory** |
| **Governance** | Passports, delegations, NL policies, signalId verify | **FortSignal** |

**Internal label:** governance layer for agent memory operations.  
**User-facing label:** local agent memory (human-readable vault) with proof of writes.

**One-liner:**  
> Persistent agent memory as real notes — with proof of every write.

FortSignal is load-bearing infrastructure.  
FortMemory is the **application** agents and operators use daily.

---

## Principle: value first, governance embedded

1. **Home is memory ops** — search, recent activity, vault health — not Composer.  
2. **Composer is settings / deep link** — occasional rule authoring, not the lobby.  
3. **Security shows up as structure** — allow/deny rows, signalId, path limits — not a lecture.  
4. **Default policies on init** — day-one useful without visiting Composer.  
5. **Agent “UI” is often API** — clear tools + deny reasons; operator UI is for humans.

---

## Two surfaces called “agent UI”

| Surface | Audience | Priority |
|---------|----------|----------|
| **Operator dashboard** | Human (founder, team admin) | High for “click” and trust |
| **Agent interface** | Models / MCP / scripts | High for adoption (API shape > pixels) |

Most leverage early: **operator IA + agent API clarity**, not chat chrome or heavy graphics.

---

## Information architecture (operator)

### Shell

```
┌─ sidebar ──────────┬─ main ─────────────────────────────┐
│ FortMemory         │  [ global search .............. ]  │
│ ● vault status     │────────────────────────────────────│
│                    │  page body                         │
│ Home               │                                    │
│ Search             │                                    │
│ Activity           │                                    │
│ Agents             │                                    │
│ Vault              │                                    │
│ Settings           │                                    │
│                    │                                    │
│ FortSignal ●       │                                    │
└────────────────────┴────────────────────────────────────┘
```

Hybrid delivery: daemon serves UI at `http://127.0.0.1:7432/` (see UI-HYBRID). Minimal ops aesthetic — dense, calm, tables, monospace IDs. Not a lifestyle consumer app.

### Home (default route)

**Job:** Answer “is memory working and what happened?”

- Vault name, path, daemon ● running, note count / index lag  
- **Recent activity** (last N): decision · agent · action · path · signalId/copy · deny reason  
- Primary CTA: focus search or “View agents”  
- Empty state: “Point at a vault · add an agent · write Scratch/hello.md”

**Not on home:** full policy editor, Composer iframe as hero.

### Search

**Job:** Non-security value — recall.

- Query box (always reachable)  
- Results: path, excerpt, score, last_signal_id if any  
- Open path / copy path / preview Markdown  

### Activity (receipts)

**Job:** Trust + debugging (“what did it write / why blocked?”).

| Column | Purpose |
|--------|---------|
| Time | When |
| Decision | allow / deny badge |
| Agent | Who |
| Action | memory.write etc. |
| Path | Where |
| signalId | Proof (copy) |
| Reason | Deny clarity |

This feed is a **product differentiator** vs pure vector memory tools.

### Agents

**Job:** Multi-agent ops without chaos.

| Column | Purpose |
|--------|---------|
| agentId | Identity |
| Status | active / needs key / needs FortSignal delegation |
| Write scope summary | e.g. Scratch/**, Inbox/** |
| Key configured | yes/no |
| Last activity | Liveness |

Actions: add agent (token once), link key file, **Edit rules in FortSignal** (deep link), list only.

### Vault

**Job:** Human mental model + Obsidian coexistence.

- Root path, open folder  
- Conventions: `Inbox/`, `Scratch/`, `Private/`, `.fortmemory/`  
- Index stats, reindex button  
- Note: human Obsidian edits are OK; may lack signalId  

### Settings (secondary)

- FortSignal API base / key env (never show raw key)  
- **Deep link: Policy Composer + Dashboard**  
- Default path policy summary (local)  
- Tunnel docs (Cloudflare primary)  
- Advanced / danger  

**Composer lives here**, not as step 1 of onboarding wall.

---

## Agent-facing structure (API / MCP)

Preferred tools (names stable):

```
memory_search({ q, topK?, pathPrefix? })
memory_read({ path })
memory_write({ path, content, mode? }) → { decision, signalId } | { decision: "deny", reason }
memory_delete({ path })  // later
```

**Deny reasons must be stable and showable in UI** (examples):

- `path_not_allowed`  
- `action_not_allowed`  
- `amount_exceeds_policy`  
- `delegation_invalid`  
- `parameters_tampered`  
- `fortsignal_unavailable`  

Agents and the Activity feed use the same vocabulary.

**Supported path:** agents talk to FortMemory.  
**Unsupported:** raw filesystem writes to the vault in official guides.

---

## Default vault structure (product contract)

```
Vault/
  Inbox/          # agent landings, human triage
  Scratch/        # agent drafts / working memory
  Private/        # local deny by default for agents
  Projects/       # optional human + scoped agent notes
  .fortmemory/    # config, index, receipts, agents.json (hidden)
```

Init should create these and document them.  
Default local + FortSignal recipient templates should align (`personal/Scratch/*`, `personal/Inbox/*`).

---

## Default policies (no Composer required on day one)

On init / first agent, ship **safe defaults** (local jail + recommended FortSignal NL):

1. **Research (default)** — write Scratch + Inbox only; no Private; no delete; size cap  
2. **Read-only** — search/read only  
3. **Coder notes** — narrow path for AGENT_NOTES.md  

Power users open Composer to refine.  
**Value path must work before they ever see Composer.**

---

## Demo script (defines “click”)

1. Search finds an existing note (value)  
2. Agent write → note appears under Scratch/ (value + Obsidian)  
3. Agent write Private/ → deny in Activity (control)  
4. Copy signalId on allow (proof)  
5. Optional: “Edit rules” → FortSignal (governance depth)

If demos start at Composer, product messaging failed.

---

## What not to build in UI (for now)

- Heavy branding / illustration systems  
- Graph as core navigation  
- Chat-first “second brain” as primary shell  
- In-app fork of NL Policy Composer  
- Requiring Obsidian plugin for core loops  

---

## Implementation phases (UI)

| Phase | Ship |
|-------|------|
| MVP | CLI + HTTP; optional single static page: health + activity + search |
| Next | Full IA above (Home/Activity/Agents/Vault/Settings) |
| Later | Thin Obsidian plugin (status + recent signals + open Composer) |
| Later | MCP tool surface documented in UI “Connect agent” |

---

## Success metrics (product, not only security)

| Signal | Means |
|--------|--------|
| Search + write used weekly | Memory value |
| Activity feed checked | Trust/debug value |
| Time-to-first successful recall | Onboarding |
| Composer visits | Power-user depth (secondary) |
| Deny rate with clear reasons | Control without confusion |

---

## One-line reminder for future agents

> Structure the UI around **memory operations and proof of those operations** — not around policy authoring. Policy authoring is FortSignal, linked from Settings.
