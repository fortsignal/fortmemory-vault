// Package auth provides HTTP bearer authentication for local API tokens.
//
// These tokens are FortMemory-local. They are NOT FortSignal API keys.
// FortSignal API keys stay in process env and never go to agent clients.
package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/fortsignal/fortmemory-vault/internal/agent"
)

type ctxKey int

const agentKey ctxKey = 1

// Middleware validates Authorization: Bearer <token> via agent.Store.
func Middleware(store agent.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get("Authorization")
			if raw == "" || !strings.HasPrefix(strings.ToLower(raw), "bearer ") {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			token := strings.TrimSpace(raw[7:])
			rec, err := store.LookupByToken(r.Context(), token)
			if err != nil || rec == nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), agentKey, rec)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AgentFromContext returns the authenticated agent record.
func AgentFromContext(ctx context.Context) (*agent.Record, error) {
	rec, ok := ctx.Value(agentKey).(*agent.Record)
	if !ok || rec == nil {
		return nil, fmt.Errorf("no agent in context")
	}
	return rec, nil
}
