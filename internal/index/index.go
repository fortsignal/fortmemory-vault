// Package index maintains the rebuildable search index under .fortmemory/index.sqlite.
//
// MVP: SQLite FTS5 keyword search.
// The index is never source of truth — Markdown files are.
package index

import (
	"context"
	"encoding/json"
)

// Hit is one search result.
type Hit struct {
	Path         string  `json:"path"`
	Score        float64 `json:"score"`
	Excerpt      string  `json:"excerpt"`
	Tags         []string `json:"tags,omitempty"`
	Sensitivity  string  `json:"sensitivity,omitempty"`
	LastSignalID string  `json:"lastSignalId,omitempty"`
}

// SearchRequest is the engine query.
type SearchRequest struct {
	Query      string   `json:"q"`
	TopK       int      `json:"topK"`
	Tags       []string `json:"tags,omitempty"`
	PathPrefix string   `json:"pathPrefix,omitempty"`
}

// UnmarshalJSON accepts both "q" and "query".
func (r *SearchRequest) UnmarshalJSON(b []byte) error {
	type alias SearchRequest
	var a struct {
		alias
		QueryAlt string `json:"query"`
	}
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*r = SearchRequest(a.alias)
	if r.Query == "" {
		r.Query = a.QueryAlt
	}
	return nil
}

// Index is the search surface.
type Index interface {
	Search(ctx context.Context, req SearchRequest) ([]Hit, error)
	Upsert(ctx context.Context, path string, content []byte, lastSignalID string) error
	Remove(ctx context.Context, path string) error
	ReindexAll(ctx context.Context) error
	Stats(ctx context.Context) (files int, embedPending int, err error)
	Close() error
}
