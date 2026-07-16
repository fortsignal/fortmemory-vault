# Dashboard source

`index.html` is the operator UI (Home / Search / Activity / Settings).

**Build note:** copy into `internal/server/static/` for `//go:embed` (done in tree).
Edit either file and keep them in sync, or copy after changes:

```bash
cp web/index.html internal/server/static/index.html
```

Served at `http://127.0.0.1:7432/` by `fortmemory serve`.
