# Landing Page Copy — FortMemory (Validation)

Use for a waitlist page (fortsignal.com/memory or fortmemory.dev). Keep visual minimal: logo, short diagram, form.

---

## Hero

**Eyebrow:** FortSignal · Local-first memory  

**Headline:** Memory you can prove.

**Subhead:** Persistent agent memory as real Markdown on your machine — every write gated by FortSignal cryptographic intent, policy, and `signalId` receipts.

**Primary CTA:** Join the waitlist  
**Secondary CTA:** Read the architecture  

**Microcopy under form:** Local by default. No cloud vault required. Built by the team behind FortSignal.

---

## Problem (3 beats)

### Session tokens are not memory governance

Agents store notes in vector DBs and chat logs. Login proves identity — not whether that write was authorized.

### Opaque memory is a liability

When memory is an unreadable embedding store, you cannot audit what changed, who changed it, or whether policy allowed it.

### Cloud-first memory fights privacy

Personal agents and regulated teams need long-term memory that stays on disk — openable in Obsidian — not locked in a vendor brain.

---

## Solution

**FortMemory** is a local memory server for agentic systems:

1. Point it at your Markdown / Obsidian vault  
2. Agents search and write through a localhost API  
3. Every mutation runs FortSignal challenge → sign → verify  
4. Allow returns a `signalId`; deny leaves the vault untouched  

```
Agent → FortMemory (Go) → FortSignal allow/deny → Markdown file + receipt
```

---

## Who it’s for

- Indie builders running local LLMs and agents  
- Teams that need **accountable** agent knowledge ops  
- FortSignal users who already have agent passports and policies  
- Anyone who refuses to put private memory in a black-box SaaS  

---

## What it’s not

- Not another “SOTA memory benchmark” race  
- Not a pure browser app  
- Not a replacement for FortSignal (it **uses** FortSignal)  
- Not multi-master cloud CRDT day one  

---

## Differentiation

| Typical agent memory | FortMemory |
|----------------------|------------|
| Cloud default | Local-first |
| Opaque vectors | Human-readable Markdown |
| Best-effort logs | FortSignal `signalId` receipts |
| Policy optional | Policy + passports on writes |

**Line:** Mem0 makes agents remember. FortMemory makes agent memory **accountable**.

---

## How it works (short)

1. **Install** one Go binary — `fortmemory serve`  
2. **Connect** an existing vault folder  
3. **Delegate** an agent in the FortSignal dashboard  
4. **Write** via API — only on `decision: allow`  

Optional later: Cloudflare Tunnel (mTLS), Tailscale, peer share, FortVault cloud sync (R2).

---

## Social proof block (placeholder)

> “I want my coding agent to keep notes I can open in Obsidian — and a receipt when it overwrites something important.”  
> — target user quote (replace with real interviews)

---

## FAQ

**Does my data leave my machine?**  
Not by default. Cloud sync (FortVault) is optional and later.

**Do I need FortSignal?**  
Yes for governed writes. FortMemory is the memory state plane; FortSignal is the enforcement plane.

**Can humans still edit files?**  
Yes. Obsidian/VS Code always work. External edits are indexed and labeled as ungoverned human edits.

**Browser app?**  
No. Local daemon + optional localhost dashboard. Agents need a real process.

**Open source?**  
Open-core planned: local server free; team cloud commercial.

---

## Final CTA

**Headline:** Prove every memory write.

**Subhead:** Join the waitlist for local dogfood builds.

**Button:** Request access  

**Contact:** hr@fortsignal.com  

---

## SEO / meta

- **Title:** FortMemory — Verifiable local agent memory  
- **Description:** Local-first Markdown memory for AI agents with FortSignal cryptographic receipts on every write.  
- **OG line:** Memory you can prove.  

---

## Interview CTAs (same page secondary)

- Book a 15-min founder call  
- “I run local agents” checkbox on form  
- “I need compliance/audit” checkbox  
