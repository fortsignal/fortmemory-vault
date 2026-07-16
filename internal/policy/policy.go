// Package policy evaluates FortMemory-local path/tag rules (ADR-010).
package policy

import (
	"path"
	"strings"

	"github.com/fortsignal/fortmemory-vault/internal/config"
)

// Decision is a local allow/deny.
type Decision struct {
	Allow  bool
	Reason string
}

// Engine evaluates local globs.
type Engine struct {
	cfg config.LocalPolicyConfig
}

// New builds an engine from config.
func New(cfg config.LocalPolicyConfig) *Engine {
	return &Engine{cfg: cfg}
}

// CheckWrite returns deny if path is outside allow_write or in deny_write.
func (e *Engine) CheckWrite(rel string) Decision {
	rel = normalize(rel)
	if rel == ".fortmemory" || strings.HasPrefix(rel, ".fortmemory/") {
		return Decision{Allow: false, Reason: "path_not_allowed"}
	}
	if matchAny(e.cfg.DenyWrite, rel) {
		return Decision{Allow: false, Reason: "path_not_allowed"}
	}
	// Empty allow list = allow all (subject to deny).
	if len(e.cfg.AllowWrite) > 0 && !matchAny(e.cfg.AllowWrite, rel) {
		return Decision{Allow: false, Reason: "path_not_allowed"}
	}
	return Decision{Allow: true}
}

// CheckRead returns deny for deny_read globs.
func (e *Engine) CheckRead(rel string) Decision {
	rel = normalize(rel)
	if matchAny(e.cfg.DenyRead, rel) {
		return Decision{Allow: false, Reason: "path_not_allowed"}
	}
	return Decision{Allow: true}
}

func normalize(rel string) string {
	rel = path.Clean("/" + strings.ReplaceAll(rel, "\\", "/"))
	return strings.TrimPrefix(rel, "/")
}

// matchAny supports:
//   - exact match
//   - trailing /** (prefix dir)
//   - trailing /* (single segment or prefix/)
//   - path.Match patterns
func matchAny(patterns []string, name string) bool {
	for _, p := range patterns {
		if matchGlob(p, name) {
			return true
		}
	}
	return false
}

func matchGlob(pattern, name string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return false
	}
	pattern = strings.ReplaceAll(pattern, "\\", "/")
	name = strings.ReplaceAll(name, "\\", "/")

	if pattern == name {
		return true
	}
	// dir/**
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		return name == prefix || strings.HasPrefix(name, prefix+"/")
	}
	// dir/*  → anything under dir/ (multi-segment, pragmatic)
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return name == prefix || strings.HasPrefix(name, prefix+"/")
	}
	// shell-style
	ok, err := path.Match(pattern, name)
	return err == nil && ok
}
