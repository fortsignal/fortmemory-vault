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
| API-key health via challenge ping (doctor) | **Done** |
| `GET /agent/list` (dashboard session only) | N/A for API keys |
| `fortmemory doctor` | **Done** |
| Human WebAuthn step-up from FM UI | Not MVP |
| FortSignal `/agent/register` from CLI | Optional (dashboard path preferred) |
| Signal Views / peer share | Later |

**You are done for core integration.** Remaining work is ops polish, dogfood, and product features—not a new architecture.

---

## End-to-end setup (any operator)

Everything below uses **your** values. Nothing is tied to a specific person or machine.

| Concept | What it is | You choose |
|---------|------------|------------|
| **Vault folder** | Markdown directory on disk | e.g. `~/Vaults/FortMemory` (default) |
| **Vault id** | Short slug; FortSignal recipient prefix | e.g. `personal`, `work`, `team-a` |
| **Agent id** | FortSignal passport name | e.g. `research-01`, `coder` |
| **Agent key file** | Downloaded Ed25519 JSON (private key stays local) | e.g. `~/Downloads/<agent>-key.json` |
| **API key** | FortSignal tenant key | `export FORTSIGNAL_API_KEY=fs_live_…` |

Recipient scheme FortMemory always uses:

```text
{vaultId}/{relative/path.md}
```

So if vault id is `work` and path is `Scratch/note.md` → recipient `work/Scratch/note.md`.  
**Policy `allowedRecipients` must use that same vault id** (not someone else’s).

### 1. FortMemory vault

```bash
# First run (interactive): default folder + choose vault id
fortmemory

# Or explicit:
fortmemory init ~/Vaults/MyVault --id myvault
fortmemory
```

### 2. FortSignal — agent passport

1. [Dashboard](https://fortsignal.com/dashboard) → **Agent Passports**  
2. Create agent → download key JSON (private key never uploaded to FortSignal)  
3. Audience / Signal Views: leave empty for basic setup  

### 3. FortSignal — policy (form or Composer)

Use **your** `{vaultId}` everywhere below.

| Field | Value |
|-------|--------|
| **Name** | e.g. `FortMemory writes` |
| **Allowed actions** | `memory.write` (add `memory.delete` only if needed) |
| **Max amount per action** | `65536` (bytes of file content) — **not `0`** (0 means max zero) |
| **Allowed recipients** | `{vaultId}/Scratch/*` (optional: `{vaultId}/Inbox/*`) |
| **Require passkey per action** | off for agent path |

Composer NL example (replace `myvault` with your vault id):

> Agent may memory.write under myvault/Scratch/* only. Max 65536 bytes per write. Deny myvault/Private/*.

### 4. FortSignal — delegation

Attach that policy to **your** agent → passkey approve → Active.

### 5. FortMemory — tokens + agent key

```bash
export FORTSIGNAL_API_KEY=fs_live_…   # your tenant key

# Local dashboard search (not FortSignal):
fortmemory token
# paste fm_… into http://127.0.0.1:7432/ Bearer → Save

# Wire FortSignal signing key for governed writes:
fortmemory agent add <agentId> --key /path/to/<agent>-key.json
```

### 6. Verify

```bash
fortmemory doctor --key /path/to/<agent>-key.json
fortmemory doctor --key /path/to/<agent>-key.json --write-probe
```

PASS write-probe → file `Scratch/_doctor_probe.md` with `last_signal_id`.

### 7. Day to day

```bash
fortmemory          # start
# Ctrl+C to stop
```

---

## Policy templates (copy into Composer — replace `myvault`)

### Research (default)

```
Agent may use memory.write under myvault/Scratch/* and myvault/Inbox/* only.
Max 65536 bytes per write. No access to myvault/Private/*.
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
| API-key ping (`POST /challenge/start`) | `fortmemory doctor` |
| `POST /agent/register` | Dashboard preferred; client method exists |
| Composer / dashboard | Settings deep links only |
| Delegation | Required; never created by FortMemory |

---

## Troubleshooting

| Symptom | Likely cause |
|---------|----------------|
| `missing FORTSIGNAL_API_KEY` | Export your `fs_live_…` key / config `api_key_env` |
| `amount_exceeds_policy` | Amount = **file size in bytes**. Set policy max to `65536` or leave empty — **not `0`** |
| `action_not_allowed` | Policy `allowedActions` must include `memory.write` |
| `recipient_not_allowed` | Recipients must match `{yourVaultId}/Scratch/*` (see vault id in config / home UI) |
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
