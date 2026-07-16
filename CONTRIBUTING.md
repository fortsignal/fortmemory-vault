# Contributing to FortMemory

Thank you for helping build verifiable local memory.

## License

By contributing, you agree that your contributions are licensed under the **Apache License 2.0** (see `LICENSE`).

## What we welcome

- Bug fixes and tests  
- Vault / path-jail / index improvements  
- Tunnel helpers (Cloudflare, Tailscale)  
- Docs, examples, agent adapters  
- Performance and security hardening of the **local** server  

## What belongs elsewhere

| Idea | Where it goes |
|------|----------------|
| NL Policy Composer features | FortSignal product (not this repo) |
| R2 multi-tenant sync billing | FortVault commercial |
| SSO / enterprise IdP | Commercial / Enterprise |
| “Make FortSignal free unlimited” | Out of scope |

See [docs/OPEN-CORE.md](./docs/OPEN-CORE.md) and [docs/COMMERCIAL.md](./docs/COMMERCIAL.md).

## Dev setup

```bash
go test ./...
go build -o bin/fortmemory ./cmd/fortmemory
```

## Security

Report vulnerabilities privately to **hr@fortsignal.com** (subject: `FortMemory security`) before public issues when possible.

## Code style

- Keep the local-first / fail-closed security model  
- No phone-home in the OSS binary by default  
- Prefer small PRs with tests  
