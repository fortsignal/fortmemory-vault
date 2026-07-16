package index

import (
	"context"
	"path/filepath"
	"testing"
)

func TestFTSRoundTrip(t *testing.T) {
	dir := t.TempDir()
	idx, err := Open(filepath.Join(dir, "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()

	ctx := context.Background()
	if err := idx.Upsert(ctx, "Scratch/a.md", []byte("# Stripe webhooks\n\ntimeout handling"), "sig1"); err != nil {
		t.Fatal(err)
	}
	if err := idx.Upsert(ctx, "Scratch/b.md", []byte("# Unrelated note\n\ncooking recipes"), ""); err != nil {
		t.Fatal(err)
	}

	hits, err := idx.Search(ctx, SearchRequest{Query: "stripe timeout", TopK: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) == 0 {
		t.Fatal("expected hits")
	}
	if hits[0].Path != "Scratch/a.md" {
		t.Fatalf("got %+v", hits)
	}
	n, _, err := idx.Stats(ctx)
	if err != nil || n != 2 {
		t.Fatalf("stats %d %v", n, err)
	}
}
