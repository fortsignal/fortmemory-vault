# Concrete Next Steps

Ordered for a solo founder. Docs in this repo are complete for design phase.

## This week (no large code)

1. **Read** [PRODUCT.md](./PRODUCT.md) + [DECISIONS.md](./DECISIONS.md) and mark any ADR dissent.  
2. **Confirm** Go decision stays (ADR-001).  
3. **Run** 5 interviews with [VALIDATION.md](./VALIDATION.md) script.  
4. **Draft** public one-pager outline (problem, diagram, waitlist).  
5. **Create** 3 FortSignal Composer memory policy templates from [FORTSIGNAL.md](./FORTSIGNAL.md).  

## Done in repo

- Docs suite (product, architecture, hybrid UI, FortSignal API alignment)  
- Go skeleton: `cmd/fortmemory`, `internal/*` stubs, `go.mod`  
- Working unit tests: `internal/fortsignal` recipient encoding + content hash  
- Landing copy: [LANDING.md](./LANDING.md)  

## Context for future sessions

| Topic | Doc |
|-------|-----|
| Whole system architecture | [SYSTEM-ARCHITECTURE.md](./SYSTEM-ARCHITECTURE.md) |
| Product/UI structure (not Composer-first) | [PRODUCT-SURFACE.md](./PRODUCT-SURFACE.md) |
| Strategy freeze | [STRATEGY-LOCK.md](./STRATEGY-LOCK.md) |
| Build status | [IMPLEMENTATION.md](./IMPLEMENTATION.md) |

Demand is validated — prefer building. Do **not** expand Cloudflare engineering without need.

## Implementation progress

See [IMPLEMENTATION.md](./IMPLEMENTATION.md).

**Done:** config/init · vault jail · FortSignal client · `memory.Write` · CLI `write` · light CF helpers  

**When unfreezing code (in order):**

1. `serve` + agent bearer tokens  
2. Read + FTS search  
3. Watcher / reindex  
4. Dogfood 14 days  
5. Tunnel polish only if demanded  

## Validation (parallel, high leverage)

1. Ship waitlist from [LANDING-OUTLINE.md](./LANDING-OUTLINE.md)  
2. Run interviews per [VALIDATION.md](./VALIDATION.md)  

## Explicitly not next

- R2 / FortVault  
- Peer protocol implementation  
- Tauri  
- Graph DB  
- Merging into coding-agent monorepo  

## Optional artifacts on request

- `cmd/fortmemory` Go skeleton  
- Expanded package file stubs with comments  
- Landing page copy  
- Interview tracker spreadsheet template  
