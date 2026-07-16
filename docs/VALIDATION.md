# Validation Plan (Phase 0)

**Goal:** Reduce traction risk before investing in full Go implementation.

## Hypothesis

Agent builders and local-LLM users will run a local daemon that makes memory writes **accountable** (FortSignal receipts) if:

1. Setup is under 15 minutes  
2. Notes remain normal Markdown  
3. Read path stays low-friction  

## Experiments

### 1. Problem interviews (N=10–15)

**Script (10 min):**

1. How do your agents remember across sessions today?  
2. Where does data live? Who can read it?  
3. Any bad write / secret leak / wrong memory incident?  
4. Accept biometric or policy deny on sensitive memory writes?  
5. Local daemon vs hosted API preference?  
6. What monthly price if it prevented one incident?  

**Log:** persona, current stack, pain (1–5), WTP, quote.

### 2. Waitlist landing

One page: problem → diagram → “Memory you can prove” → email capture.

**Success:** ≥50 emails from builders in 3 weeks (adjust for traffic reality; qualitative > vanity).

### 3. Wizard of Oz demo

Script (bash/python) that:

- Writes Markdown to a vault folder  
- Computes content hash  
- Calls FortSignal challenge/verify for a test agent  
- Prints `signalId`  

Share Loom (2 min). Measure replies: “I want this as a daemon.”

### 4. Composer templates

Publish 3 memory policy templates as FortSignal Composer examples. Measure engagement from existing FortSignal audience.

## Decision gate

| Signal | Build Phase 1 | Park / narrow |
|--------|---------------|---------------|
| Interview “must have” | ≥5/12 | <3/12 |
| You dogfood script ≥4 days/week | yes | abandon |
| FortSignal synergy asks | users request agent memory | only payments interest |
| Waitlist quality | builders with agents in prod/side | generic curiosity only |

**Narrow alternative if partial pull:** sell as “governed knowledge actions” vertical on FortSignal without separate FortMemory brand.

## Anti-patterns

- Building FortVault because demos look better with cloud  
- Optimizing embedding benchmarks before first external user  
- Forking UI/policy systems that FortSignal already owns  

## After go decision

1. Scaffold Go module per [PROJECT-LAYOUT.md](./PROJECT-LAYOUT.md)  
2. 14-day personal dogfood  
3. One design partner  
