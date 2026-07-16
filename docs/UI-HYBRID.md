# UI Strategy — Hybrid Local-First (Not Browser-First)

**Status:** Accepted (ADR-011)  
**Decision:** FortMemory is **not** a pure browser app for MVP or as the primary product shape.

## Verdict

| Approach | Role |
|----------|------|
| **Go local daemon** | Primary product (memory authority, FS, index, FortSignal gate) |
| **System tray / desktop shell** | Optional lightweight host (Tauri **or** pure Go tray later) |
| **Embedded web dashboard** | Secondary UI at `http://127.0.0.1:7432/` |
| **Obsidian plugin** | Strong add-on (talks to local API) |
| **Pure browser SaaS memory** | **Rejected** as main product |

## Why not pure browser MVP

1. **No reliable Obsidian vault FS access** — browsers cannot watch/index arbitrary user folders.  
2. **Weaker local crypto/key custody** for agent signing and API keys.  
3. **Sandbox fights local-first** — the core promise is data on disk.  
4. **Agents need a stable `localhost` API** — a tab is not a reliable service.  
5. **mTLS / tunnels / daemons** are natural from a native process, awkward from a SPA.  
6. **Performance** — fsnotify, SQLite FTS, embedding queues belong in a process, not WebWorker theater.

## Recommended product surface (phased)

```
┌─────────────────────────────────────────────────────────────┐
│  User machine                                               │
│                                                             │
│   fortmemory (Go)  ◄── agents / MCP / CLI / Obsidian plugin │
│        │                                                    │
│        ├── vault FS (Markdown)                              │
│        ├── SQLite index                                     │
│        ├── FortSignal HTTPS client                          │
│        └── serves static dashboard at /                     │
│                                                             │
│   [optional] tray icon → open dashboard / start-stop        │
│   [optional] Tauri webview wrapping same localhost UI       │
└─────────────────────────────────────────────────────────────┘
```

### Phase 1 (MVP)

1. **CLI + daemon only** (`fortmemory serve`)  
2. **JSON API** for agents  
3. **Minimal embedded HTML dashboard** (health, activity, search) — open in any browser  
4. **No Tauri required**

### Phase 2

5. System tray (Go `systray` or similar) for status  
6. Richer dashboard (agents list, deep-link to FortSignal Composer)  

### Phase 3+

7. Optional Tauri wrapper if distribution/UX needs a “real app”  
8. Obsidian community plugin  

## Desktop vs browser tradeoffs

| Dimension | Pure browser | Hybrid (daemon + local web UI) |
|-----------|--------------|--------------------------------|
| Vault access | Poor | Native FS |
| Agent localhost API | Unreliable | Stable |
| Security posture | API keys in browser risk | Keys in process / env |
| Install friction | Lowest | One binary |
| Auto-start | Hard | OS service / tray |
| UI polish velocity | High | Medium (static or light SPA) |
| Local-first story | Weak | Strong |

## UI flows (local server + browser dashboard)

### First-run

1. User downloads `fortmemory`  
2. `fortmemory init ~/Vaults/Personal`  
3. Sets `FORTSIGNAL_API_KEY`  
4. `fortmemory serve`  
5. Opens `http://127.0.0.1:7432` → sees vault path, FortSignal configured badge  
6. `fortmemory agent add research-01` → token once  
7. Approves FortSignal delegation in fortsignal.com/dashboard  

### Daily

- Tray/daemon green = running  
- Browser dashboard for search + activity/`signalId`  
- Agents hit API without opening UI  

### Sensitive write (human step-up, later)

- Dashboard triggers human challenge via FortSignal WebAuthn  
- Daemon holds API key; browser only does WebAuthn ceremony on localhost origin (RPID constraints must be designed carefully — prefer agent-only mutates for MVP)

## Obsidian plugin (add-on, not MVP)

- Search vault via FortMemory  
- Show last `signalId` on note  
- Quick “write via governed API” for agent-origin notes  
- Status: daemon up/down  

Does **not** replace the daemon.

## Explicit non-goals

- Hosted multi-tenant SPA as the only way to use FortMemory  
- Electron mega-app  
- Rebuilding FortSignal Composer inside FortMemory  

Policies stay in FortSignal Composer; dashboard deep-links.
