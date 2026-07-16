package agent

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSignerAndSign(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	seed := priv.Seed()
	pub := priv.Public().(ed25519.PublicKey)

	kf := map[string]string{
		"agentId":    "research-01",
		"privateKey": base64.RawURLEncoding.EncodeToString(seed),
		"publicKey":  base64.RawURLEncoding.EncodeToString(pub),
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "agent-key.json")
	b, _ := json.Marshal(kf)
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}

	s, err := LoadSigner(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.AgentID() != "research-01" {
		t.Fatal(s.AgentID())
	}

	// Sign a fake challenge (32 random-ish bytes as base64url)
	msg := []byte("challenge-bytes-for-test-xx")
	ch := base64.RawURLEncoding.EncodeToString(msg)
	sigB64, err := s.SignChallenge(ch)
	if err != nil {
		t.Fatal(err)
	}
	sig, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil {
		t.Fatal(err)
	}
	if !ed25519.Verify(pub, msg, sig) {
		t.Fatal("signature verify failed")
	}
}
