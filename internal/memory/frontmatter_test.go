package memory

import (
	"strings"
	"testing"
)

func TestAnnotatePrepend(t *testing.T) {
	out := string(annotateFrontmatter([]byte("# Hi\n"), "sig-1", "agent-1"))
	if !strings.Contains(out, "last_signal_id: sig-1") {
		t.Fatal(out)
	}
	if !strings.Contains(out, "# Hi") {
		t.Fatal(out)
	}
	if !strings.HasPrefix(out, "---\n") {
		t.Fatal(out)
	}
}

func TestAnnotateMerge(t *testing.T) {
	in := "---\ntitle: Note\ntags: [a]\n---\n\nBody\n"
	out := string(annotateFrontmatter([]byte(in), "sig-2", "a1"))
	if !strings.Contains(out, "title: Note") {
		t.Fatal(out)
	}
	if !strings.Contains(out, "last_signal_id: sig-2") {
		t.Fatal(out)
	}
	if !strings.Contains(out, "Body") {
		t.Fatal(out)
	}
}
