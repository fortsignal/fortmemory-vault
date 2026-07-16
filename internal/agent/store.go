package agent

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileStore persists agent tokens under .fortmemory/agents.json.
type FileStore struct {
	path string
	mu   sync.Mutex
	data storeFile
}

type storeFile struct {
	Agents []persistedAgent `json:"agents"`
}

type persistedAgent struct {
	AgentID   string `json:"agentId"`
	TokenHash string `json:"tokenHash"`
	KeyPath   string `json:"keyPath,omitempty"`
}

// OpenFileStore loads or creates agents.json.
func OpenFileStore(path string) (*FileStore, error) {
	s := &FileStore{path: path, data: storeFile{Agents: []persistedAgent{}}}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				return nil, err
			}
			return s, s.save()
		}
		return nil, err
	}
	if len(b) > 0 {
		if err := json.Unmarshal(b, &s.data); err != nil {
			return nil, fmt.Errorf("parse agents store: %w", err)
		}
	}
	return s, nil
}

func (s *FileStore) save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// addLocked registers or rotates a local API token (caller holds lock).
func (s *FileStore) addLocked(agentID, keyPath string) (plainToken string, err error) {
	if agentID == "" {
		return "", fmt.Errorf("agentId required")
	}
	tok, err := randomToken()
	if err != nil {
		return "", err
	}
	hash := hashToken(tok)
	found := false
	for i := range s.data.Agents {
		if s.data.Agents[i].AgentID == agentID {
			s.data.Agents[i].TokenHash = hash
			if keyPath != "" {
				s.data.Agents[i].KeyPath = keyPath
			}
			found = true
			break
		}
	}
	if !found {
		s.data.Agents = append(s.data.Agents, persistedAgent{
			AgentID:   agentID,
			TokenHash: hash,
			KeyPath:   keyPath,
		})
	}
	if err := s.save(); err != nil {
		return "", err
	}
	return tok, nil
}

// Register creates/rotates token and optional key path.
func (s *FileStore) Register(ctx context.Context, agentID, keyPath string) (plainToken string, err error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addLocked(agentID, keyPath)
}

// SetKeyPath updates the signing key path for an agent.
func (s *FileStore) SetKeyPath(agentID, keyPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Agents {
		if s.data.Agents[i].AgentID == agentID {
			s.data.Agents[i].KeyPath = keyPath
			return s.save()
		}
	}
	return fmt.Errorf("agent %q not found", agentID)
}

func (s *FileStore) LookupByToken(ctx context.Context, token string) (*Record, error) {
	_ = ctx
	if token == "" {
		return nil, fmt.Errorf("unauthorized")
	}
	h := hashToken(token)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range s.data.Agents {
		if a.TokenHash == h {
			return &Record{AgentID: a.AgentID, TokenHash: a.TokenHash, KeyPath: a.KeyPath}, nil
		}
	}
	return nil, fmt.Errorf("unauthorized")
}

func (s *FileStore) Get(ctx context.Context, agentID string) (*Record, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range s.data.Agents {
		if a.AgentID == agentID {
			return &Record{AgentID: a.AgentID, TokenHash: a.TokenHash, KeyPath: a.KeyPath}, nil
		}
	}
	return nil, fmt.Errorf("agent not found")
}

func (s *FileStore) List(ctx context.Context) ([]Record, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Record, 0, len(s.data.Agents))
	for _, a := range s.data.Agents {
		out = append(out, Record{AgentID: a.AgentID, TokenHash: a.TokenHash, KeyPath: a.KeyPath})
	}
	return out, nil
}

func hashToken(tok string) string {
	sum := sha256.Sum256([]byte(tok))
	return hex.EncodeToString(sum[:])
}

func randomToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "fm_" + hex.EncodeToString(b[:]), nil
}
