package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/fortsignal/fortmemory-vault/internal/agent"
	"github.com/fortsignal/fortmemory-vault/internal/config"
	"github.com/fortsignal/fortmemory-vault/internal/index"
	"github.com/fortsignal/fortmemory-vault/internal/memory"
	"github.com/fortsignal/fortmemory-vault/internal/policy"
	"github.com/fortsignal/fortmemory-vault/internal/receipts"
	"github.com/fortsignal/fortmemory-vault/internal/vault"
)

func TestHealthAndRead(t *testing.T) {
	root := t.TempDir()
	st, err := vault.New(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := st.Write(ctx, "Scratch/hi.md", []byte("# hi\n"), vault.ModeOverwrite); err != nil {
		t.Fatal(err)
	}
	idx, err := index.Open(filepath.Join(root, "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	_ = idx.Upsert(ctx, "Scratch/hi.md", []byte("# hi\n"), "")

	agents := agent.NewMemoryStore()
	tok, err := agents.Add(ctx, "research-01")
	if err != nil {
		t.Fatal(err)
	}
	recs, _ := receipts.OpenJSONL(filepath.Join(root, "receipts.jsonl"))

	cfg := config.Default()
	cfg.VaultID = "personal"
	cfg.VaultPath = root
	cfg.Bind = "127.0.0.1"
	cfg.Port = 7432

	svc := &memory.Service{
		Cfg:    cfg,
		Vault:  st,
		Index:  idx,
		Policy: policy.New(cfg.Policy),
	}

	srv := New(Deps{
		Config:   cfg,
		Version:  "test",
		Vault:    st,
		Index:    idx,
		Receipts: recs,
		Agents:   agents,
		Memory:   svc,
	})
	h := srv.Handler()

	// health no auth
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/health", nil))
	if rr.Code != 200 {
		t.Fatalf("health %d %s", rr.Code, rr.Body.String())
	}

	// read without auth
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/read?path=Scratch/hi.md", nil))
	if rr.Code != 401 {
		t.Fatalf("expected 401 got %d", rr.Code)
	}

	// read with auth
	rr = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/read?path=Scratch/hi.md", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("read %d %s", rr.Code, rr.Body.String())
	}

	// search
	rr = httptest.NewRecorder()
	body, _ := json.Marshal(map[string]any{"q": "hi", "topK": 5})
	req = httptest.NewRequest(http.MethodPost, "/v1/search", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("search %d %s", rr.Code, rr.Body.String())
	}
}
