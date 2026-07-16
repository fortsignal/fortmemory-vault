package memory

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"crypto/ed25519"

	"github.com/fortsignal/fortmemory-vault/internal/agent"
	"github.com/fortsignal/fortmemory-vault/internal/config"
	"github.com/fortsignal/fortmemory-vault/internal/fortsignal"
	"github.com/fortsignal/fortmemory-vault/internal/policy"
	"github.com/fortsignal/fortmemory-vault/internal/vault"
)

type fakeEnforcer struct {
	start  *fortsignal.ChallengeStartResult
	verify *fortsignal.VerifyResult
	startErr error
}

func (f *fakeEnforcer) ChallengeStart(ctx context.Context, req fortsignal.ChallengeStartRequest) (*fortsignal.ChallengeStartResult, error) {
	_ = ctx
	_ = req
	if f.startErr != nil {
		return nil, f.startErr
	}
	return f.start, nil
}

func (f *fakeEnforcer) ChallengeVerify(ctx context.Context, req fortsignal.ChallengeVerifyRequest) (*fortsignal.VerifyResult, error) {
	_ = ctx
	_ = req
	return f.verify, nil
}

type seedSigner struct {
	id   string
	priv ed25519.PrivateKey
}

func (s *seedSigner) AgentID() string      { return s.id }
func (s *seedSigner) PublicKeyB64() string { return "" }
func (s *seedSigner) SignChallenge(challengeB64 string) (string, error) {
	msg, err := base64.RawURLEncoding.DecodeString(challengeB64)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(ed25519.Sign(s.priv, msg)), nil
}

func TestWriteAllow(t *testing.T) {
	root := t.TempDir()
	st, err := vault.New(root)
	if err != nil {
		t.Fatal(err)
	}
	_, priv, _ := ed25519.GenerateKey(nil)
	msg := []byte("chal")
	ch := base64.RawURLEncoding.EncodeToString(msg)

	svc := &Service{
		Cfg: config.Config{
			VaultID: "personal",
			Security: config.SecurityConfig{FailClosedOnFortSignal: true},
		},
		Vault:  st,
		Policy: policy.New(config.LocalPolicyConfig{AllowWrite: []string{"Scratch/**"}}),
		FortSignal: &fakeEnforcer{
			start: &fortsignal.ChallengeStartResult{
				Challenge:    ch,
				AgentID:      "a1",
				DelegationID: "d1",
			},
			verify: &fortsignal.VerifyResult{
				Decision:     "allow",
				SignalID:     "sig-99",
				VerifiedBy:   "agent",
				VerifiedAt:   "2026-07-16T12:00:00Z",
				DelegationID: "d1",
			},
		},
		Signer: &seedSigner{id: "a1", priv: priv},
	}

	res, err := svc.Write(context.Background(), WriteInput{
		AgentID: "a1",
		Path:    "Scratch/note.md",
		Content: []byte("# hi\n"),
		Mode:    vault.ModeOverwrite,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Decision != "allow" || res.SignalID != "sig-99" {
		t.Fatalf("%+v", res)
	}
	got, err := st.Read(context.Background(), "Scratch/note.md")
	if err != nil {
		t.Fatal(err)
	}
	// Body is annotated with last_signal_id after allow (not part of signed contentHash).
	if !strings.Contains(string(got), "# hi") || !strings.Contains(string(got), "last_signal_id: sig-99") {
		t.Fatalf("file: %q", got)
	}
}

func TestWriteLocalPolicyDenyNoFS(t *testing.T) {
	root := t.TempDir()
	st, _ := vault.New(root)
	_, priv, _ := ed25519.GenerateKey(nil)
	svc := &Service{
		Cfg:    config.Config{VaultID: "personal"},
		Vault:  st,
		Policy: policy.New(config.LocalPolicyConfig{AllowWrite: []string{"Scratch/**"}}),
		FortSignal: &fakeEnforcer{
			start: &fortsignal.ChallengeStartResult{Challenge: "x"},
		},
		Signer: &seedSigner{id: "a1", priv: priv},
	}
	res, err := svc.Write(context.Background(), WriteInput{
		AgentID: "a1",
		Path:    "Private/secret.md",
		Content: []byte("nope"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Decision != "deny" {
		t.Fatalf("%+v", res)
	}
	if _, err := st.Read(context.Background(), "Private/secret.md"); err == nil {
		t.Fatal("file should not exist")
	}
}

// ensure agent.Signer interface used
var _ agent.Signer = (*seedSigner)(nil)
