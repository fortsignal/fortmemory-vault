# Peer Memory Sharing (Post-MVP)

**Status:** Design only — Phase 3. Not required for MVP.

## Goal

Allow agent on Node A to request scoped memory from Node B under policy, with FortSignal receipts and provenance.

## Non-goals

- Public global discovery marketplace  
- Multi-master shared folder without authority  
- Blind full-vault replication  

## Trust model

Peers are **invite-based semi-trusted** (Cloudflare Tunnel peer, Tailscale teammate, or second machine you control). Each share is a governed action.

## Flow

```
Agent A → POST share_request → Node B FortMemory
                                → local peer policy
                                → FortSignal memory.share_* / memory.read
                                → optional human step-up
                                → scoped results + grant signalId
Agent A → store under Inbox/from-B/ with provenance frontmatter
```

## Example request

```json
{
  "fromVaultId": "vlt_alice",
  "fromAgentId": "ops-agent-01",
  "purpose": "Investigate checkout latency",
  "query": "stripe webhook timeout runbook",
  "scopes": ["search", "read"],
  "maxResults": 5,
  "maxBytes": 50000
}
```

## Provenance frontmatter (importer)

```yaml
---
type: episode
provenance:
  remoteVault: vlt_bob
  remotePath: Runbooks/stripe-webhooks.md
  grantSignalId: 9a3f2c11-…
  retrievedAt: 2026-07-16T18:02:00Z
---
```

## Vault profile (opt-in)

```json
{
  "vaultId": "vlt_bob_ops",
  "displayName": "Bob Ops Knowledge",
  "description": "Runbooks and postmortems (no secrets).",
  "tags": ["sre", "runbooks"],
  "accessLevel": "invite_only",
  "endpoints": ["http://bob-laptop.ts.net:7432"],
  "capabilities": ["search", "read"],
  "policySummary": "Internal tags only · no writes · 50KB cap"
}
```

## Discovery MVP for peers

1. Paste peer tunnel URL (Cloudflare Tunnel primary; Tailscale hostname OK)  
2. Exchange profile over authenticated channel (mTLS preferred)  
3. Human approves trust once (passkey)  

No DHT in v1 of peer feature.

## Remote access defaults

| Method | Recommendation |
|--------|----------------|
| Cloudflare Tunnel + mTLS | **Primary** |
| Tailscale | Supported |
| Public bind | Discouraged |

See [SECURITY.md](./SECURITY.md).
