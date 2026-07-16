package fortsignal

import (
	"strings"
	"testing"
)

func TestContentHashStable(t *testing.T) {
	h1 := ContentHash([]byte("hello"))
	h2 := ContentHash([]byte("hello"))
	if h1 != h2 {
		t.Fatalf("hash unstable")
	}
	if !strings.HasPrefix(h1, "sha256:") {
		t.Fatalf("prefix: %s", h1)
	}
}

func TestEncodeRecipientShort(t *testing.T) {
	rec, metaPath, err := EncodeRecipient("personal", "Scratch/note.md")
	if err != nil {
		t.Fatal(err)
	}
	if rec != "personal/Scratch/note.md" {
		t.Fatalf("got %q", rec)
	}
	if metaPath != "" {
		t.Fatalf("expected empty meta path, got %q", metaPath)
	}
}

func TestEncodeRecipientLongFallsBack(t *testing.T) {
	long := strings.Repeat("a/", 200) + "note.md"
	rec, metaPath, err := EncodeRecipient("personal", long)
	if err != nil {
		t.Fatal(err)
	}
	if len(rec) > MaxRecipientLen {
		t.Fatalf("recipient too long: %d", len(rec))
	}
	if !strings.Contains(rec, "/#/") {
		t.Fatalf("expected hash fallback, got %q", rec)
	}
	if metaPath == "" {
		t.Fatal("expected meta path for long recipient")
	}
}

func TestEncodeRecipientTraversal(t *testing.T) {
	_, _, err := EncodeRecipient("personal", "../etc/passwd")
	if err == nil {
		t.Fatal("expected error")
	}
}
