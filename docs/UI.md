# UI Descriptions (Operator Dashboard)

**Authority for IA / product principles:** [PRODUCT-SURFACE.md](./PRODUCT-SURFACE.md)  
**Delivery:** [UI-HYBRID.md](./UI-HYBRID.md) — daemon + localhost web, not pure browser product  
**Whole system:** [SYSTEM-ARCHITECTURE.md](./SYSTEM-ARCHITECTURE.md)

---

## Priority

1. CLI + HTTP API (agents live here)  
2. Thin localhost dashboard with **memory-first IA**  
3. Optional tray / Obsidian plugin later  

Design posture: dense ops console — not lifestyle consumer app. Dark neutral, system fonts, monospace for `signalId` / agent ids. No custom illustration required.

---

## Principle (do not regress)

**Home = memory + activity.**  
**Settings = FortSignal Composer deep link.**  

Do not make Policy Composer the lobby or first-run wall.  
Default path policies exist so day-one works without Composer.

---

## Shell layout

```
┌─ sidebar ──────────┬─ main ─────────────────────────────┐
│ FortMemory         │  [ Search memory ............... ] │
│ ● Personal vault   │────────────────────────────────────│
│                    │                                    │
│ Home               │         page content               │
│ Search             │                                    │
│ Activity           │                                    │
│ Agents             │                                    │
│ Vault              │                                    │
│ Settings           │                                    │
│                    │                                    │
│ FortSignal ● linked│                                    │
└────────────────────┴────────────────────────────────────┘
```

---

## Screens

### Home

- Vault status: path, running, file/index counts  
- Recent activity table (allow/deny · agent · path · signalId)  
- Empty state: init vault → add agent → write Scratch/  

### Search

- Central query; results with path, excerpt, score, last_signal_id  
- Preview pane read-only Markdown  

### Activity

| Time | Decision | Agent | Action | Path | signalId | Reason |
|------|----------|-------|--------|------|----------|--------|
| … | allow/deny | … | memory.write | … | copy | … |

Killer differentiator vs opaque vector memory UIs.

### Agents

| agentId | Status | Scope summary | Key | Last seen |
|---------|--------|---------------|-----|-----------|
| research-01 | active | Scratch/**, Inbox/** | yes | … |

Actions: add (token once), configure key, **Edit rules in FortSignal**, revoke (dashboard link).

### Vault

- Path, open in OS / “works with Obsidian”  
- Folder conventions (Inbox, Scratch, Private, .fortmemory)  
- Reindex control  

### Settings

- FortSignal connection (configured yes/no, not raw secrets)  
- Deep links: Composer, Dashboard, docs  
- Local policy summary  
- Tunnel help (Cloudflare primary)  

---

## Decision badges

| State | Meaning |
|-------|---------|
| allow | green — governed write applied |
| deny | red + reason |
| human_external | gray — Obsidian/human edit, no signalId |
| step-up pending | yellow — later |

---

## Agent-facing (non-GUI)

Stable tools: `memory_search`, `memory_read`, `memory_write` → decision envelope.  
See PRODUCT-SURFACE.md for deny reason vocabulary.

---

## Out of scope for early UI

- Graph-first navigation  
- Chat shell as primary product  
- In-app NL policy compiler  
- Marketing-site visual system inside the daemon  

---

## Implementation note

Serve static UI from Go (`embed.FS`) at `/` when ready. MVP may be health JSON only until Activity/Search pages land.
