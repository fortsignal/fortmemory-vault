<div align="center">

# FortMemory

**Local-first agentic memory for Markdown vaults — with cryptographic proof on every write**

Memory you can prove.

Obsidian-compatible files · FortSignal-governed mutates · HTTP + MCP for agents

<br/>

[![License](https://img.shields.io/badge/license-Apache%202.0-blue?style=flat-square)](./LICENSE)
[![Go](https://img.shields.io/badge/go-1.23+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev/)
[![FortSignal](https://img.shields.io/badge/FortSignal-governance-111827?style=flat-square)](https://fortsignal.com)
[![Obsidian](https://img.shields.io/badge/Obsidian-compatible-7C3AED?style=flat-square&logo=obsidian&logoColor=white)](https://obsidian.md)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macOS%20%7C%20windows-6b7280?style=flat-square)](#quick-start)
[![Local-first](https://img.shields.io/badge/local--first-yes-22c55e?style=flat-square)](#features)
[![MCP](https://img.shields.io/badge/MCP-supported-8b5cf6?style=flat-square)](#mcp)
[![Status](https://img.shields.io/badge/status-MVP-f59e0b?style=flat-square)](#status)

[![Stars](https://img.shields.io/github/stars/fortsignal/fortmemory-vault?style=flat-square&logo=github)](https://github.com/fortsignal/fortmemory-vault/stargazers)
[![Forks](https://img.shields.io/github/forks/fortsignal/fortmemory-vault?style=flat-square&logo=github)](https://github.com/fortsignal/fortmemory-vault/network/members)
[![Issues](https://img.shields.io/github/issues/fortsignal/fortmemory-vault?style=flat-square)](https://github.com/fortsignal/fortmemory-vault/issues)
[![Last commit](https://img.shields.io/github/last-commit/fortsignal/fortmemory-vault?style=flat-square)](https://github.com/fortsignal/fortmemory-vault/commits/main)

<br/>

[Quick start](#quick-start)
&nbsp;·&nbsp;
[Features](#features)
&nbsp;·&nbsp;
[Architecture](#architecture)
&nbsp;·&nbsp;
[HTTP API](#http-api)
&nbsp;·&nbsp;
[MCP](#mcp)
&nbsp;·&nbsp;
[Remote access](#remote-access)
&nbsp;·&nbsp;
[Docs](#documentation)
&nbsp;·&nbsp;
[Security](#security)
&nbsp;·&nbsp;
[License](#license)

</div>

---

## Why FortMemory?

Most agent “memory” is either:

- a **black-box vector DB** you can’t open, or  
- a **raw folder write** with no authorization proof  

FortMemory is both **human-readable** and **governed**:

| Problem | FortMemory |
|---------|------------|
| Opaque embeddings-only stores | Real **Markdown** on disk (open in Obsidian) |
| Agents overwrite anything | **Path policy** + FortSignal **allow/deny** |
| No audit trail | Immutable **`signalId`** receipts |
| Cloud-by-default | **Local-first** (no domain required) |

```
Agent / MCP / CLI
       │
       ▼
 FortMemory (this repo)  ──mutate──►  FortSignal challenge / verify
       │                                    │
       │ allow + signalId                   │
       ▼                                    ▼
 Markdown vault + FTS index           Audit receipt (local + FortSignal)
```

---

## Ecosystem

| Product | Role | License / model |
|---------|------|-----------------|
| **FortMemory** | Local memory server · CLI · API · MCP · dashboard | **Apache-2.0** (this repo) |
| **[FortSignal](https://fortsignal.com)** | Passports · policies · NL Composer · `signalId` enforcement | Separate SaaS / enterprise |
| **FortVault** | Optional team cloud sync (R2) | Commercial (later) |

Open-core details: [docs/OPEN-CORE.md](./docs/OPEN-CORE.md)

---

## Features

### Memory product
- **Obsidian-compatible vaults** — plain `.md` files; engine state in `.fortmemory/`
- **Full-text search** — SQLite FTS5 (Porter tokenizer)
- **Optional hybrid embeddings** — Ollama (`nomic-embed-text`, etc.)
- **Vault watcher** — reindex on external edits (e.g. Obsidian)
- **Operator dashboard** — Home · Search · Activity · Agents · Settings
- **Agent tokens** — local `fm_…` bearer auth for the HTTP API

### Governance (via FortSignal)
- **`memory.write` / `memory.delete`** gated by challenge → Ed25519 sign → verify  
- **Agent passports** + passkey-approved **delegations**  
- **NL Policy Composer** for path/action rules (deep-linked; not reimplemented here)  
- **Local receipt log** + optional frontmatter `last_signal_id`

### Agent integration
- **REST API** on `127.0.0.1:7432`  
- **MCP stdio tools**: `memory_search`, `memory_read`, `memory_write`, `memory_delete`  
- **CLI** for init, write, delete, serve, reindex, agent, mcp, tunnels  

### Remote (optional)
- **Tailscale** — mesh (`fortmemory tailscale`)  
- **Cloudflare Tunnel** — hostname / Access / mTLS (`fortmemory cloudflare`)  
- Default remains **localhost only**

---

## Quick start

### Requirements

| Requirement | Notes |
|-------------|--------|
| **Go 1.23+** | Build toolchain |
| **Markdown folder** | Existing Obsidian vault or empty path |
| **FortSignal** | API key + agent passport for **writes/deletes** |
| Optional: **Ollama** | Hybrid semantic search |

Reads/search need only a local FortMemory agent token. Mutates also need FortSignal.

### Install

```bash
git clone https://github.com/fortsignal/fortmemory-vault.git
cd fortmemory-vault
make test && make build
# → ./bin/fortmemory

# After main is on GitHub:
go install github.com/fortsignal/fortmemory-vault/cmd/fortmemory@latest
```

### 60-second local path

```bash
# 1. Initialize vault metadata
./bin/fortmemory init ~/Vaults/Personal --id personal

# 2. FortSignal (dashboard)
#    - Create agent passport → download agent-key.json
#    - Policy: memory.write (+ memory.delete if needed)
#      recipients: personal/Scratch/*, personal/Inbox/*
#    - Passkey-approve delegation

export FORTSIGNAL_API_KEY=fs_live_...

# 3. Local API agent (prints fm_ token once)
./bin/fortmemory agent add research-01 \
  --config ~/Vaults/Personal/.fortmemory/config.toml \
  --key ~/path/to/agent-key.json

# 4. Index + serve
./bin/fortmemory reindex --config ~/Vaults/Personal/.fortmemory/config.toml
./bin/fortmemory serve   --config ~/Vaults/Personal/.fortmemory/config.toml
```

| Surface | URL |
|---------|-----|
| Dashboard | http://127.0.0.1:7432/ |
| Health | http://127.0.0.1:7432/v1/health |

Paste the `fm_…` token in the dashboard for search & activity.

### Example API calls

```bash
curl -sS http://127.0.0.1:7432/v1/health | jq .

curl -sS -H "Authorization: Bearer fm_…" \
  "http://127.0.0.1:7432/v1/read?path=Scratch/hello.md" | jq .

curl -sS -H "Authorization: Bearer fm_…" \
  -H "Content-Type: application/json" \
  -d '{"q":"preferences","topK":5}' \
  http://127.0.0.1:7432/v1/search | jq .
```

### CLI governed write

```bash
./bin/fortmemory write \
  --config ~/Vaults/Personal/.fortmemory/config.toml \
  --key ~/path/to/agent-key.json \
  --path Scratch/hello.md \
  --body $'# Hello from FortMemory\n\nGoverned write.\n'
```

On allow, the note is written and annotated with `last_signal_id` (post-allow metadata; the FortSignal hash covers the agent payload).

---

## MCP

```bash
./bin/fortmemory mcp \
  --config ~/Vaults/Personal/.fortmemory/config.toml \
  --agent research-01 \
  --key ~/path/to/agent-key.json
```

| Tool | Description |
|------|-------------|
| `memory_search` | Full-text (and hybrid) search |
| `memory_read` | Read a vault path |
| `memory_write` | FortSignal-gated write |
| `memory_delete` | FortSignal-gated delete |

Example client config: [`examples/mcp.cursor.json`](./examples/mcp.cursor.json) · Guide: [`docs/MCP.md`](./docs/MCP.md)

---

## Configuration highlights

`.fortmemory/config.toml` (created by `init`):

```toml
vault_id = "personal"
vault_path = "/home/you/Vaults/Personal"
bind = "127.0.0.1"
port = 7432

[fortsignal]
api_base = "https://api.fortsignal.com"
api_key_env = "FORTSIGNAL_API_KEY"

[embeddings]
provider = "none"          # or "ollama"
ollama_url = "http://127.0.0.1:11434"
model = "nomic-embed-text"

[policy]
deny_read = ["Private/**"]
deny_write = [".fortmemory/**", ".fortmemory/*"]

[security]
fail_closed_on_fortsignal = true
allow_ungated_reads = true
```

---

## Architecture

```
┌──────────────┐     ┌─────────────────────┐
│   Obsidian   │     │ FortSignal Dashboard│
│  (human UX)  │     │ Composer · Passports│
└──────┬───────┘     └──────────▲──────────┘
       │ files                  │ rules / verify
       ▼                        │
┌──────────────────────────────────────────┐
│           FortMemory (local)             │
│  HTTP · MCP · CLI · FTS · receipts       │
│  path jail · watcher · dashboard         │
└──────────────────▲───────────────────────┘
                   │ memory.* tools only
            ┌──────┴──────┐
            │   Agents    │
            └─────────────┘
```

| Doc | Contents |
|-----|----------|
| [docs/SYSTEM-ARCHITECTURE.md](./docs/SYSTEM-ARCHITECTURE.md) | Full FortSignal × FortMemory × Obsidian design |
| [docs/FORTSIGNAL-INTEGRATION.md](./docs/FORTSIGNAL-INTEGRATION.md) | Wire-level challenge/verify mapping |
| [docs/PRODUCT-SURFACE.md](./docs/PRODUCT-SURFACE.md) | Memory-first UI (Composer is Settings, not home) |
| [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) | Internal layers |

---

## HTTP API

| Method | Path | Auth | Notes |
|--------|------|------|--------|
| `GET` | `/v1/health` | — | Liveness, index stats |
| `GET` | `/v1/read?path=` | Bearer | Read Markdown |
| `POST` | `/v1/search` | Bearer | `{ "q", "topK?", "pathPrefix?" }` |
| `POST` | `/v1/write` | Bearer + FortSignal | Governed write |
| `POST` | `/v1/delete` | Bearer + FortSignal | Governed delete |
| `GET` | `/v1/receipts` | Bearer | Local audit log |
| `GET` | `/v1/agents` | Bearer | Registered local agents |
| `GET` | `/` | — | Operator dashboard |

OpenAPI sketch: [`docs/openapi.yaml`](./docs/openapi.yaml)

**Auth reminder**

| Credential | Purpose |
|------------|---------|
| `FORTSIGNAL_API_KEY` (`fs_live_…`) | Tenant → FortSignal API (server process only) |
| Agent Ed25519 key file | Signs challenges (never uploaded to FortSignal private key) |
| `fm_…` bearer | Auth to FortMemory HTTP API |

---

## Remote access

| Situation | Command / approach |
|-----------|-------------------|
| Same machine | `http://127.0.0.1:7432` (**default**) |
| Tailscale tailnet | `fortmemory tailscale print-serve` |
| Hostname / Access / mTLS | `fortmemory cloudflare …` |

```bash
fortmemory tailscale
fortmemory tailscale print-serve
# → tailscale serve --bg http://127.0.0.1:7432

fortmemory cloudflare install
fortmemory cloudflare check
fortmemory cloudflare quick    # temporary trycloud URL
```

Keep `bind = "127.0.0.1"`. Do **not** expose the API raw on the public internet.  
Full guide: [docs/REMOTE-ACCESS.md](./docs/REMOTE-ACCESS.md)

---

## Project layout

```text
fortmemory-vault/
├── cmd/fortmemory/           # CLI entrypoint
├── internal/
│   ├── server/               # HTTP API + embedded dashboard
│   ├── vault/                # Path jail + Markdown I/O
│   ├── index/                # SQLite FTS5 (+ optional vectors)
│   ├── fortsignal/           # Enforcement client
│   ├── memory/               # Write/delete orchestration
│   ├── mcpserver/            # MCP stdio tools
│   ├── agent/                # Local tokens + Ed25519 signer
│   ├── watcher/              # fsnotify reindex
│   └── tunnel/               # Cloudflare + Tailscale helpers
├── web/                      # Dashboard source
├── docs/                     # Architecture & product docs
├── examples/                 # MCP client samples
├── LICENSE                   # Apache-2.0
├── NOTICE
└── SECURITY.md
```

---

## Documentation

| Document | Description |
|----------|-------------|
| [docs/INDEX.md](./docs/INDEX.md) | Full documentation index |
| [docs/MVP.md](./docs/MVP.md) | Scope & non-goals |
| [docs/IMPLEMENTATION.md](./docs/IMPLEMENTATION.md) | Build status |
| [docs/MCP.md](./docs/MCP.md) | MCP setup |
| [docs/SECURITY.md](./docs/SECURITY.md) | Threat model |
| [docs/CLI.md](./docs/CLI.md) | CLI reference |
| [CONTRIBUTING.md](./CONTRIBUTING.md) | Contribution guide |

---

## Security

- **Default bind:** `127.0.0.1` only  
- **Fail closed** on mutates if FortSignal is unreachable  
- **Path jail** — no escape via `..` / absolute paths; `.fortmemory/` reserved  
- **Private keys** stay on your machine  
- **Report vulnerabilities:** [hr@fortsignal.com](mailto:hr@fortsignal.com) — subject `FortMemory security`  

See [SECURITY.md](./SECURITY.md) and [docs/SECURITY.md](./docs/SECURITY.md).

---

## Development

```bash
make test      # go test ./...
make build     # bin/fortmemory
make install   # go install ./cmd/fortmemory
make sync-ui   # copy web/index.html → embed path
```

---

## Status

**MVP** — local server usable for dogfood: init, serve, search, FortSignal-gated write/delete, MCP, dashboard, vault watcher.

Roadmap themes: semantic retrieval polish, Obsidian companion plugin, FortVault sync (commercial).

---

## Suggested GitHub topics

```
agent-memory  local-first  obsidian  markdown  rag  mcp
fortsignal  ed25519  webauthn  go  sqlite  open-core
ai-agents  privacy  audit-log  passkeys
```

*(Repo → About → Topics)*

---

## License

Copyright © 2026 Jeffrey Walters / FortSignal.

Licensed under the **Apache License, Version 2.0** — see [LICENSE](./LICENSE) and [NOTICE](./NOTICE).

```
SPDX-License-Identifier: Apache-2.0
```

**Trademarks:** FortMemory, FortVault, and FortSignal are trademarks of FortSignal / Jeffrey Walters.  
The Apache License does **not** grant trademark rights.

---

## Links

| Resource | URL |
|----------|-----|
| Website | [fortsignal.com](https://fortsignal.com) |
| FortSignal docs | [fortsignal.com/docs](https://fortsignal.com/docs) |
| Policy Composer | [fortsignal.com/composer](https://fortsignal.com/composer) |
| This repository | [github.com/fortsignal/fortmemory-vault](https://github.com/fortsignal/fortmemory-vault) |

---

<p align="center">
  <sub>Built for agent builders who want memory that is human-readable, local-first, and accountable.</sub>
</p>
