# HTTP API Overview

Base URL (default): `http://127.0.0.1:7432`

Auth (MVP): `Authorization: Bearer <local_agent_token>`

Content-Type: `application/json`

Mutating endpoints return FortSignal-shaped decisions.

## Endpoints

### `GET /v1/health`

Liveness + basic vault status.

```json
{
  "ok": true,
  "version": "0.1.0",
  "vaultId": "personal",
  "vaultPath": "/home/user/Vaults/Personal",
  "index": { "files": 120, "embedPending": 2 },
  "fortsignal": { "configured": true }
}
```

### `POST /v1/search`

```json
// request
{
  "q": "stripe webhook timeout",
  "topK": 8,
  "tags": ["runbook"],
  "pathPrefix": "Runbooks/"
}

// response
{
  "results": [
    {
      "path": "Runbooks/stripe-webhooks.md",
      "score": 0.82,
      "excerpt": "If webhooks timeout…",
      "tags": ["runbook", "payments"],
      "sensitivity": "internal",
      "lastSignalId": "5564c849-…"
    }
  ]
}
```

### `GET /v1/read?path=Runbooks/stripe-webhooks.md`

```json
{
  "path": "Runbooks/stripe-webhooks.md",
  "content": "# Stripe webhooks\n…",
  "contentHash": "sha256:…",
  "lastSignalId": "5564c849-…"
}
```

### `POST /v1/write`

```json
// request
{
  "path": "Scratch/agent-note.md",
  "content": "# Note\n\nBody…\n",
  "mode": "overwrite"
}

// allow
{
  "decision": "allow",
  "signalId": "9a3f2c11-…",
  "path": "Scratch/agent-note.md",
  "contentHash": "sha256:…",
  "verifiedBy": "agent",
  "verifiedAt": "2026-07-16T18:00:00Z"
}

// deny
{
  "decision": "deny",
  "reason": "path_not_allowed"
}
```

Modes: `create` | `append` | `overwrite`

### `POST /v1/delete`

```json
{
  "path": "Scratch/old.md"
}
```

Same decision envelope as write.

### `GET /v1/receipts?limit=50`

Local receipt log (newest first). Optional filters: `action`, `decision`, `from`, `to`, `pathPrefix`.

## Error model

| HTTP | When |
|------|------|
| 200 | Decision returned (including `decision: deny`) for mutate |
| 400 | Malformed JSON, missing fields |
| 401 | Bad/missing local token |
| 404 | Path not found (read) |
| 503 | FortSignal unavailable on mutate |

**Note:** Prefer `200 + decision: deny` for policy/crypto denials so agents can branch on `reason` without treating governance as transport failure. Use 503 only for infrastructure failure.

## Versioning

Prefix `/v1`. Breaking changes bump to `/v2`.

## CORS

Disabled or loopback-only for MVP. Agents should call from same machine or tunnel, not random websites.

## Idempotency (future)

Optional `Idempotency-Key` header for writes — not required in MVP.

Full machine-readable contract: [openapi.yaml](./openapi.yaml).
