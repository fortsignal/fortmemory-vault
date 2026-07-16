# FortSignal × FortMemory Integration Guide

**Status of integration code:** MVP path is **implemented**  
Use `fortmemory doctor` to verify your environment.

---

## Are we done integrating?

| Capability | Status |
|------------|--------|
| FortSignal `challenge/start` + `verify` (agent) | **Done** |
| Parameter bind: `memory.write` / `memory.delete`, path, contentHash, bytes | **Done** |
| Local agent tokens + Ed25519 key files | **Done** |
| HTTP API write/delete gated | **Done** |
| CLI write/delete gated | **Done** |
| MCP tools gated | **Done** |
| Local receipts + frontmatter `last_signal_id` | **Done** |
| `GET /agent/list` health (doctor) | **Done** |
| `fortmemory doctor` | **Done** |
| Human WebAuthn step-up from FM UI | Not MVP |
| FortSignal `/agent/register` from CLI | Optional (dashboard path preferred) |
| Signal Views / peer share | Later |

**You are done for core integration.** Remaining work is ops polish, dogfood, and product features—not a new architecture.

---

## End-to-end setup (operator)

### 1. FortMemory vault

```bash
fortmemory init ~/Vaults/Personal --id personal
export FORTSIGNAL_API_KEY=fs_live_...
```

`vault_id` (`personal`) becomes the **prefix** of FortSignal `recipient` fields.

### 2. FortSignal — agent passport

1. Open [dashboard](https://fortsignal.com/dashboard) → **Agent Passports**  
2. Create agent → download `agent-key.json` (private key never uploaded)  
3. Or API: `POST /agent/register` with public key only  

### 3. FortSignal — policy (NL Composer or form)

Open [Composer](https://fortsignal.com/composer). Example NL:

> Research agents may write Markdown under Scratch and Inbox only. Max 64KB per write. No deletes of Private. Allow memory.write and memory.search and memory.read. Deny everything under Private.

**Compiled constraints you need (conceptually):**

| Field | Example |
|-------|---------|
| `allowedActions` | `memory.write`, `memory.delete` (if used), optionally read/search if you gate them later |
| `maxAmountPerAction` | `65536` (bytes) |
| `allowedRecipients` | `personal/Scratch/*`, `personal/Inbox/*` |

Recipient scheme FortMemory uses:

```text
{vaultId}/{relative/path.md}
→ personal/Scratch/hello.md
```

### 4. FortSignal — delegation

In dashboard: attach policy to agent → **passkey approve** → note expiry.

Without an active delegation, `challenge/start` returns **deny** (`delegation_invalid` / policy reasons).

### 5. FortMemory — local agent

```bash
fortmemory agent add research-01 \
  --config ~/Vaults/Personal/.fortmemory/config.toml \
  --key ~/Downloads/agent-key.json
# save fm_… token once
```

### 6. Verify integration

```bash
# API key + list agents + challenge/verify (no file write)
fortmemory doctor --key ~/Downloads/agent-key.json

# Full write probe → Scratch/_doctor_probe.md
fortmemory doctor --key ~/Downloads/agent-key.json --write-probe
```

### 7. Run

```bash
fortmemory reindex --config …/.fortmemory/config.toml
fortmemory serve   --config …/.fortmemory/config.toml
# http://127.0.0.1:7432/
```

---

## Policy templates (copy into Composer)

### Research (default)

```
Agent may use memory.write under personal/Scratch/* and personal/Inbox/* only.
Max 65536 bytes per write. No access to personal/Private/*.
Allow memory.read and memory.search if needed for tools.
```

### Read-only auditor

```
Agent may only memory.read and memory.search.
No memory.write or memory.delete.
```

### Coder notes

```
Agent may memory.write only to personal/Projects/*/AGENT_NOTES.md.
Max 32768 bytes. No delete. No Private.
```

### After save

Dashboard → assign policy to agent → passkey **delegation**.

---

## What FortMemory sends on write

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

Then: sign challenge → verify → on allow, write file (+ annotate `last_signal_id`).

---

## Integration map (code)

| FortSignal | FortMemory |
|------------|------------|
| `POST /challenge/start` | `internal/fortsignal` + `memory.Write` / `Delete` |
| `POST /challenge/verify` | same |
| `GET /agent/list` | `fortmemory doctor` |
| `POST /agent/register` | Dashboard preferred; client method exists |
| Composer / dashboard | Settings deep links only |
| Delegation | Required; never created by FortMemory |

---

## Troubleshooting

| Symptom | Likely cause |
|---------|----------------|
| `missing FORTSIGNAL_API_KEY` | Export key / config `api_key_env` |
| `agent_not_found` | Register passport / wrong agentId |
| `delegation_invalid` | No or expired delegation |
| `action_not_allowed` | Policy missing `memory.write` |
| `recipient_not_allowed` | Path not under allowed prefix (check vault_id) |
| `amount_exceeds_policy` | Body larger than maxAmountPerAction |
| Doctor PASS but HTTP 401 | Wrong/missing `fm_…` bearer on API |

---

## Not required for integration

- Domain or Cloudflare  
- Tailscale  
- Obsidian plugin  
- FortVault  

Local loopback is enough.
