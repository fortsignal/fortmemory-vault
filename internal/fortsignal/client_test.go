package fortsignal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChallengeStartAllow(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/challenge/start" {
			t.Fatalf("path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("auth %q", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"challenge":    "abc",
			"agentId":      "a1",
			"delegationId": "d1",
			"expiresIn":    60,
		})
	}))
	defer srv.Close()

	c := New("test-key", srv.URL)
	out, err := c.ChallengeStart(context.Background(), ChallengeStartRequest{
		AgentID:   "a1",
		Action:    ActionWrite,
		Amount:    10,
		Recipient: "personal/Scratch/x.md",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Challenge != "abc" || out.Decision == "deny" {
		t.Fatalf("%+v", out)
	}
}

func TestChallengeStartDeny403(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"decision": "deny",
			"reason":   "action_not_allowed",
		})
	}))
	defer srv.Close()

	c := New("test-key", srv.URL)
	out, err := c.ChallengeStart(context.Background(), ChallengeStartRequest{
		AgentID:   "a1",
		Action:    ActionWrite,
		Recipient: "personal/Private/x.md",
	})
	if err != nil {
		t.Fatalf("403 deny should not be transport error: %v", err)
	}
	if out.Decision != "deny" || out.Reason != "action_not_allowed" {
		t.Fatalf("%+v", out)
	}
}

func TestPingAcceptsPolicyDeny(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/challenge/start" {
			t.Fatalf("path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("auth")
		}
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"decision": "deny",
			"reason":   "delegation_invalid",
		})
	}))
	defer srv.Close()

	c := New("test-key", srv.URL)
	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("policy deny should count as ping success: %v", err)
	}
}

func TestPingRejectsUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "unauthorized"})
	}))
	defer srv.Close()

	c := New("bad-key", srv.URL)
	if err := c.Ping(context.Background()); err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestChallengeVerifyAllow(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/challenge/verify" {
			t.Fatalf("path %s", r.URL.Path)
		}
		// FortSignal returns metadata as a JSON string, not an object.
		_ = json.NewEncoder(w).Encode(map[string]any{
			"decision":   "allow",
			"signalId":   "sig-1",
			"verifiedBy": "agent",
			"verifiedAt": "2026-07-16T00:00:00Z",
			"agentId":    "a1",
			"metadata":   `{"vaultId":"jeff","contentHash":"sha256:abc"}`,
		})
	}))
	defer srv.Close()

	c := New("test-key", srv.URL)
	out, err := c.ChallengeVerify(context.Background(), ChallengeVerifyRequest{
		AgentID:   "a1",
		Challenge: "abc",
		Signature: "sig",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Decision != "allow" || out.SignalID != "sig-1" {
		t.Fatalf("%+v", out)
	}
}
