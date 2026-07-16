# Commercial Surface (Not Open Source)

This document defines what **is not** part of the Apache-2.0 FortMemory core.  
See [OPEN-CORE.md](./OPEN-CORE.md) for the full strategy.

## FortVault (paid cloud)

- Encrypted object sync to **Cloudflare R2**  
- Multi-device pairing and conflict policy UI  
- Team shared vaults and seat management  
- Hosted audit export / retention policies  
- Managed Cloudflare Tunnel identity for orgs (optional)  

## Enterprise pack

- SSO / SAML / OIDC for team console  
- SIEM-friendly bulk export + webhooks  
- Legal hold / retention admin  
- Self-host runbooks with FortSignal Enterprise  
- Security review / procurement support  
- SLA and dedicated onboarding  

## FortSignal (existing product)

Not open-sourced as “free unlimited governance”:

- Hosted challenge/verify API  
- NL Policy Composer  
- Agent Passports dashboard + delegation ceremonies  
- Signal Views administration  
- Usage metering and plan limits  
- Enterprise self-host of FortSignal  

FortMemory OSS **integrates** via public API; customers pay FortSignal as they scale verifications.

## Support

| Free | Paid |
|------|------|
| GitHub issues / discussions | Email / shared Slack |
| Community examples | Architecture review |
| Docs | SLA, private incident help |

## What we will not sell as a fake open-core trap

- License keys required to run `fortmemory serve` locally  
- Disabling FortSignal client in OSS builds  
- Closed-source path jail or receipts (trust-critical path stays open)  
