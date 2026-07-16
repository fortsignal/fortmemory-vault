# FortSignal API Alignment (from fortsignal-api + SDK)

This document is grounded in the **live FortSignal implementation** under `~/projects/fortsignal-api` and types in `~/projects/fortsignal-sdk` (not marketing copy alone).

**Base URL:** `https://api.fortsignal.com` (override for self-host)  
**Auth:** `Authorization: Bearer fs_live_…` (or `fs_demo_key` sandbox)  
**Go client:** FortMemory implements HTTP itself — no dependency on `@fortsignal/sdk`.

---

## Boundary (do not blur)

| FortSignal owns | FortMemory owns |
|-----------------|-----------------|
| Signature verify (WebAuthn / Ed25519) | Markdown vault + SQLite index |
| Parameter binding + nonce TTL | Path jail + write planner |
| Policy + delegation checks | Local path/tag policy (extra) |
| `signalId` issuance (90d cloud retention) | Local receipt mirror (durable) |
| Allow/deny decision | Execute file mutation **only on allow** |

FortSignal is **stateless enforcement**. Cumulative counters (writes/day) live in FortMemory if needed.

---

## Endpoints FortMemory uses

| Method | Path | Who | MVP |
|--------|------|-----|-----|
| `POST` | `/challenge/start` | agent or human | **Yes** (agent) |
| `POST` | `/challenge/verify` | agent or human | **Yes** (agent) |
| `POST` | `/agent/register` | setup | Optional helper |
| `GET` | `/agent/list` | setup/UI | Optional |
| `GET` | `/signal/:signalId` | audit | Optional |
| `GET` | `/audit` | audit export | Optional |
| `POST` | `/register/*` | human passkey | Phase 2 UI |
| Dashboard | delegate / revoke | human owner | **Manual** (not API key) |

**Critical:** Delegation issue/revoke requires **dashboard owner auth** (passkey session). A compromised FortMemory API key must not mint agent powers. Document setup as: register key → human approves delegation in dashboard.

---

## Validation limits (hard constraints)

From `fortsignal-api/src/lib/validate.ts`:

| Field | Rule |
|-------|------|
| `action` | string, **≤ 64** chars |
| `amount` | finite number (optional; default 0) |
| `recipient` | required string, **≤ 256** chars |
| `source` | optional string, **≤ 256** chars |
| `metadata` | optional **object**, JSON serialized **≤ 2048** chars |
| `userId` / `agentId` | string identity rules |

From `assertPolicy` (`policy.ts`):

- `allowedActions` compared **lowercased**  
- `maxAmountPerAction` / per-action amount limits  
- `allowedRecipients` supports wildcards:  
  - `prefix/*` → exact prefix without `/*` **or** startswith `prefix/`  
  - `prefix*` → startswith  
  - `*suffix` → endswith  
  - exact match  
- `requiredMetadata` is **exact key → value** string equality (not “key present only”)  
- `requireBiometric` applies to human flow (agents use delegation-time biometric)

---

## Agent challenge flow (exact shapes)

### Start — `POST /challenge/start`

```json
{
  "agentId": "research-01",
  "action": "memory.write",
  "amount": 1842,
  "recipient": "personal/Scratch/note.md",
  "source": "research-01",
  "metadata": {
    "vaultId": "personal",
    "contentHash": "sha256:…",
    "mode": "overwrite"
  }
}
```

**Success (agent):**

```json
{
  "challenge": "<base64url of SHA-256 intent hash bytes>",
  "agentId": "research-01",
  "delegationId": "del_…",
  "expiresIn": 60
}
```

**Policy/delegation deny at start (agent fast-fail):** HTTP **403** with body:

```json
{ "decision": "deny", "reason": "action_not_allowed" }
```

Reasons include: `delegation_invalid`, `action_not_allowed`, `amount_exceeds_policy`, `recipient_not_allowed`, etc.

Go client **must** handle both 200 challenge and 403 deny without treating all non-200 as transport failure.

### Sign

Sign the **raw challenge bytes** (base64url-decode `challenge`) with Ed25519 private key.  
Submit signature as **base64url**.

