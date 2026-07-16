# FortMemory

**Local-first agent memory for Obsidian-style Markdown vaults — with FortSignal proof on every write.**

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](./LICENSE)
[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)](https://go.dev/)

> Memory you can prove.

Agents get a real vault of human-readable notes. Humans keep editing in [Obsidian](https://obsidian.md). High-stakes memory operations go through [FortSignal](https://fortsignal.com) — parameter-bound signatures, agent passports, policies, and immutable `signalId` receipts.

| Layer | Role |
|-------|------|
| **FortMemory** (this repo) | Local daemon · CLI · HTTP API · search · MCP |
| **FortSignal** | Governance / enforcement (SaaS or self-host) |
| **FortVault** | Optional cloud sync (commercial, later) |

**Open-core:** this repository is free forever (Apache-2.0). FortSignal and FortVault are separate commercial surfaces. See [docs/OPEN-CORE.md](./docs/OPEN-CORE.md).

---

## Features

- **Local-first** — data stays in your Markdown vault; no cloud required  
- **Obsidian-compatible** — plain files; engine state under `.fortmemory/`  
- **FortSignal-gated writes/deletes** — allow/deny + `signalId`  
- **HTTP API** for agents — search, read, write, delete, receipts  
- **MCP tools** — `memory_search` / `memory_read` / `memory_write` / `memory_delete`  
- **FTS search** — SQLite FTS5; optional Ollama hybrid embeddings  
- **Vault watcher** — reindex when you edit notes in Obsidian  
- **Minimal dashboard** — Home · Search · Activity · Agents · Settings  
- **Remote helpers** — Cloudflare Tunnel (primary for hostnames) · Tailscale (mesh)

---

## Quick start

### Requirements

- Go **1.23+**  
- A folder of Markdown notes (or empty path to init)  
- [FortSignal](https://fortsignal.com) API key + agent passport for **governed writes** (reads/search work with a local token alone)

### Install

```bash
git clone https://github.com/fortsignal/fortmemory-vault.git
cd fortmemory-vault
make test && make build
# binary: ./bin/fortmemory

# or, after the module is published on the default branch:
# go install github.com/fortsignal/fortmemory-vault/cmd/fortmemory@latest
```

### Run locally (no domain / tunnel)

```bash
./bin/fortmemory init ~/Vaults/Personal --id personal

# 1) FortSignal dashboard: create agent passport, download key JSON
# 2) Policy: allow memory.write (and memory.delete if needed)
#    recipients e.g. personal/Scratch/*, personal/Inbox/*
# 3) Passkey-approve delegation for that agent

export FORTSIGNAL_API_KEY=fs_live_...

./bin/fortmemory agent add research-01 \
  --config ~/Vaults/Personal/.fortmemory/config.toml \
  --key ~/Downloads/agent-key.json
# save the fm_… bearer token shown once

./bin/fortmemory reindex --config ~/Vaults/Personal/.fortmemory/config.toml
./bin/fortmemory serve --config ~/Vaults/Personal/.fortmemory/config.toml
```

- **API / dashboard:** http://127.0.0.1:7432/  
- Paste the `fm_…` token in the dashboard for search & activity  

```bash
curl -sS http://127.0.0.1:7432/v1/health
curl -sS -H "Authorization: Bearer fm_…" \
  "http://127.0.0.1:7432/v1/read?path=Scratch/hello.md"
curl -sS -H "Authorization: Bearer fm_…" -H 'Content-Type: application/json' \
  -d '{"q":"preferences","topK":5}' http://127.0.0.1:7432/v1/search
```

### CLI write (same FortSignal gate)

```bash
./bin/fortmemory write \
  --config ~/Vaults/Personal/.fortmemory/config.toml \
  --key ~/Downloads/agent-key.json \
  --path Scratch/hello.md \
  --body $'# Hello\n\nGoverned memory.\n'
```

### MCP (Cursor / Claude)

```bash
./bin/fortmemory mcp \
  --config ~/Vaults/Personal/.fortmemory/config.toml \
  --agent research-01 \
  --key ~/Downloads/agent-key.json
```

Example client config: [examples/mcp.cursor.json](./examples/mcp.cursor.json) · guide: [docs/MCP.md](./docs/MCP.md)

### Optional: hybrid embeddings (Ollama)

```toml
# in .fortmemory/config.toml
[embeddings]
provider = "ollama"
ollama_url = "http://127.0.0.1:11434"
model = "nomic-embed-text"
```

If Ollama is down, search stays FTS-only.

---

## Architecture (one picture)

```
Humans:  Obsidian (notes)  +  FortSignal dashboard (policies, passports)
Agents:  HTTP / MCP  →  FortMemory (this repo)
                        ├─ path jail + local policy
                        ├─ FortSignal challenge/start → sign → verify
                        └─ write Markdown only on allow + signalId
```

Full system design: [docs/SYSTEM-ARCHITECTURE.md](./docs/SYSTEM-ARCHITECTURE.md)  
Product / UI IA (memory-first, not Composer lobby): [docs/PRODUCT-SURFACE.md](./docs/PRODUCT-SURFACE.md)

---

## Remote access (optional)

| Situation | Approach |
|-----------|----------|
| Same machine | `http://127.0.0.1:7432` — **default** |
| Devices on your Tailscale tailnet | `fortmemory tailscale print-serve` |
| Hostname / Access / mTLS | `fortmemory cloudflare …` |

```bash
fortmemory tailscale                 # guide + status
fortmemory tailscale print-serve     # → tailscale serve --bg http://127.0.0.1:7432

fortmemory cloudflare install
fortmemory cloudflare check
```

Keep `bind = "127.0.0.1"`. Do not expose the API raw on the public internet.  
Docs: [docs/REMOTE-ACCESS.md](./docs/REMOTE-ACCESS.md)

---

## HTTP API (MVP)

| Method | Path | Auth |
|--------|------|------|
| `GET` | `/v1/health` | no |
| `GET` | `/v1/read?path=` | bearer |
| `POST` | `/v1/search` | bearer |
| `POST` | `/v1/write` | bearer + FortSignal |
| `POST` | `/v1/delete` | bearer + FortSignal |
| `GET` | `/v1/receipts` | bearer |
| `GET` | `/v1/agents` | bearer |
| `GET` | `/` | dashboard |

OpenAPI sketch: [docs/openapi.yaml](./docs/openapi.yaml)

---

## Project layout

```
cmd/fortmemory/          CLI entrypoint
internal/                server, vault, index, fortsignal client, mcp, …
web/ + internal/server/static/   operator dashboard
docs/                    architecture, integration, product
examples/                MCP client config sample
```

---

## Documentation

| Doc | Topic |
|-----|--------|
| [docs/INDEX.md](./docs/INDEX.md) | Full index |
| [docs/SYSTEM-ARCHITECTURE.md](./docs/SYSTEM-ARCHITECTURE.md) | FortSignal × FortMemory × Obsidian |
| [docs/FORTSIGNAL-INTEGRATION.md](./docs/FORTSIGNAL-INTEGRATION.md) | API-aligned governance contract |
| [docs/MVP.md](./docs/MVP.md) | Scope |
| [docs/OPEN-CORE.md](./docs/OPEN-CORE.md) | OSS vs commercial |
| [docs/SECURITY.md](./docs/SECURITY.md) | Threat model |
| [docs/PUBLISH.md](./docs/PUBLISH.md) | Maintainers |

---

## Security

- Default bind: **127.0.0.1 only**  
- Local API tokens (`fm_…`) ≠ FortSignal API keys (`fs_live_…`)  
- Agent Ed25519 private keys never leave your machine  
- Mutates fail closed if FortSignal is unreachable  
- Report vulnerabilities: **hr@fortsignal.com** (subject: `FortMemory security`)

See [docs/SECURITY.md](./docs/SECURITY.md) and [CONTRIBUTING.md](./CONTRIBUTING.md).

---

## License

Copyright 2026 Jeffrey Walters / FortSignal.

Licensed under the **Apache License, Version 2.0** — see [LICENSE](./LICENSE) and [NOTICE](./NOTICE).

**Trademarks:** FortMemory, FortVault, and FortSignal are trademarks. The Apache License does not grant trademark rights.

---

## Related

- [FortSignal](https://fortsignal.com) — You approve the action. We enforce it.  
- [FortSignal developer docs](https://fortsignal.com/docs)  
- [NL Policy Composer](https://fortsignal.com/composer)  
