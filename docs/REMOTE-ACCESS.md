# Remote Access Options

FortMemory is **local-first**. The server should stay on **loopback** (`127.0.0.1`).  
Remote access is optional and uses **tunnels or mesh** — never bind the API raw to the public internet.

## Choose a path

| Situation | Use |
|-----------|-----|
| Agent on **same machine** as vault | **Nothing** — `http://127.0.0.1:7432` |
| Other devices on **your Tailscale tailnet** | **Tailscale** (supported) |
| Hostname, demos, hybrid/public edge, Access/mTLS | **Cloudflare Tunnel** (**primary** for that class) |
| Multi-device encrypted vault sync | **FortVault** (commercial, later) — not a tunnel |

```
                    ┌─────────────────────┐
                    │ fortmemory serve    │
                    │ 127.0.0.1:7432      │
                    └──────────▲──────────┘
                               │
          ┌────────────────────┼────────────────────┐
          │                    │                    │
   same host              Tailscale            Cloudflare
   (default)            Serve / mesh         Tunnel+Access
```

## Security rules (all remotes)

1. Keep `bind = "127.0.0.1"` in config (default).  
2. Require FortMemory **bearer tokens** for API calls.  
3. Mutates still need **FortSignal** allow + `signalId`.  
4. Prefer identity in front of the tunnel (Tailscale ACLs / Cloudflare Access or mTLS).  
5. Never commit API keys or agent private keys.

---

## Option A — Local only (default)

```bash
fortmemory serve --config …/.fortmemory/config.toml
# Agents / MCP / dashboard → http://127.0.0.1:7432
```

No domain. No Cloudflare. No Tailscale.

---

## Option B — Tailscale (supported)

Best when your agents and machines already share a **tailnet**.

### Prerequisites

- [Tailscale](https://tailscale.com/download) installed and logged in on the FortMemory host  
- `fortmemory serve` running on loopback  

### Expose on the tailnet

```bash
# Print host-specific guide (reads bind/port from config if available)
fortmemory tailscale
# or
fortmemory tunnel tailscale --print-guide

# Typical command (adjust port if needed):
tailscale serve --bg http://127.0.0.1:7432
```

### Call from another tailnet device

```bash
# MagicDNS name of the FortMemory host, e.g. desktop
curl -sS http://desktop:7432/v1/health
curl -sS -H "Authorization: Bearer fm_…" \
  -H 'Content-Type: application/json' \
  -d '{"q":"runbook","topK":5}' \
  http://desktop:7432/v1/search
```

### ACLs

Lock down who can reach the node/port in your Tailscale ACL policy so the whole tailnet is not automatically trusted.

### CLI helpers

```bash
fortmemory tailscale              # status + guide
fortmemory tailscale check        # is tailscale on PATH / status snippet
fortmemory tailscale print-serve  # exact serve command for your config port
```

---

## Option C — Cloudflare Tunnel (primary for hostname / Access)

Use when you want a stable hostname, Zero Trust Access, mTLS, or easy demos off-tailnet.

Full guide: [CLOUDFLARE-TUNNEL.md](./CLOUDFLARE-TUNNEL.md)

```bash
fortmemory cloudflare install
fortmemory cloudflare check
fortmemory cloudflare config --hostname memory.example.com
fortmemory cloudflare run --name fortmemory --cf-config ~/.cloudflared/config-fortmemory.yml
# dogfood temp URL:
fortmemory cloudflare quick
```

Put **Cloudflare Access** (or mTLS) in front before production agents.

---

## MCP over remote

MCP (`fortmemory mcp`) usually runs **on the vault host** (stdio to Cursor/Claude on that machine).

If the LLM host is remote:

- Prefer **HTTP API** through Tailscale/Cloudflare to `fortmemory serve`, or  
- Run MCP on the vault host and use a remote desktop / SSH workflow  

Don’t expose unauthenticated MCP on a public URL.

---

## What we intentionally do not build

- Auto-install of `tailscaled`  
- Embedding Tailscale as a Go library mesh  
- Replacing Tailscale ACLs or Cloudflare Access with our own IdP  
- Binding `0.0.0.0` as a documented happy path  

Tunnels stay **thin helpers** around industry tools.

---

## Related

- [CLOUDFLARE-TUNNEL.md](./CLOUDFLARE-TUNNEL.md)  
- [SECURITY.md](./SECURITY.md)  
- [PRODUCT-SURFACE.md](./PRODUCT-SURFACE.md)  
- [SYSTEM-ARCHITECTURE.md](./SYSTEM-ARCHITECTURE.md)  
