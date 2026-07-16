# Cloudflare Tunnel Plugin (Primary Remote Access)

FortMemory ships a first-class **Cloudflare Tunnel plugin** (open source).  
It wraps [`cloudflared`](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/) — not a browser extension.

**Preference:** Cloudflare Tunnel + mTLS/Access **primary**; Tailscale supported.

## Commands

```bash
fortmemory cloudflare install      # download cloudflared → ~/.local/bin
fortmemory cloudflare check        # detect binary + local target URL
fortmemory cloudflare config       # write named-tunnel YAML
fortmemory cloudflare quick        # temporary trycloud URL (dogfood)
fortmemory cloudflare run          # run named tunnel
fortmemory cloudflare help
```

Alias: `fortmemory tunnel cloudflare …`

## Install cloudflared

```bash
fortmemory cloudflare install
# ensure ~/.local/bin is on PATH
export PATH="$HOME/.local/bin:$PATH"
fortmemory cloudflare check
```

Or system package / brew — see `fortmemory cloudflare print-install`.

## Dogfood: quick tunnel (no DNS)

Keeps FortMemory on loopback; Cloudflare gives a temporary `*.trycloudflare.com` URL.

```bash
# terminal 1 — when serve exists:
# fortmemory serve

# terminal 2
fortmemory cloudflare quick --url http://127.0.0.1:7432
```

**Not for production** — no stable hostname, limited Access controls. Use named tunnels + Zero Trust for real agents.

## Production: named tunnel + Access / mTLS

### 1. Auth + create tunnel

```bash
cloudflared tunnel login
cloudflared tunnel create fortmemory
# note Tunnel UUID
```

### 2. Generate config

```bash
fortmemory cloudflare config \
  --hostname memory.example.com \
  --url http://127.0.0.1:7432 \
  --name fortmemory \
  --tunnel-id <TUNNEL_UUID> \
  --write-config ~/.cloudflared/config-fortmemory.yml
```

### 3. DNS

```bash
cloudflared tunnel route dns fortmemory memory.example.com
```

### 4. Run

```bash
fortmemory cloudflare run \
  --name fortmemory \
  --cf-config ~/.cloudflared/config-fortmemory.yml
```

### 5. Lock down (required for agents)

Do **not** expose an open memory API on the public Internet.

1. Cloudflare Zero Trust → Access → Application on `memory.example.com`  
2. Policy: **Service tokens** and/or **mTLS client certificates**  
3. FortMemory **bearer tokens** still required  
4. FortSignal still gates **writes**  

Docs: https://developers.cloudflare.com/cloudflare-one/

## Security checklist

- [ ] FortMemory bind = `127.0.0.1` only  
- [ ] Ingress only via tunnel  
- [ ] Access policy or mTLS enabled  
- [ ] Agent tokens rotated  
- [ ] FortSignal path-scoped delegation  
- [ ] FortSignal API key not on untrusted remote agents  

## Architecture

```
Remote agent
    │  HTTPS (+ Access/mTLS)
    ▼
Cloudflare edge
    │  cloudflared (separate process)
    ▼
127.0.0.1:7432  fortmemory (loopback)
    │
    ▼
FortSignal (writes)
```

FortMemory does **not** embed your Cloudflare API token. `cloudflared` runs as its own process (cleaner blast radius).

## Tailscale (supported mesh)

See [TAILSCALE.md](./TAILSCALE.md) and [REMOTE-ACCESS.md](./REMOTE-ACCESS.md).

```bash
fortmemory tailscale print-serve
```

## Commercial note

Self-serve tunnel plugin = **OSS** (Apache-2.0).  
Managed org tunnels / team device inventory = **FortVault Team** — see [COMMERCIAL.md](./COMMERCIAL.md).
