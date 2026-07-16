package agent

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// keyFileJSON matches fortsignal-deepagents / dashboard download format.
type keyFileJSON struct {
	AgentID    string `json:"agentId"`
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

// LoadSigner opens a Deep Agents-style JSON key file and returns a FileSigner.
func LoadSigner(path string) (*FileSigner, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var kf keyFileJSON
	if err := json.Unmarshal(data, &kf); err != nil {
		return nil, fmt.Errorf("parse agent key file: %w", err)
	}
	if strings.TrimSpace(kf.AgentID) == "" {
		return nil, fmt.Errorf("agent key file missing agentId")
	}
	if strings.TrimSpace(kf.PrivateKey) == "" {
		return nil, fmt.Errorf("agent key file missing privateKey")
	}

	seedOrKey, err := decodeB64URL(kf.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("privateKey: %w", err)
	}

	var priv ed25519.PrivateKey
	switch len(seedOrKey) {
	case ed25519.SeedSize:
		priv = ed25519.NewKeyFromSeed(seedOrKey)
	case ed25519.PrivateKeySize:
		priv = ed25519.PrivateKey(seedOrKey)
	default:
		return nil, fmt.Errorf("privateKey must be 32-byte seed or 64-byte key (got %d bytes)", len(seedOrKey))
	}

	pub := priv.Public().(ed25519.PublicKey)
	pubB64 := base64.RawURLEncoding.EncodeToString(pub)
	if kf.PublicKey != "" {
		// Optional consistency check
		want, err := decodeB64URL(kf.PublicKey)
		if err == nil && len(want) == ed25519.PublicKeySize {
			if !bytesEqual(want, pub) {
				return nil, fmt.Errorf("publicKey does not match privateKey")
			}
			pubB64 = base64.RawURLEncoding.EncodeToString(want)
		}
	}

	return &FileSigner{
		agentID:    kf.AgentID,
		privateKey: priv,
		publicB64:  pubB64,
	}, nil
}

// SignChallenge signs base64url-encoded challenge bytes; returns base64url signature.
func (s *FileSigner) SignChallenge(challengeB64 string) (string, error) {
	if s == nil || len(s.privateKey) == 0 {
		return "", fmt.Errorf("signer not loaded")
	}
	msg, err := decodeB64URL(challengeB64)
	if err != nil {
		return "", fmt.Errorf("challenge: %w", err)
	}
	sig := ed25519.Sign(s.privateKey, msg)
	return base64.RawURLEncoding.EncodeToString(sig), nil
}

func (s *FileSigner) PublicKeyB64() string { return s.publicB64 }
func (s *FileSigner) AgentID() string      { return s.agentID }

func decodeB64URL(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	// Accept both raw and standard base64url (with/without padding).
	if b, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	return base64.URLEncoding.DecodeString(s)
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
