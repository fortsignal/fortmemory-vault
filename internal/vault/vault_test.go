package vault

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPathJailTraversal(t *testing.T) {
	root := t.TempDir()
	s, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"../etc/passwd", "..\\secret", "/etc/passwd", "foo/../../etc/passwd"} {
		if _, _, err := s.Resolve(p, true); err == nil {
			t.Fatalf("expected deny for %q", p)
		}
	}
}

func TestWriteReadRoundTrip(t *testing.T) {
	root := t.TempDir()
	s, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	body := []byte("# hello\n\nworld\n")
	if err := s.Write(ctx, "Scratch/note.md", body, ModeOverwrite); err != nil {
		t.Fatal(err)
	}
	got, err := s.Read(ctx, "Scratch/note.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(body) {
		t.Fatalf("got %q", got)
	}
}

func TestRejectFortMemoryWrite(t *testing.T) {
	root := t.TempDir()
	s, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Write(context.Background(), ".fortmemory/config.toml", []byte("x"), ModeOverwrite); err == nil {
		t.Fatal("expected reject")
	}
}

func TestCreateModeFailsIfExists(t *testing.T) {
	root := t.TempDir()
	s, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	_ = os.MkdirAll(filepath.Join(root, "Scratch"), 0o755)
	if err := s.Write(ctx, "Scratch/a.md", []byte("1"), ModeCreate); err != nil {
		t.Fatal(err)
	}
	if err := s.Write(ctx, "Scratch/a.md", []byte("2"), ModeCreate); err == nil {
		t.Fatal("expected exists error")
	}
}

func TestAppend(t *testing.T) {
	root := t.TempDir()
	s, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := s.Write(ctx, "a.md", []byte("hello"), ModeOverwrite); err != nil {
		t.Fatal(err)
	}
	if err := s.Write(ctx, "a.md", []byte(" world"), ModeAppend); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Read(ctx, "a.md")
	if string(got) != "hello world" {
		t.Fatalf("got %q", got)
	}
}
