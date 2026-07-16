package receipts

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"
)

// JSONL is a simple append-only receipt log (MVP).
type JSONL struct {
	Path string
	mu   sync.Mutex
}

// OpenJSONL creates parent dir if needed.
func OpenJSONL(path string) (*JSONL, error) {
	if err := os.MkdirAll(dirOf(path), 0o700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	_ = f.Close()
	return &JSONL{Path: path}, nil
}

func (j *JSONL) Append(ctx context.Context, rec Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	if rec.ID == "" {
		rec.ID = newID()
	}
	if rec.VerifiedAt.IsZero() {
		rec.VerifiedAt = time.Now().UTC()
	}
	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(j.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(append(b, '\n')); err != nil {
		return err
	}
	return nil
}

func (j *JSONL) List(ctx context.Context, q Query) ([]Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	j.mu.Lock()
	defer j.mu.Unlock()

	f, err := os.Open(j.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	var all []Record
	sc := bufio.NewScanner(f)
	// increase buffer for large lines
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var rec Record
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}
		if q.Action != "" && rec.Action != q.Action {
			continue
		}
		if q.Decision != "" && rec.Decision != q.Decision {
			continue
		}
		if q.AgentID != "" && rec.AgentID != q.AgentID {
			continue
		}
		if q.PathPrefix != "" && !strings.HasPrefix(rec.Path, q.PathPrefix) {
			continue
		}
		all = append(all, rec)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	// newest last in file → reverse
	for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
		all[i], all[j] = all[j], all[i]
	}
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (j *JSONL) Close() error { return nil }

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return "."
}

func newID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// Ensure interface
var _ Store = (*JSONL)(nil)
