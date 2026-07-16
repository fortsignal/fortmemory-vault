// Package watcher reindexes the vault when Markdown files change on disk
// (e.g. human edits in Obsidian).
package watcher

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/fortsignal/fortmemory-vault/internal/index"
	"github.com/fortsignal/fortmemory-vault/internal/vault"
)

// Watcher debounces FS events and updates the search index.
type Watcher struct {
	vault *vault.Store
	index index.Index
	log   *slog.Logger

	debounce time.Duration
	mu       sync.Mutex
	pending  map[string]struct{} // rel paths to refresh; empty path means full reindex not used
	timer    *time.Timer
}

// New creates a watcher (not started).
func New(v *vault.Store, idx index.Index, log *slog.Logger) *Watcher {
	if log == nil {
		log = slog.Default()
	}
	return &Watcher{
		vault:    v,
		index:    idx,
		log:      log,
		debounce: 400 * time.Millisecond,
		pending:  map[string]struct{}{},
	}
}

// Start watches vault root recursively until ctx is cancelled.
func (w *Watcher) Start(ctx context.Context) error {
	if w.index == nil || w.vault == nil {
		return nil
	}
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	root := w.vault.Root
	if err := addRecursive(fw, root); err != nil {
		_ = fw.Close()
		return err
	}
	w.log.Info("vault watcher started", "root", root)

	go func() {
		defer fw.Close()
		for {
			select {
			case <-ctx.Done():
				w.flush(context.Background())
				return
			case err, ok := <-fw.Errors:
				if !ok {
					return
				}
				w.log.Warn("watcher error", "err", err)
			case ev, ok := <-fw.Events:
				if !ok {
					return
				}
				w.handleEvent(fw, root, ev)
			}
		}
	}()
	return nil
}

func (w *Watcher) handleEvent(fw *fsnotify.Watcher, root string, ev fsnotify.Event) {
	// Ignore our own engine dir
	rel, err := filepath.Rel(root, ev.Name)
	if err != nil {
		return
	}
	rel = filepath.ToSlash(rel)
	if rel == ".fortmemory" || strings.HasPrefix(rel, ".fortmemory/") {
		return
	}

	// New directories: watch them
	if ev.Has(fsnotify.Create) {
		if st, err := os.Stat(ev.Name); err == nil && st.IsDir() {
			_ = addRecursive(fw, ev.Name)
			return
		}
	}

	if !isMarkdown(rel) && !ev.Has(fsnotify.Remove) && !ev.Has(fsnotify.Rename) {
		return
	}
	// For remove/rename of non-md, skip
	if !isMarkdown(rel) {
		return
	}

	w.schedule(rel)
}

func (w *Watcher) schedule(rel string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.pending[rel] = struct{}{}
	if w.timer != nil {
		w.timer.Stop()
	}
	w.timer = time.AfterFunc(w.debounce, func() {
		w.flush(context.Background())
	})
}

func (w *Watcher) flush(ctx context.Context) {
	w.mu.Lock()
	batch := w.pending
	w.pending = map[string]struct{}{}
	w.timer = nil
	w.mu.Unlock()
	if len(batch) == 0 {
		return
	}
	for rel := range batch {
		w.refreshOne(ctx, rel)
	}
}

func (w *Watcher) refreshOne(ctx context.Context, rel string) {
	body, err := w.vault.Read(ctx, rel)
	if err != nil {
		// deleted or missing
		if err := w.index.Remove(ctx, rel); err != nil {
			w.log.Debug("index remove", "path", rel, "err", err)
		} else {
			w.log.Info("index removed", "path", rel)
		}
		return
	}
	if err := w.index.Upsert(ctx, rel, body, ""); err != nil {
		w.log.Warn("index upsert", "path", rel, "err", err)
		return
	}
	w.log.Info("index refreshed", "path", rel)
}

func isMarkdown(rel string) bool {
	return strings.HasSuffix(strings.ToLower(rel), ".md")
}

func addRecursive(fw *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if base == ".fortmemory" || base == ".git" || base == "node_modules" {
			return fs.SkipDir
		}
		return fw.Add(path)
	})
}
