// Package server exposes the FortMemory HTTP API.
//
//	GET  /v1/health   (no auth)
//	POST /v1/search
//	GET  /v1/read
//	POST /v1/write
//	POST /v1/delete
//	GET  /v1/receipts
//	GET  /            (dashboard)
//
// Bind 127.0.0.1 by default.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/fortsignal/fortmemory-vault/internal/agent"
	"github.com/fortsignal/fortmemory-vault/internal/auth"
	"github.com/fortsignal/fortmemory-vault/internal/config"
	"github.com/fortsignal/fortmemory-vault/internal/index"
	"github.com/fortsignal/fortmemory-vault/internal/memory"
	"github.com/fortsignal/fortmemory-vault/internal/receipts"
	"github.com/fortsignal/fortmemory-vault/internal/vault"
)

// Deps wires subsystems.
type Deps struct {
	Config   config.Config
	Version  string
	Vault    *vault.Store
	Index    index.Index
	Receipts receipts.Store
	Agents   agent.Store
	Memory   *memory.Service
}

// Server is the HTTP surface.
type Server struct {
	deps Deps
	http *http.Server
	log  *slog.Logger
}

// New constructs a server.
func New(deps Deps) *Server {
	return &Server{
		deps: deps,
		log:  slog.Default(),
	}
}

// Handler returns the root mux.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/health", s.handleHealth)

	authed := auth.Middleware(s.deps.Agents)
	mux.Handle("GET /v1/read", authed(http.HandlerFunc(s.handleRead)))
	mux.Handle("POST /v1/search", authed(http.HandlerFunc(s.handleSearch)))
	mux.Handle("POST /v1/write", authed(http.HandlerFunc(s.handleWrite)))
	mux.Handle("POST /v1/delete", authed(http.HandlerFunc(s.handleDelete)))
	mux.Handle("GET /v1/receipts", authed(http.HandlerFunc(s.handleReceipts)))
	mux.Handle("GET /v1/agents", authed(http.HandlerFunc(s.handleAgents)))

	// Dashboard (PRODUCT-SURFACE: Home / Search / Activity / Settings)
	dash := dashboardHandler()
	mux.Handle("GET /{$}", dash)
	mux.Handle("GET /index.html", dash)

	return mux
}

// ListenAndServe starts the listener.
func (s *Server) ListenAndServe() error {
	bind := s.deps.Config.Bind
	if bind == "" {
		bind = config.DefaultBind
	}
	port := s.deps.Config.Port
	if port == 0 {
		port = config.DefaultPort
	}
	addr := fmt.Sprintf("%s:%d", bind, port)
	s.http = &http.Server{
		Addr:              addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	s.log.Info("fortmemory listening", "addr", addr, "vault", s.deps.Config.VaultPath)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return s.http.Serve(ln)
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.http == nil {
		return nil
	}
	return s.http.Shutdown(ctx)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	files, embedPending := 0, 0
	if s.deps.Index != nil {
		files, embedPending, _ = s.deps.Index.Stats(r.Context())
	}
	ver := s.deps.Version
	if ver == "" {
		ver = "0.0.0-dev"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"version":   ver,
		"vaultId":   s.deps.Config.VaultID,
		"vaultPath": s.deps.Config.VaultPath,
		"listen": map[string]any{
			"bind": s.deps.Config.Bind,
			"port": s.deps.Config.Port,
		},
		"index": map[string]any{
			"files":        files,
			"embedPending": embedPending,
		},
		"fortsignal": map[string]any{
			"configured": os.Getenv(s.deps.Config.FortSignal.APIKeyEnv) != "" || os.Getenv("FORTSIGNAL_API_KEY") != "",
		},
		// Hints only — not auto-configured tunnels
		"remoteAccess": map[string]any{
			"default":     "localhost",
			"tailscale":   "fortmemory tailscale print-serve",
			"cloudflare":  "fortmemory cloudflare check",
			"docs":        "docs/REMOTE-ACCESS.md",
		},
	})
}

func (s *Server) handleRead(w http.ResponseWriter, r *http.Request) {
	rec, err := auth.AgentFromContext(r.Context())
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	path := r.URL.Query().Get("path")
	if path == "" {
		writeErr(w, http.StatusBadRequest, "path required")
		return
	}
	body, hash, err := s.deps.Memory.Read(r.Context(), rec.AgentID, path)
	if err != nil {
		if os.IsNotExist(err) {
			writeErr(w, http.StatusNotFound, "not_found")
			return
		}
		writeErr(w, http.StatusForbidden, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"path":        path,
		"content":     string(body),
		"contentHash": hash,
	})
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	rec, err := auth.AgentFromContext(r.Context())
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req index.SearchRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	hits, err := s.deps.Memory.Search(r.Context(), rec.AgentID, req)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": hits})
}

func (s *Server) handleWrite(w http.ResponseWriter, r *http.Request) {
	rec, err := auth.AgentFromContext(r.Context())
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
		Mode    string `json:"mode"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 8<<20)).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Path == "" {
		writeErr(w, http.StatusBadRequest, "path required")
		return
	}
	mode := vault.WriteMode(req.Mode)
	if mode == "" {
		mode = vault.ModeOverwrite
	}
	res, err := s.deps.Memory.Write(r.Context(), memory.WriteInput{
		AgentID: rec.AgentID,
		Path:    req.Path,
		Content: []byte(req.Content),
		Mode:    mode,
	})
	if err != nil {
		// Infrastructure / FortSignal down
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"error":   "fortsignal_unavailable",
			"message": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.AgentFromContext(r.Context()); err != nil {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if s.deps.Agents == nil {
		writeJSON(w, http.StatusOK, map[string]any{"agents": []any{}})
		return
	}
	list, err := s.deps.Agents.List(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	type row struct {
		AgentID    string `json:"agentId"`
		HasKey     bool   `json:"hasKey"`
		KeyPathSet bool   `json:"keyPathSet"`
	}
	out := make([]row, 0, len(list))
	for _, a := range list {
		out = append(out, row{
			AgentID:    a.AgentID,
			HasKey:     a.KeyPath != "",
			KeyPathSet: a.KeyPath != "",
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"agents": out})
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	rec, err := auth.AgentFromContext(r.Context())
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Path == "" {
		writeErr(w, http.StatusBadRequest, "path required")
		return
	}
	res, err := s.deps.Memory.Delete(r.Context(), rec.AgentID, req.Path)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"error":   "fortsignal_unavailable",
			"message": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) handleReceipts(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.AgentFromContext(r.Context()); err != nil {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if s.deps.Receipts == nil {
		writeJSON(w, http.StatusOK, map[string]any{"records": []any{}})
		return
	}
	q := receipts.Query{
		Limit:      50,
		Action:     r.URL.Query().Get("action"),
		Decision:   r.URL.Query().Get("decision"),
		PathPrefix: r.URL.Query().Get("pathPrefix"),
		AgentID:    r.URL.Query().Get("agentId"),
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			q.Limit = n
		}
	}
	recs, err := s.deps.Receipts.List(r.Context(), q)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"records": recs})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"error": msg})
}
