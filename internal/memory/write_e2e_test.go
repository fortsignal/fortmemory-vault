package memory_test

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fortsignal/fortmemory-vault/internal/agent"
	"github.com/fortsignal/fortmemory-vault/internal/config"
	"github.com/fortsignal/fortmemory-vault/internal/fortsignal"
	"github.com/fortsignal/fortmemory-vault/internal/memory"
	"github.com/fortsignal/fortmemory-vault/internal/policy"
	"github.com/fortsignal/fortmemory-vault/internal/receipts"
	"github.com/fortsignal/fortmemory-vault/internal/vault"
)

// End-to-end: real HTTP fortsignal client + vault write (mocked FortSignal server).
func TestWriteE2EWithHTTPClient(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	pub := priv.Public().(ed25519.PublicKey)

	var gotStart map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/challenge/start":
			_ = json.NewDecoder(r.Body).Decode(&gotStart)
			msg := []byte("fixed-challenge-bytes!!") // 24 bytes
			_ = json.NewEncoder(w).Encode(map[string]any{
				"challenge":    base64.RawURLEncoding.EncodeToString(msg),
				"agentId":      "research-01",
				"delegationId": "del_1",
				"expiresIn":    60,
			})
		case "/challenge/verify":
			var body struct {
				AgentID   string `json:"agentId"`
				Challenge string `json:"challenge"`
				Signature string `json:"signature"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			msg, _ := base64.RawURLEncoding.DecodeString(body.Challenge)
			sig, _ := base64.RawURLEncoding.DecodeString(body.Signature)
			if !ed25519.Verify(pub, msg, sig) {
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]any{"decision": "deny", "reason": "verification_failed"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"decision":     "allow",
				"signalId":     "e2e-signal",
				"verifiedBy":   "agent",
				"verifiedAt":   "2026-07-16T15:00:00Z",
				"agentId":      "research-01",
				"delegationId": "del_1",
				"policyId":     "pol_1",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	root := t.TempDir()
	// key file
	kf := map[string]string{
		"agentId":    "research-01",
		"privateKey": base64.RawURLEncoding.EncodeToString(priv.Seed()),
		"publicKey":  base64.RawURLEncoding.EncodeToString(pub),
	}
	keyPath := filepath.Join(root, "agent-key.json")
	b, _ := json.Marshal(kf)
	if err := os.WriteFile(keyPath, b, 0o600); err != nil {
		t.Fatal(err)
	}
	signer, err := agent.LoadSigner(keyPath)
	if err != nil {
		t.Fatal(err)
	}

	store, _ := vault.New(root)
	recPath := filepath.Join(root, ".fortmemory", "receipts.jsonl")
	rec, err := receipts.OpenJSONL(recPath)
	if err != nil {
		t.Fatal(err)
	}

	svc := &memory.Service{
		Cfg: config.Config{
			VaultID:  "personal",
			Security: config.SecurityConfig{FailClosedOnFortSignal: true},
		},
		Vault:      store,
		Receipts:   rec,
		Policy:     policy.New(config.LocalPolicyConfig{AllowWrite: []string{"Scratch/**"}}),
		FortSignal: fortsignal.New("test-key", srv.URL),
		Signer:     signer,
	}

	res, err := svc.Write(context.Background(), memory.WriteInput{
		AgentID: "research-01",
		Path:    "Scratch/e2e.md",
		Content: []byte("# e2e\n"),
		Mode:    vault.ModeOverwrite,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Decision != "allow" || res.SignalID != "e2e-signal" {
		t.Fatalf("%+v", res)
	}
	if gotStart["action"] != "memory.write" {
		t.Fatalf("start action: %v", gotStart["action"])
	}
	if gotStart["recipient"] != "personal/Scratch/e2e.md" {
		t.Fatalf("recipient: %v", gotStart["recipient"])
	}
	data, err := store.Read(context.Background(), "Scratch/e2e.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "# e2e") || !strings.Contains(string(data), "last_signal_id: e2e-signal") {
		t.Fatalf("vault: %q", data)
	}
}
