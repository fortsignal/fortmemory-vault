# FortMemory Documentation Index

**Product:** FortMemory (local) · FortVault (cloud, later)  
**Repo:** `fortmemory-vault`  
**Stack decision:** Go local memory server  
**Governance:** FortSignal on every mutating operation  
**Status:** Design docs — pre-implementation  

## Start here

0. [SYSTEM-ARCHITECTURE.md](./SYSTEM-ARCHITECTURE.md) — **whole system: FS × FM × Obsidian**  
1. [STRATEGY-LOCK.md](./STRATEGY-LOCK.md) — frozen strategy + MVP + open-core  
2. [FOUNDING-CONTEXT.md](./FOUNDING-CONTEXT.md) — canonical founder context  
3. [PRODUCT.md](./PRODUCT.md) — what we are building and why  
4. [MVP.md](./MVP.md) — exact MVP scope  
5. [OPEN-CORE.md](./OPEN-CORE.md) — open-core boundaries  
6. [ARCHITECTURE.md](./ARCHITECTURE.md) — FortMemory-internal layers  
7. [FORTSIGNAL-INTEGRATION.md](./FORTSIGNAL-INTEGRATION.md) — wire-level FS API  
8. [STACK.md](./STACK.md) · [DECISIONS.md](./DECISIONS.md)  

## Build surface

| Doc | Audience |
|-----|----------|
| [PROJECT-LAYOUT.md](./PROJECT-LAYOUT.md) | Implementers (Go tree) |
| [API.md](./API.md) | API consumers |
| [openapi.yaml](./openapi.yaml) | Codegen / contract |
| [CLI.md](./CLI.md) | Operator UX |
| [FORTSIGNAL.md](./FORTSIGNAL.md) | Integration overview |
| [FORTSIGNAL-INTEGRATION.md](./FORTSIGNAL-INTEGRATION.md) | **API-aligned contract** (from fortsignal-api) |
| [PRODUCT-SURFACE.md](./PRODUCT-SURFACE.md) | **UI IA, value-first UX, agent surfaces** |
| [UI.md](./UI.md) | Dashboard screen specs |
| [UI-HYBRID.md](./UI-HYBRID.md) | Desktop/local hybrid delivery |
| [LANDING.md](./LANDING.md) | Validation landing page copy |
| [OPEN-CORE.md](./OPEN-CORE.md) | **Open-core / monetization strategy** |
| [COMMERCIAL.md](./COMMERCIAL.md) | Paid / proprietary boundary |
| [REMOTE-ACCESS.md](./REMOTE-ACCESS.md) | **Local vs Tailscale vs Cloudflare** |
| [TAILSCALE.md](./TAILSCALE.md) | Tailscale supported remote |
| [CLOUDFLARE-TUNNEL.md](./CLOUDFLARE-TUNNEL.md) | Cloudflare Tunnel + mTLS (primary for hostname) |
| [IMPLEMENTATION.md](./IMPLEMENTATION.md) | Build status |
| [MCP.md](./MCP.md) | MCP tools for Cursor/Claude agents |
| [PUBLISH.md](./PUBLISH.md) | Public repo checklist |

## Product & GTM

| Doc | Audience |
|-----|----------|
| [ROADMAP.md](./ROADMAP.md) | Founder sequencing |
| [VALIDATION.md](./VALIDATION.md) | Pre-build demand tests |
| [COMPETITORS.md](./COMPETITORS.md) | Positioning |
| [UI.md](./UI.md) | Minimal dashboard |
| [PEER.md](./PEER.md) | Post-MVP sharing |
| [SECURITY.md](./SECURITY.md) | Threat model |

## Reading order for a new engineer

```
SYSTEM-ARCHITECTURE → PRODUCT-SURFACE → FORTSIGNAL-INTEGRATION → MVP → IMPLEMENTATION → UI
```

## Reading order for validation-only (no code yet)

```
PRODUCT → VALIDATION → MVP → ROADMAP → COMPETITORS
```
