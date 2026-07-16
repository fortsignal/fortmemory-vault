# Security Model

## Mission

Treat agent memory as a **high-stakes action surface**, not a convenience cache. FortMemory fails closed on unauthorized mutation.

## Assets

| Asset | Sensitivity |
|-------|-------------|
| Vault Markdown content | User secrets, PII, business knowledge |
| Embeddings / index | Derivative of content |
| Local receipts | Audit integrity |
| Agent API tokens | Local auth |
| FortSignal API key | Tenant credential (server only) |
| Agent Ed25519 private keys | Agent identity (never sent to FortSignal or peers) |
| Tunnel endpoints | Remote attack surface |

## Adversaries (MVP threat model)

1. **Rogue or prompt-injected agent** trying to overwrite Policies or exfil Private notes  
2. **Compromised agent process** with a valid local API token  
3. **Local malware** with user FS permissions (partially out of scope — OS boundary)  
4. **Malicious peer** (post-MVP) requesting over-broad share  
5. **Network attacker** if port exposed beyond loopback without tunnel auth  

## Non-goals (honest)

- Protecting vault content from an attacker who already has the user’s OS user account and disk encryption keys  
- Preventing the user from hand-editing Markdown (by design)  
- Stopping a stolen FortSignal API key from calling FortSignal (rotate keys; FortSignal side)  
- Perfect LLM alignment  

## Controls

### 1. Local-first default

- Bind `127.0.0.1` only by default  
- Explicit config to bind LAN / tunnel  

### 2. Path jail

- All ops resolve under vault root  
- Reject `..`, symlink escape (document policy: do not follow symlinks out of vault)  

### 3. FortSignal on mutate

- Parameter-bound content hash  
- Agent passport + delegation for autonomous agents  
- Human passkey step-up for sensitive classes when policy requires  

### 4. Dual policy

- FortSignal: action, caps, recipients, expiry, biometric  
- Local: path globs, tags, sensitivity  

### 5. Least privilege agents

- Separate agentIds and local tokens per agent  
- Scoped allowed path prefixes in local policy  

### 6. Receipt integrity

- Store `signalId`, params, path, decision, timestamp  
- Writes without signalId are either human ungoverned edits or bugs  

### 7. Secrets handling

| Secret | Location |
|--------|----------|
| FortSignal API key | env or OS keychain / config with 0600 |
| Agent Ed25519 private key | agent host secrets — **not** in vault git |
| Local API tokens | `.fortmemory/` with restrictive perms |

Never log full note bodies or private keys at info level.

## Human vs agent vs human-edit

| Actor | Mutate path |
|-------|-------------|
| Agent via API | FortSignal required |
| Human via UI (future) | Passkey challenge when policy says |
| Human via Obsidian | Always possible; no signalId (label as `verifiedBy: human_external`) |

External human edits are a feature. Index them; do not pretend they were governed.

## Sensitive paths (recommended defaults)

```
Private/**          → no agent read/write by default
Policies/**         → human step-up for overwrite
.fortmemory/**      → engine only
**/*secret*         → optional deny patterns
```

## Remote access guidance

| Method | Guidance |
|--------|----------|
| Cloudflare Tunnel + mTLS | **Primary** remote path; do not expose unauthenticated API |
| Tailscale | Supported alternative |
| 0.0.0.0 on public IP | Forbidden in docs; refuse or loud warn |

## FortSignal dependency

MVP: **fail closed** if FortSignal is unreachable for mutates.

Future option: offline signed intent queue — only with clear product labeling and replay controls.

## Abuse cases & responses

| Abuse | Response |
|-------|----------|
| Agent overwrites AGENTS.md policy notes | Path policy deny + FortSignal deny |
| Agent dumps vault via search | Rate limits + tag/path policy + optional gated search |
| Replay old challenge | FortSignal nonce/TTL |
| Body swap after start | `parameters_tampered` |
| Peer requests `restricted` | Share policy never allow |

## Compliance narrative (enterprise)

- Memory mutations are **authorized actions** with stable reason codes on deny  
- `signalId` correlates to change in vault (content hash)  
- Exportable local audit for incident review  
- Aligns with FortSignal VI-oriented “proof of what was authorized,” applied to knowledge ops  

## Security review checklist (pre-public)

- [ ] Path traversal tests  
- [ ] Symlink tests  
- [ ] Token brute-force resistance (local rate limit)  
- [ ] No secrets in receipts export by default  
- [ ] Default bind address review  
- [ ] Dependency audit (`go.gov` / `govulncheck`)  
