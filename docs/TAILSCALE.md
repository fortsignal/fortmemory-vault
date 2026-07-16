# Tailscale (Supported Remote Access)

**Role:** Easy mesh access for machines already on your tailnet.  
**Priority:** Supported alternative — not required for local MVP.  
**Primary hostname/Access path:** [CLOUDFLARE-TUNNEL.md](./CLOUDFLARE-TUNNEL.md)  
**Overview:** [REMOTE-ACCESS.md](./REMOTE-ACCESS.md)

## When to use Tailscale

- Agent or laptop B is on the **same Tailscale tailnet** as the vault host  
- You want private mesh connectivity without managing DNS/public certs  
- You’re fine with Tailscale identity + ACLs as the network gate  

When to prefer Cloudflare instead: public/demo hostname, Zero Trust org policies, mTLS to arbitrary clients, non-Tailscale users.

## Setup

1. Install Tailscale and `tailscale up` on the FortMemory host.  
2. Run FortMemory on loopback (default):

   ```bash
   fortmemory serve --config ~/Vaults/Personal/.fortmemory/config.toml
   ```

3. Serve it on the tailnet:

   ```bash
   fortmemory tailscale print-serve
   # example output → run:
   tailscale serve --bg http://127.0.0.1:7432
   ```

4. From another node:

   ```bash
   curl -sS http://<magicdns-or-tailscale-ip>:7432/v1/health
   ```

Use the FortMemory **bearer token** and FortSignal for writes as usual.

## CLI

```bash
fortmemory tailscale                 # check + guide
fortmemory tailscale check
fortmemory tailscale print-serve     # exact command for config port
fortmemory tunnel tailscale …        # alias
```

## Security notes

- FortMemory still authenticates agents with `fm_…` tokens.  
- Writes still require FortSignal.  
- Tighten Tailscale **ACLs** so only trusted nodes can reach the service.  
- Tailscale does not replace FortSignal policy — it only replaces “how packets arrive.”

## Non-goals

- Bundling or auto-updating Tailscale  
- Custom ACL editor in FortMemory  
- Treating Tailscale as the only remote story  
