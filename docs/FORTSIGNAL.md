# FortSignal Integration Overview

FortMemory is a **state plane**. FortSignal is the **enforcement plane**.

> **Authoritative, API-aligned detail** (field limits, exact JSON shapes, recipient encoding, gaps found in `fortsignal-api` source):  
> **[FORTSIGNAL-INTEGRATION.md](./FORTSIGNAL-INTEGRATION.md)**

Public reference: [FortSignal docs](https://fortsignal.com/docs) · [fortsignal.com](https://fortsignal.com)  
Local source: `~/projects/fortsignal-api`, `~/projects/fortsignal-sdk`

## Integration shape

```
FortMemory mutate handler
  → POST /challenge/start   (agentId or userId + bound params)
  → sign challenge          (Ed25519 agent key OR WebAuthn human)
  → POST /challenge/verify
  → on allow: apply vault mutation + store signalId
  → on deny: no mutation; return reason
```

Use FortSignal account API key (`Authorization: Bearer fs_live_…`) **only inside FortMemory process** (or a trusted local proxy). Never ship API keys to browser agents.

## Phased FortSignal adoption (aligned with their docs)

| Phase | FortSignal capability | FortMemory use |
|-------|----------------------|----------------|
| 1 | Human passkey challenge | UI step-up / owner actions |
| 2 | Policy profiles | Memory action constraints |
| 3 | Agent register + delegation | Autonomous memory agents |
| 4 | Enterprise audit / mTLS | Team & regulated |

MVP targets **Phase 3-shaped** agent writes: Ed25519 agent + passkey-approved delegation.

## Action names

Register and document these action strings in policies:

```
memory.read
memory.write
memory.delete
memory.search
memory.export
memory.share_request
memory.share_grant
memory.summarize
```

MVP enforces at least: `memory.write`, `memory.delete`.

## Parameter binding recipe

FortSignal challenge derivation (conceptual):

```
challenge = SHA-256(nonce + ":" + action + ":" + amount + ":" + recipient + ":" + source + ":" + metadata)
```

### memory.write example

```json
{
  "agentId": "research-01",
  "action": "memory.write",
  "amount": 1842,
  "recipient": "vault:personal/Scratch/note.md",
  "source": "research-01",
  "metadata": {
    "vaultId": "personal",
    "contentHash": "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    "mode": "overwrite",
    "bytes": 1842
  }
}
```

### Rules

1. **Hash the exact bytes** that will be written (UTF-8 Markdown body).  
2. Do not put full body in `metadata` (size limits; privacy).  
3. `recipient` must be stable and policy-matchable (prefix allowlists).  
4. On verify allow, write **only** those exact bytes.  
5. Store full FortSignal allow payload fields needed for audit.  

## Agent passport flow (autonomous memory agent)

1. Generate Ed25519 keypair on agent host (or FortSignal dashboard Agent Passports).  
2. `POST /agent/register` with public key only.  
3. Human approves **delegation** in FortSignal dashboard (passkey) with memory policy + expiry.  
4. FortMemory config references `agentId`; agent signs challenges with private key.  
5. Revoke in dashboard → next memory write denied (`delegation_invalid`).  

## Human step-up

For sensitive paths (e.g. `Policies/**`, export, peer grant):

- Use `userId` + WebAuthn via a small local UI or CLI ceremony  
- Or require biometric at delegation time and keep agent scope narrow  

MVP can ship agent-only mutates first; human UI ceremony Phase 2.

## Deny reasons (pass through)

Surface FortSignal reasons to API clients:

- `parameters_tampered`  
- `delegation_invalid`  
- `action_not_allowed`  
- `amount_exceeds_policy`  
- `recipient_not_allowed`  
- `policy_expired` / `policy_not_found`  
- `verification_failed`  
- `invalid_challenge`  

Add FortMemory-local reasons:

- `path_not_allowed`  
- `sensitivity_forbidden`  
- `path_traversal`  
- `vault_not_found`  
- `fortsignal_unavailable`  

## NL policy templates (Composer)

### Research agent

> research-01 may search and read all notes except Private/. It may write only under Inbox/research/ and Scratch/. No deletes. Delegation expires in 14 days.

### Coding agent

> coder-02 may read Projects/**/*.md and write only Projects/**/AGENT_NOTES.md. Overwrite requires biometric. Max write 32768 bytes.

### No export of confidential

> No agent may export or share notes tagged confidential or restricted. Biometric required for any memory.export.

### Peer runbooks (post-MVP)

> Peer vaults may search tags runbook and read matching files under 50KB. Writes from peers denied. Max 10 results per hour.

## Signal Views

When peer/team APIs exist, use Signal Views so partners get `audit` or `proof` scopes on receipts — not full metadata.

## Local receipt record (suggested schema)

```json
{
  "id": "local-uuid",
  "signalId": "5564c849-…",
  "decision": "allow",
  "reason": null,
  "action": "memory.write",
  "path": "Scratch/note.md",
  "contentHash": "sha256:…",
  "agentId": "research-01",
  "delegationId": "del_…",
  "policyId": "pol_…",
  "verifiedBy": "agent",
  "verifiedAt": "2026-07-16T18:00:00Z",
  "vaultId": "personal"
}
```

## Go client sketch

```go
// internal/fortsignal/client.go — conceptual
type Client interface {
    ChallengeStart(ctx context.Context, req ChallengeStartRequest) (*ChallengeStartResponse, error)
    ChallengeVerify(ctx context.Context, req ChallengeVerifyRequest) (*VerifyResult, error)
}
```

HTTP base: `https://api.fortsignal.com` (configurable for enterprise self-host).

## Testing against FortSignal

- Use sandbox/demo key if available (`fs_demo_key` per FortSignal docs) for non-prod  
- Integration tests should skip when `FORTSIGNAL_API_KEY` unset  
- Contract tests: allow path, deny path, tamper path  

## Boundary reminder (from FortSignal docs)

FortSignal answers allow/deny.  
FortMemory:

- stores files  
- maintains index  
- enforces path jail  
- mirrors receipts  
- implements memory-specific semantics  

Do not reimplement WebAuthn server-side crypto inside FortMemory beyond client flows.
