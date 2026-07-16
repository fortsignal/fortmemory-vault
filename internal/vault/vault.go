// Package vault owns the Markdown filesystem under a single vault root.
//
// Responsibilities:
//   - Path jail (no escape via .. or symlinks outside root)
//   - Atomic write (temp + rename)
//   - Read / delete
//   - Refuse writes into .fortmemory/ from API callers
//
// Must NOT call FortSignal. Enforcement happens in higher layers before Write.
package vault

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Store is the filesystem view of one vault.
type Store struct {
	Root string // absolute
}

// New opens a vault root (must exist and be a directory). Root is cleaned to absolute.
func New(root string) (*Store, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	fi, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("vault root is not a directory: %s", abs)
	}
	// Evaluate symlinks on root so jail comparisons are stable.
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	return &Store{Root: abs}, nil
}

// WriteMode controls create/append/overwrite semantics.
type WriteMode string

const (
	ModeCreate    WriteMode = "create"
	ModeAppend    WriteMode = "append"
	ModeOverwrite WriteMode = "overwrite"
)

// Resolve maps a relative path to an absolute path under Root or errors.
// Does not require the target to exist. Rejects path traversal and .fortmemory writes
// when forWrite is true.
func (s *Store) Resolve(rel string, forWrite bool) (abs string, cleanRel string, err error) {
	cleanRel, err = cleanRelative(rel)
	if err != nil {
		return "", "", err
	}
	if forWrite && isFortMemoryInternal(cleanRel) {
		return "", "", fmt.Errorf("path_not_allowed: .fortmemory is reserved")
	}

	candidate := filepath.Join(s.Root, filepath.FromSlash(cleanRel))
	// Ensure candidate is still under root after Join (extra safety).
	if !underRoot(s.Root, candidate) {
		return "", "", fmt.Errorf("path_traversal: escapes vault root")
	}

	// If path exists (or a parent symlink chain), resolve and re-check jail.
	if resolved, err := filepath.EvalSymlinks(candidate); err == nil {
		if !underRoot(s.Root, resolved) {
			return "", "", fmt.Errorf("path_traversal: symlink escapes vault root")
		}
		return resolved, cleanRel, nil
	}

	// For non-existent files, eval symlinks on the nearest existing parent.
	parent := filepath.Dir(candidate)
	if resolvedParent, err := filepath.EvalSymlinks(parent); err == nil {
		base := filepath.Base(candidate)
		resolved := filepath.Join(resolvedParent, base)
		if !underRoot(s.Root, resolved) {
			return "", "", fmt.Errorf("path_traversal: parent escapes vault root")
		}
		return resolved, cleanRel, nil
	}

	return candidate, cleanRel, nil
}

// Read returns file bytes for a relative path.
func (s *Store) Read(ctx context.Context, rel string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	abs, _, err := s.Resolve(rel, false)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(abs)
}

// Write persists exact bytes after governance allow.
func (s *Store) Write(ctx context.Context, rel string, body []byte, mode WriteMode) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	abs, cleanRel, err := s.Resolve(rel, true)
	if err != nil {
		return err
	}
	_ = cleanRel

	exists := false
	if st, err := os.Stat(abs); err == nil && !st.IsDir() {
		exists = true
	} else if err == nil && st.IsDir() {
		return fmt.Errorf("path is a directory")
	}

	switch mode {
	case ModeCreate:
		if exists {
			return fmt.Errorf("file already exists (mode=create)")
		}
	case ModeAppend:
		if exists {
			prev, err := os.ReadFile(abs)
			if err != nil {
				return err
			}
			body = append(prev, body...)
		}
	case ModeOverwrite, "":
		// ok
	default:
		return fmt.Errorf("unknown write mode %q", mode)
	}

	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}

	// Atomic write: temp in same dir then rename.
	tmp, err := os.CreateTemp(filepath.Dir(abs), ".fm-write-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(body); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, abs)
}

// Delete removes a file.
func (s *Store) Delete(ctx context.Context, rel string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	abs, _, err := s.Resolve(rel, true)
	if err != nil {
		return err
	}
	return os.Remove(abs)
}

// WalkMarkdown lists .md files for indexing.
func (s *Store) WalkMarkdown(ctx context.Context, fn func(rel string, info os.FileInfo) error) error {
	return filepath.WalkDir(s.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		rel, err := filepath.Rel(s.Root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel == ".fortmemory" || strings.HasPrefix(rel, ".fortmemory/") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(rel), ".md") {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		return fn(rel, info)
	})
}

func cleanRelative(rel string) (string, error) {
	if rel == "" {
		return "", fmt.Errorf("path required")
	}
	// Disallow absolute paths.
	if filepath.IsAbs(rel) || strings.HasPrefix(rel, "/") || strings.HasPrefix(rel, "\\") {
		return "", fmt.Errorf("absolute path not allowed")
	}
	// Windows drive letter
	if len(rel) >= 2 && rel[1] == ':' {
		return "", fmt.Errorf("absolute path not allowed")
	}

	// Always normalize Windows separators — filepath.ToSlash is a no-op for '\' on Unix.
	raw := strings.ReplaceAll(rel, "\\", "/")
	raw = filepath.ToSlash(raw)
	for _, seg := range strings.Split(raw, "/") {
		if seg == ".." {
			return "", fmt.Errorf("path_traversal")
		}
	}

	cleaned := pathCleanSlash(raw)
	if cleaned == "." || cleaned == "" {
		return "", fmt.Errorf("path required")
	}
	if strings.HasPrefix(cleaned, "../") || cleaned == ".." {
		return "", fmt.Errorf("path_traversal")
	}
	return cleaned, nil
}

func pathCleanSlash(p string) string {
	// Use path-style clean without OS volume.
	parts := strings.Split(p, "/")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		switch part {
		case "", ".":
			continue
		case "..":
			// already rejected; keep defensive
			if len(out) > 0 {
				out = out[:len(out)-1]
			}
		default:
			out = append(out, part)
		}
	}
	return strings.Join(out, "/")
}

func isFortMemoryInternal(rel string) bool {
	return rel == ".fortmemory" || strings.HasPrefix(rel, ".fortmemory/")
}

func underRoot(root, candidate string) bool {
	root = filepath.Clean(root)
	candidate = filepath.Clean(candidate)
	sep := string(os.PathSeparator)
	if candidate == root {
		return true
	}
	return strings.HasPrefix(candidate, root+sep)
}
