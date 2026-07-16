# MCP integration

FortMemory exposes agent tools over the [Model Context Protocol](https://modelcontextprotocol.io) on **stdio**.

## Run

```bash
export FORTSIGNAL_API_KEY=fs_live_...
fortmemory mcp \
  --config ~/Vaults/Personal/.fortmemory/config.toml \
  --agent research-01 \
  --key ~/agent-key.json   # if not already registered
```

## Tools

| Tool | Description |
|------|-------------|
| `memory_search` | FTS search (`q`, optional `topK`, `pathPrefix`) |
| `memory_read` | Read note by relative `path` |
| `memory_write` | FortSignal-gated write (`path`, `content`, optional `mode`) |
| `memory_delete` | FortSignal-gated delete (`path`) |

Writes/deletes return JSON `{ decision, signalId, reason? }`.

## Cursor / Claude example config

```json
{
  "mcpServers": {
    "fortmemory": {
      "command": "/path/to/fortmemory",
      "args": [
        "mcp",
        "--config", "/home/you/Vaults/Personal/.fortmemory/config.toml",
        "--agent", "research-01",
        "--key", "/home/you/agent-key.json"
      ],
      "env": {
        "FORTSIGNAL_API_KEY": "fs_live_..."
      }
    }
  }
}
```

Agents should use these tools instead of raw filesystem writes into the vault.
