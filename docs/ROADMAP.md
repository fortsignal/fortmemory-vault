# Roadmap

## Phase 0 — Validation (2–3 weeks)

**Goal:** Evidence of demand before architecture debt.

- Landing one-pager + waitlist  
- 10–15 interviews (local LLM / agent builders)  
- Wizard-of-Oz: Markdown + FortSignal hash demo (script)  
- Kill / proceed gate ([VALIDATION.md](./VALIDATION.md))

**Exit:** Proceed to Phase 1 or park as FortSignal vertical narrative only.

---

## Phase 1 — Local core (MVP)

**Goal:** Go daemon + vault + search + FortSignal-gated writes.

- Project layout per [PROJECT-LAYOUT.md](./PROJECT-LAYOUT.md)  
- CLI: init / serve / reindex  
- API: health, search, read, write, delete, receipts  
- FTS5; optional Ollama embed  
- Path jail, single writer  
- Curl docs + demo script  

**Exit:** Dogfood daily for 14 days on a real vault.

---

## Phase 2 — Agent & policy depth

- Full action taxonomy docs + templates for FortSignal Composer  
- Sensitivity-gated reads  
- MCP tools: `memory_search`, `memory_read`, `memory_write`  
- Adapters notes: Deep Agents, OpenCode, generic HTTP agents  
- Local path/tag policy file  
- Optional minimal localhost activity UI  

**Exit:** 3 external builders run agents against it.

---

## Phase 3 — Remote access & soft P2P

- Cloudflare Tunnel + mTLS guide (primary) + optional CLI helper  
- Tailscale guide (supported)  
- Invite-based peer profiles  
- `memory.share_request` / grant with Signal Views  
- Provenance frontmatter for imported notes  

**Exit:** Two machines share a runbook under policy with dual signalIds.

---

## Phase 4 — FortVault hybrid

- Encrypted R2 object sync  
- Device pairing  
- Team shared vault  
- Admin audit export  
- Pricing tier live  

**Exit:** One design-partner team on hybrid.

---

## Phase 5 — Enterprise

- Self-host runbooks with FortSignal enterprise  
- mTLS  
- SIEM export  
- Retention / legal hold hooks  
- Security review package  

---

## Sequencing rules

1. Do not start Phase 3+ before Phase 1 dogfood.  
2. Do not build FortVault to “look complete.”  
3. Prefer docs + scripts over features when blocked on FortSignal product changes.  
4. Keep FortMemory process separate from any coding-agent monolith.  

## Near-term milestone checklist

- [ ] ADRs reviewed (this repo)  
- [ ] Phase 0 interviews logged  
- [ ] Go module scaffold  
- [ ] FortSignal write path green in integration test  
- [ ] First external dogfooder  