### Verify — `POST /challenge/verify`

```json
{
  "agentId": "research-01",
  "challenge": "<same base64url challenge>",
  "signature": "<base64url Ed25519 signature>"
}
```

**Allow:**

```json
{
  "decision": "allow",
  "signalId": "uuid",
  "verifiedBy": "agent",
  "verifiedAt": "ISO-8601",
  "agentId": "research-01",
  "delegationId": "del_…",
  "policyId": "pol_…",
  "action": "memory.write",
  "amount": 1842,
  "recipient": "personal/Scratch/note.md",
  "source": "research-01",
  "metadata": { … }
}
```

**Deny:** `{ "decision": "deny", "reason": "…" }`

Intent hash formula (server-side, for understanding only):

```
SHA-256( nonce + ":" + action + ":" + amount + ":" + recipient + ":" + source + ":" + metadataJSON )
```

FortMemory never recomputes the nonce; it trusts FortSignal verify.

Challenge TTL: default **60s** (account may set 5–300s). Mutate path must complete start→sign→verify quickly.

---

## Memory action mapping

| FortMemory op | `action` | `amount` | `recipient` |
|---------------|----------|----------|-------------|
| write | `memory.write` | byte length of body | path binding (below) |
| delete | `memory.delete` | 0 | path binding |
| read (optional gate) | `memory.read` | 0 or maxChars | path binding |
| search (optional) | `memory.search` | topK | vault id or `vault/{id}` |

Actions must appear in FortSignal policy `allowedActions` (lowercase match).

---

## Recipient encoding (≤256) + policy wildcards

**Problem:** Full vault paths can exceed 256 characters.  
**Also:** Policies use recipient allowlists with `/*` wildcards.

### Canonical MVP scheme

```
recipient = "{vaultId}/{relative/path.md}"
```

Examples:

- `personal/Scratch/note.md`  
- `personal/Inbox/research/2026-07-16.md`  

Policy examples:

```
allowedRecipients: [
  "personal/Scratch/*",
  "personal/Inbox/*"
]
```

Matches `assertPolicy` `prefix/*` rule.

### Long path fallback

If `len(vaultId)+1+len(path) > 256`:

```
recipient = "{vaultId}/#/{sha256hex16}"
metadata.path = "{relative/path.md}"   // still path-jailed server-side
metadata.contentHash = "sha256:…"
```

Local policy **always** enforces the real path; FortSignal binds hash for integrity.

### Do not put full file bodies in metadata

Metadata budget is **2048** serialized characters. Use `contentHash` only.

Recommended metadata keys (keep small):

```json
{
  "vaultId": "personal",
  "contentHash": "sha256:e3b0c4…",
  "mode": "overwrite",
  "path": "Scratch/note.md"
}
```

If `path` already fully represented in `recipient`, omit `path` from metadata to save space.

---

## requiredMetadata caveat

FortSignal `requiredMetadata` checks **exact string equality** on values. Prefer:

- either no `requiredMetadata` for flexible agents  
- or fixed constants, e.g. `"vaultId": "personal"`  

Do **not** put dynamic `contentHash` in `requiredMetadata` (would pin one hash forever).

---

## Signing modes in FortMemory

| Mode | Who holds Ed25519 private key | MVP |
|------|-------------------------------|-----|
| **A. Local signer** | FortMemory loads key file for co-located agent | **Yes (dogfood)** |
| **B. Sidecar sign** | External agent signs; FortMemory only start+verify with provided signature | Phase 2 |

Deep Agents key file format (from FortSignal docs):

```json
{ "agentId": "…", "privateKey": "<base64url>", "publicKey": "…" }
```

Support this path for interoperability with `fortsignal-deepagents`.

---

## Human / WebAuthn flow (Phase 2+)

Human `challenge/start` returns **WebAuthn options**, not a raw challenge string.  
Verify body is the WebAuthn assertion JSON from `@simplewebauthn/browser`.

For localhost dashboard WebAuthn, RPID/origin must match FortSignal account configuration — **non-trivial**. MVP stays **agent-delegation-first** to avoid RPID hell.

