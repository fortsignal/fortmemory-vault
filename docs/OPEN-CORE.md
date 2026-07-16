# Open-Core Strategy — FortMemory

**Decision:** Open-source the local core; monetize cloud, team, and enterprise.  
**License (OSS core):** **Apache License 2.0**  
**Status:** Accepted (ADR-013)

## Why open-core (with a twist)

FortMemory’s wedge is **trust**:

- Local-first builders will not run a closed binary against their Obsidian vault without source.  
- Security tooling that cannot be audited loses the “verifiable” story.  
- Solo founder distribution needs GitHub / community pull; FortSignal already needs a funnel.  

The twist: **open the state plane (memory server), sell the control-plane extras and cloud**—and keep FortSignal itself as the paid governance meter where usage scales.

```
┌─────────────────────────────────────────────────────────────┐
│  OPEN SOURCE (this repo, Apache-2.0)                        │
│  fortmemory CLI · local API · vault · search · tunnels      │
│  FortSignal *client* · local policy · receipts on disk      │
└───────────────────────────┬─────────────────────────────────┘
                            │ uses
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  FORTSIGNAL (existing SaaS / self-host — separate product)  │
│  challenge/verify · passports · NL Composer · signalId      │
│  Monetized today (Pro / Business / Enterprise)              │
└───────────────────────────┬─────────────────────────────────┘
                            │ optional
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  COMMERCIAL (FortVault + enterprise packs — not this repo)  │
│  R2 sync · team admin · SSO · compliance export · support   │
└─────────────────────────────────────────────────────────────┘
```

**Monetization twist:**  
OSS drives adoption of FortMemory → every governed write can consume FortSignal verifications → FortVault Team sells sync/sharing → Enterprise sells compliance/SSO/self-host packages. You do **not** need to cripple the local server to make money.

---

## What is open source (Apache 2.0)

Ship in the public `fortmemory` / `fortmemory-vault` repository:

| Area | Include |
|------|---------|
| Local Memory Server | Go CLI + HTTP API |
| Vault | Path jail, Markdown R/W, watcher |
| Search | FTS / basic hybrid RAG hooks (local models) |
| FortSignal integration | HTTP client, examples, policy mapping docs |
| Local policy | Path globs, sensitivity basics |
| Receipts | Local JSONL/SQLite audit mirror |
| Tunnel helpers | **Cloudflare Tunnel (primary)** + Tailscale guides/CLI |
| Dashboard | Minimal localhost UI |
| Docs | Architecture, threat model, OpenAPI |

**License choice:** Apache 2.0 over MIT for:

- Explicit patent grant (better for crypto/security-adjacent infra)  
- Familiar open-core precedent (many infra projects)  
- Clear NOTICE / attribution path  

---

## What stays proprietary / paid

Do **not** put these in the OSS default binary as free forever features (separate product, private modules, or paid cloud):

| Area | Why paid |
|------|----------|
| **FortVault cloud sync** | R2 ops, multi-device, team shared vaults, billing |
| **Team control plane** | Device pairing, org roles, shared vault directory |
| **Advanced enterprise** | SSO/SAML, SIEM export polish, retention/legal hold UX, BAAs narrative pack |
| **Managed hosting** | “We run FortMemory + FortSignal for you” |
| **Premium support** | SLAs, onboarding, security review support |
| **Deep FortSignal product surface** | NL Composer itself, dashboard Agent Passports UI, advanced Signal Views admin — already FortSignal SaaS, not reimplemented in OSS |

### Boundary clarity (important)

| Component | Open? |
|-----------|--------|
| FortMemory calling FortSignal API | **Yes** (client code OSS) |
| FortSignal server / Composer / metering | **No** (existing commercial product) |
| “Complex policy composer features” inside FortMemory | **No** — deep-link to FortSignal Composer; don’t fork |
| Basic local path policy | **Yes** |
| Peer protocol (basic) | **Yes** eventually (drives FortSignal + Team) |
| Hosted peer directory / discovery marketplace | **Paid / commercial** if it costs infra |

---

## Pricing sketch (open-core aligned)

| Tier | Price ballpark | Gets |
|------|----------------|------|
| **Community** | $0 | OSS fortmemory forever; own vault; own FortSignal free/sandbox limits |
| **Pro** | ~$19–29/mo | Multi-vault polish, priority docs, email support (optional; don’t block OSS) |
| **FortVault Team** | ~$99–199/mo | R2 sync, devices, shared vaults, admin audit |
| **Enterprise** | Custom | Self-host assist, SSO, compliance packs, joint FortSignal deal |
| **FortSignal** | Existing plans | Verifications, agents, Composer — **usage grows with memory ops** |

**Do not:**  
- Require a license key to run local search/write.  
- Phone home from the OSS binary by default.  
- Hide FortSignal client behind paywall (that kills the product thesis).

**Do:**  
- Optional “Sign in to FortVault” for sync.  
- Clear upgrade CTAs in docs/dashboard: “Team sync → FortVault”.  
- Meter value on FortSignal verification volume + FortVault seats/storage.

---

## Repo / brand layout

```
fortmemory (public, Apache-2.0)     ← this codebase direction
fortsignal-api / dashboard (private or separate)  ← enforcement SaaS
fortvault (private or closed modules)             ← R2 + team later
```

Optional later:

- `fortmemory-enterprise` private modules via build tags — only if needed; prefer SaaS over binary DRM.

---

## Cloudflare Tunnel (“plugin”) in OSS

Not a browser extension. **Open-source helper** that:

1. Detects `cloudflared` on PATH  
2. Prints install instructions if missing  
3. Emits a ready-to-run tunnel config for local FortMemory port  
4. Documents **mTLS** client cert posture for remote agents  

Tailscale remains a **supported** alternative guide.

Commercial angle: managed tunnel + identity for teams lands in **FortVault Team**, not a locked OSS build.

---

## Community rules (keep the moat)

1. **Accept PRs** on vault, index, tunnels, docs, bug fixes.  
2. **Decline PRs** that reimplement FortSignal Composer or R2 multi-tenant billing in OSS.  
3. **Trademark:** “FortMemory” / “FortVault” / “FortSignal” remain company marks; Apache license ≠ trademark license.  
4. **Security:** private disclosure path for vulns before public issues.  

---

## Go-to-market using OSS

1. Public repo + “Memory you can prove” README  
2. `curl | sh` or `go install` one-liner  
3. Deep Agents / OpenCode example using open client  
4. Every demo ends with a `signalId` (FortSignal upsell)  
5. “Need sync across laptops?” → FortVault waitlist  

---

## Solo-founder risk controls

| Risk | Mitigation |
|------|------------|
| Cloud competitors rehost OSS | Moat is FortSignal governance + brand + cloud ops, not embeddings |
| Support burden | GitHub discussions; paid support tier |
| Premature FortVault build | Validate OSS dogfood first |
| License confusion | Single Apache-2.0 LICENSE at root; COMMERCIAL.md for paid map |

---

## Checklist before public GitHub

- [ ] LICENSE + NOTICE (Apache-2.0)  
- [ ] README open-core section  
- [ ] No secrets / personal vaults in repo  
- [ ] `fortmemory write` demo path documented  
- [ ] SECURITY.md + vulnerability contact  
- [ ] Trademark line in README  
- [ ] Cloudflare tunnel helper usable  
