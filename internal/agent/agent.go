// Package agent manages local API credentials mapped to FortSignal agentIds
// and optional local Ed25519 signers for CLI / co-located dogfood.
//
// Deep Agents key file compatibility:
//
//	{"agentId":"…","privateKey":"<base64url>","publicKey":"…"}
package agent

import (
	"context"
	"crypto/ed25519"
	"fmt"
)

// Record is a locally known agent.
type Record struct {
	AgentID   string
	TokenHash string
	KeyPath   string // optional Ed25519 key file for server-side signing
}

// Store persists agent API tokens under .fortmemory/.
type Store interface {
	Add(ctx context.Context, agentID string) (plainToken string, err error)
	LookupByToken(ctx context.Context, token string) (*Record, error)
	Get(ctx context.Context, agentID string) (*Record, error)
	List(ctx context.Context) ([]Record, error)
}

// Signer signs FortSignal challenges for local-signer mode.
type Signer interface {
	SignChallenge(challengeB64 string) (signatureB64 string, err error)
	PublicKeyB64() string
	AgentID() string
}

// FileSigner holds Ed25519 material loaded from a key file.
type FileSigner struct {
	agentID    string
	privateKey ed25519.PrivateKey
	publicB64  string
}

// Add implements Store on FileStore (rotates token; keeps existing keyPath).
func (s *FileStore) Add(ctx context.Context, agentID string) (string, error) {
	return s.Register(ctx, agentID, "")
}

var _ Store = (*FileStore)(nil)

// MemoryStore is an in-memory store for tests.
type MemoryStore struct {
	byToken map[string]*Record
	byID    map[string]*Record
}

// NewMemoryStore returns an empty token map.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		byToken: map[string]*Record{},
		byID:    map[string]*Record{},
	}
}

func (m *MemoryStore) Add(ctx context.Context, agentID string) (string, error) {
	_ = ctx
	tok, err := randomToken()
	if err != nil {
		return "", err
	}
	rec := &Record{AgentID: agentID, TokenHash: hashToken(tok)}
	m.byToken[tok] = rec
	m.byID[agentID] = rec
	return tok, nil
}

// AddWithKey is a test helper to register token+keyPath.
func (m *MemoryStore) AddWithKey(ctx context.Context, agentID, keyPath string) (string, error) {
	tok, err := m.Add(ctx, agentID)
	if err != nil {
		return "", err
	}
	m.byToken[tok].KeyPath = keyPath
	m.byID[agentID].KeyPath = keyPath
	return tok, nil
}

func (m *MemoryStore) LookupByToken(ctx context.Context, token string) (*Record, error) {
	_ = ctx
	if r, ok := m.byToken[token]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("unauthorized")
}

func (m *MemoryStore) Get(ctx context.Context, agentID string) (*Record, error) {
	_ = ctx
	if r, ok := m.byID[agentID]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("agent not found")
}

func (m *MemoryStore) List(ctx context.Context) ([]Record, error) {
	_ = ctx
	out := make([]Record, 0, len(m.byID))
	for _, r := range m.byID {
		out = append(out, *r)
	}
	return out, nil
}