---

## Deny reasons FortMemory should pass through

From API / policy:

`verification_failed`, `parameters_tampered`, `invalid_challenge`, `challenge_expired`, `user_not_found`, `agent_not_found`, `delegation_invalid`, `policy_not_found`, `policy_expired`, `action_not_allowed`, `amount_exceeds_policy`, `recipient_not_allowed`, `source_not_allowed`, `metadata_mismatch`, `biometric_required`, `quota_exceeded`, `invalid_request`, `internal_error`

Local additions: `path_not_allowed`, `path_traversal`, `fortsignal_unavailable`, `sensitivity_forbidden`

---

## Sequence: memory.write

```
1. Auth local bearer → resolve agentId + key material
2. Path-jail + local policy
3. contentHash = sha256(body)
4. recipient = EncodeRecipient(vaultId, path)
5. POST /challenge/start { agentId, action:memory.write, amount:len(body), recipient, source:agentId, metadata }
6. If decision deny → return 200 {decision:deny, reason} (map 403)
7. Sign challenge bytes (mode A) or require client signature (mode B)
8. POST /challenge/verify { agentId, challenge, signature }
9. If allow → atomic write file → index → local receipt(signalId)
10. Return MutateResponse
```

Fail closed if FortSignal unreachable (`fail_closed_on_fortsignal = true`).

---

## Policy templates (Composer → allowedActions)

### Research agent

```
allowedActions: memory.read, memory.search, memory.write
maxAmountPerAction: 65536
allowedRecipients: personal/Scratch/*, personal/Inbox/*
```

NL:

> research-01 may read and search the vault notes under personal, and write only under personal/Scratch and personal/Inbox. Max 64KB per write. No deletes.

### Coder notes agent

```
allowedActions: memory.read, memory.write
allowedRecipients: personal/Projects/*
maxAmountPerAction: 32768
```

### Read-only auditor

```
allowedActions: memory.read, memory.search
allowedRecipients: []   # empty = no recipient restriction at FS layer? 
# Prefer explicit prefixes or rely on FortMemory local deny for Private/**
```

**Note:** Empty `allowedRecipients` means no recipient constraint in FortSignal — **FortMemory local policy must still deny `Private/**`**.

---

## SDK parity (TypeScript reference)

`@fortsignal/sdk` agent methods FortMemory Go client mirrors:

```ts
client.agent.register({ agentId, publicKey })
client.agent.startChallenge({ agentId, action, amount?, recipient, source?, metadata?, delegationId? })
client.agent.verify({ agentId, challenge, signature })
```

Human methods are out of MVP daemon path.

---

## Integration test plan

1. Register agent + dashboard delegation with memory policy  
2. Allow write under Scratch → file exists + signalId  
3. Deny write under Private → no file  
4. amount over cap → `amount_exceeds_policy`  
5. Wrong path prefix → `recipient_not_allowed`  
6. Mutate body after start (test only) → `parameters_tampered`  
7. Expired/revoked delegation → `delegation_invalid`  
8. Unset API key → 503 `fortsignal_unavailable`  

Use `fs_demo_key` only where sandbox credentials exist; otherwise live key on dedicated test tenant.

---

## Gaps / product risks to track

| Gap | Mitigation |
|-----|------------|
| Recipient 256 limit | Encoding + hash fallback |
| Metadata 2048 limit | Hash only, no body |
| requiredMetadata exact match | Don’t put dynamic hashes in policy requiredMetadata |
| WebAuthn RPID for localhost UI | Defer human step-up; agent delegation first |
| Delegation not API-key creatable | Setup docs + dashboard deep link |
| Cloud signal retention 90d | Local receipts are source of long-term truth |

---

## Source references

- `fortsignal-api/src/routes/challenge.ts` — agent start/verify, intent hash  
- `fortsignal-api/src/lib/validate.ts` — field limits  
- `fortsignal-api/src/lib/policy.ts` — recipient wildcards, requiredMetadata  
- `fortsignal-sdk/src/types.ts` — TS request/response shapes  
- Public docs: https://fortsignal.com/docs  
