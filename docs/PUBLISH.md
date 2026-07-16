# Publishing FortMemory (public repo checklist)

**Remote:** https://github.com/fortsignal/fortmemory-vault  
**Module:** `github.com/fortsignal/fortmemory-vault`  
**License:** Apache-2.0 (`LICENSE` + `NOTICE`)

## Before first push

- [x] No secrets in tree  
- [x] `.gitignore` covers `.env`, keys, `bin/`, sqlite, `.fortmemory/`  
- [x] `LICENSE` Apache-2.0 + `NOTICE`  
- [x] `SECURITY.md` at repo root  
- [x] `go test ./...` green  
- [x] Professional README  

## Push (from this machine)

Local `main` already has the initial commit. Authenticate as a user/org with write access to `fortsignal/fortmemory-vault`, then:

```bash
cd ~/projects/fortmemory-vault
git remote -v
# preferred:
git remote set-url origin git@github.com:fortsignal/fortmemory-vault.git
# or HTTPS with a PAT:
# git remote set-url origin https://github.com/fortsignal/fortmemory-vault.git

git push -u origin main
```

If SSH is bound to the wrong GitHub user, use:

```bash
GIT_SSH_COMMAND='ssh -i ~/.ssh/YOUR_FORTSIGNAL_KEY -o IdentitiesOnly=yes' \
  git push -u origin main
```

Or GitHub CLI (after `gh auth login`):

```bash
gh auth login
gh repo sync  # or just git push
```

## Install for others (after push)

```bash
go install github.com/fortsignal/fortmemory-vault/cmd/fortmemory@latest

# or
git clone https://github.com/fortsignal/fortmemory-vault.git
cd fortmemory-vault && make build
```

## Do not publish

- Agent private keys / `agent-key.json`  
- `FORTSIGNAL_API_KEY`  
- Real customer vaults  
- FortSignal server source (separate product)  
