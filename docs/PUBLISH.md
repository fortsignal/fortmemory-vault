# Publishing FortMemory (public repo checklist)

## Before first push

- [ ] No secrets in tree (`rg -i 'fs_live_|fm_[a-f0-9]{20}|BEGIN PRIVATE'`)  
- [ ] `.gitignore` covers `.env`, keys, `bin/`, sqlite, `.fortmemory/`  
- [ ] `LICENSE` Apache-2.0 + `NOTICE` present  
- [ ] `go test ./...` green  
- [ ] README describes: local free · FortSignal for governance · no domain required for MVP  
- [ ] Prefer empty history if this folder ever held private keys  

## Suggested remote

```text
github.com/fortsignal/fortmemory-vault
```

```bash
cd ~/projects/fortmemory-vault
git init
git add .
git status   # review carefully
git commit -m "Initial public release: local FortMemory server"
# create empty public repo on GitHub, then:
git branch -M main
git remote add origin git@github.com:fortsignal/fortmemory.git
git push -u origin main
```

## Install for others

```bash
go install github.com/fortsignal/fortmemory-vault/cmd/fortmemory@latest
# or
git clone … && cd fortmemory && make build
```

## Do not publish

- Agent private keys / `agent-key.json`  
- `FORTSIGNAL_API_KEY`  
- Real customer vaults  
- FortSignal server source (separate product)  
