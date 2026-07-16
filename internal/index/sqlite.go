package index

import (
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	_ "modernc.org/sqlite"

	"github.com/fortsignal/fortmemory-vault/internal/embed"
)

// SQLite is FTS5-backed search with optional vector store.
type SQLite struct {
	path    string
	db      *sql.DB
	embedder embed.Provider // optional
}

// Open opens or creates index.sqlite with FTS5 (+ embeddings table).
func Open(dbPath string) (*SQLite, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000;`); err != nil {
		_ = db.Close()
		return nil, err
	}
	schema := `
CREATE VIRTUAL TABLE IF NOT EXISTS docs USING fts5(
  path UNINDEXED,
  body,
  signal_id UNINDEXED,
  tokenize = 'porter'
);
CREATE TABLE IF NOT EXISTS embeddings (
  path TEXT PRIMARY KEY,
  dim INTEGER NOT NULL,
  vec BLOB NOT NULL
);
`
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}
	return &SQLite{path: dbPath, db: db}, nil
}

// SetEmbedder enables hybrid search when non-nil.
func (s *SQLite) SetEmbedder(p embed.Provider) {
	s.embedder = p
}

func (s *SQLite) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *SQLite) Upsert(ctx context.Context, path string, content []byte, lastSignalID string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM docs WHERE path = ?`, path); err != nil {
		return err
	}
	body := string(content)
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO docs(path, body, signal_id) VALUES (?, ?, ?)`,
		path, body, lastSignalID,
	); err != nil {
		return err
	}
	// Best-effort embedding — never fail the write/index on embed errors.
	if s.embedder != nil && s.embedder.Name() != "none" {
		vec, err := s.embedder.Embed(ctx, body)
		if err == nil && len(vec) > 0 {
			_ = s.storeVec(ctx, path, vec)
		} else {
			_, _ = s.db.ExecContext(ctx, `DELETE FROM embeddings WHERE path = ?`, path)
		}
	}
	return nil
}

func (s *SQLite) Remove(ctx context.Context, path string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM docs WHERE path = ?`, path)
	_, _ = s.db.ExecContext(ctx, `DELETE FROM embeddings WHERE path = ?`, path)
	return err
}

func (s *SQLite) Search(ctx context.Context, req SearchRequest) ([]Hit, error) {
	q := strings.TrimSpace(req.Query)
	if q == "" {
		return nil, fmt.Errorf("query required")
	}
	topK := req.TopK
	if topK <= 0 {
		topK = 8
	}
	if topK > 50 {
		topK = 50
	}

	ftsHits, err := s.searchFTS(ctx, q, req.PathPrefix, topK*3)
	if err != nil {
		return nil, err
	}

	// Hybrid: blend FTS with cosine when embedder + stored vectors available.
	if s.embedder != nil && s.embedder.Name() != "none" && len(ftsHits) > 0 {
		qvec, err := s.embedder.Embed(ctx, q)
		if err == nil && len(qvec) > 0 {
			for i := range ftsHits {
				if v, ok := s.loadVec(ctx, ftsHits[i].Path); ok {
					cos := cosine(qvec, v)
					// Normalize rough blend: FTS score already inverted bm25; boost by cosine [0,1]
					ftsHits[i].Score = ftsHits[i].Score + cos*10
				}
			}
			sort.Slice(ftsHits, func(i, j int) bool { return ftsHits[i].Score > ftsHits[j].Score })
		}
	}

	if len(ftsHits) > topK {
		ftsHits = ftsHits[:topK]
	}
	return ftsHits, nil
}

func (s *SQLite) searchFTS(ctx context.Context, q, pathPrefix string, limit int) ([]Hit, error) {
	match := ftsMatchQuery(q)
	rows, err := s.db.QueryContext(ctx, `
SELECT path, signal_id,
       snippet(docs, 1, '', '', '…', 24),
       bm25(docs)
FROM docs
WHERE docs MATCH ?
ORDER BY bm25(docs)
LIMIT ?`, match, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hits []Hit
	for rows.Next() {
		var h Hit
		var sig sql.NullString
		var excerpt string
		var score float64
		if err := rows.Scan(&h.Path, &sig, &excerpt, &score); err != nil {
			return nil, err
		}
		if pathPrefix != "" && !strings.HasPrefix(h.Path, pathPrefix) {
			continue
		}
		h.LastSignalID = sig.String
		h.Excerpt = excerpt
		h.Score = -score
		hits = append(hits, h)
	}
	return hits, rows.Err()
}

func (s *SQLite) Stats(ctx context.Context) (files int, embedPending int, err error) {
	err = s.db.QueryRowContext(ctx, `SELECT count(*) FROM docs`).Scan(&files)
	if err != nil {
		return 0, 0, err
	}
	var emb int
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM embeddings`).Scan(&emb)
	if files > emb {
		embedPending = files - emb
	}
	return files, embedPending, nil
}

func (s *SQLite) ReindexAll(ctx context.Context) error {
	_ = ctx
	return fmt.Errorf("use memory.Service.Reindex to rebuild from vault files")
}

func (s *SQLite) storeVec(ctx context.Context, path string, vec []float32) error {
	blob := float32ToBytes(vec)
	_, err := s.db.ExecContext(ctx, `
INSERT INTO embeddings(path, dim, vec) VALUES(?,?,?)
ON CONFLICT(path) DO UPDATE SET dim=excluded.dim, vec=excluded.vec`,
		path, len(vec), blob)
	return err
}

func (s *SQLite) loadVec(ctx context.Context, path string) ([]float32, bool) {
	var dim int
	var blob []byte
	err := s.db.QueryRowContext(ctx, `SELECT dim, vec FROM embeddings WHERE path = ?`, path).Scan(&dim, &blob)
	if err != nil || dim <= 0 {
		return nil, false
	}
	v := bytesToFloat32(blob)
	if len(v) != dim {
		return nil, false
	}
	return v, true
}

func float32ToBytes(v []float32) []byte {
	b := make([]byte, 4*len(v))
	for i, f := range v {
		binary.LittleEndian.PutUint32(b[i*4:], math.Float32bits(f))
	}
	return b
}

func bytesToFloat32(b []byte) []float32 {
	if len(b)%4 != 0 {
		return nil
	}
	n := len(b) / 4
	v := make([]float32, n)
	for i := 0; i < n; i++ {
		v[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
	}
	return v
}

func cosine(a, b []float32) float64 {
	n := len(a)
	if n == 0 || n != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := 0; i < n; i++ {
		fa, fb := float64(a[i]), float64(b[i])
		dot += fa * fb
		na += fa * fa
		nb += fb * fb
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

func ftsMatchQuery(q string) string {
	q = strings.TrimSpace(q)
	if strings.ContainsAny(q, `"*()`) {
		return q
	}
	parts := strings.Fields(q)
	if len(parts) == 0 {
		return q
	}
	for i, p := range parts {
		p = strings.ReplaceAll(p, `"`, "")
		if p == "" {
			continue
		}
		parts[i] = `"` + p + `"`
	}
	return strings.Join(parts, " ")
}

// ExcerptFallback builds a short excerpt when snippet is empty.
func ExcerptFallback(body string, max int) string {
	body = strings.TrimSpace(body)
	if max <= 0 {
		max = 160
	}
	if utf8.RuneCountInString(body) <= max {
		return body
	}
	r := []rune(body)
	return string(r[:max]) + "…"
}
