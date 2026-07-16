// Package fortsignal is a thin HTTP client for the FortSignal enforcement API.
//
// See docs/FORTSIGNAL-INTEGRATION.md.
package fortsignal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const DefaultBaseURL = "https://api.fortsignal.com"

// Client talks to FortSignal.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// New constructs a client. apiKey is required for all calls.
func New(apiKey string, baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ChallengeStartRequest is POST /challenge/start for agents.
type ChallengeStartRequest struct {
	AgentID      string         `json:"agentId"`
	Action       string         `json:"action"`
	Amount       float64        `json:"amount"`
	Recipient    string         `json:"recipient"`
	Source       string         `json:"source,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	DelegationID string         `json:"delegationId,omitempty"`
}

// ChallengeStartResult is either a challenge to sign or an immediate deny.
type ChallengeStartResult struct {
	Decision     string `json:"decision,omitempty"`
	Reason       string `json:"reason,omitempty"`
	Challenge    string `json:"challenge,omitempty"`
	AgentID      string `json:"agentId,omitempty"`
	DelegationID string `json:"delegationId,omitempty"`
	ExpiresIn    int    `json:"expiresIn,omitempty"`
	// HTTPStatus is set by the client for diagnostics (not from JSON).
	HTTPStatus int `json:"-"`
}

// ChallengeVerifyRequest is POST /challenge/verify for agents.
type ChallengeVerifyRequest struct {
	AgentID   string `json:"agentId"`
	Challenge string `json:"challenge"`
	Signature string `json:"signature"`
}

// VerifyResult is the allow/deny envelope from /challenge/verify.
type VerifyResult struct {
	Decision     string         `json:"decision"`
	Reason       string         `json:"reason,omitempty"`
	SignalID     string         `json:"signalId,omitempty"`
	PolicyID     string         `json:"policyId,omitempty"`
	DelegationID string         `json:"delegationId,omitempty"`
	VerifiedBy   string         `json:"verifiedBy,omitempty"`
	VerifiedAt   string         `json:"verifiedAt,omitempty"`
	AgentID      string         `json:"agentId,omitempty"`
	Action       string         `json:"action,omitempty"`
	Amount       float64        `json:"amount,omitempty"`
	Recipient    string         `json:"recipient,omitempty"`
	Source       string         `json:"source,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	ScopeApplied string         `json:"scope_applied,omitempty"`
	HTTPStatus   int            `json:"-"`
}

// AgentRegisterRequest is POST /agent/register.
type AgentRegisterRequest struct {
	AgentID   string `json:"agentId"`
	PublicKey string `json:"publicKey"`
}

// AgentRegisterResponse is the success body.
type AgentRegisterResponse struct {
	Status       string `json:"status"`
	AgentID      string `json:"agentId"`
	RegisteredAt string `json:"registeredAt"`
}

// Enforcer is the subset FortMemory needs at runtime.
type Enforcer interface {
	ChallengeStart(ctx context.Context, req ChallengeStartRequest) (*ChallengeStartResult, error)
	ChallengeVerify(ctx context.Context, req ChallengeVerifyRequest) (*VerifyResult, error)
}

// ChallengeStart calls FortSignal agent challenge start.
// HTTP 403 with {decision:deny, reason} is a successful governance outcome, not a transport error.
func (c *Client) ChallengeStart(ctx context.Context, req ChallengeStartRequest) (*ChallengeStartResult, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("fortsignal: API key required")
	}
	var out ChallengeStartResult
	status, err := c.doJSON(ctx, http.MethodPost, "/challenge/start", req, &out)
	out.HTTPStatus = status
	if err != nil {
		// Still return parsed body for 403 denials.
		if status == http.StatusForbidden && out.Decision == "deny" {
			return &out, nil
		}
		// Some deny paths may omit decision but include reason.
		if status == http.StatusForbidden && out.Reason != "" {
			if out.Decision == "" {
				out.Decision = "deny"
			}
			return &out, nil
		}
		return nil, err
	}
	// 200 with decision deny (defensive)
	if out.Decision == "deny" {
		return &out, nil
	}
	if out.Challenge == "" {
		return nil, fmt.Errorf("fortsignal: challenge/start missing challenge (status %d)", status)
	}
	return &out, nil
}

// ChallengeVerify calls FortSignal agent challenge verify.
func (c *Client) ChallengeVerify(ctx context.Context, req ChallengeVerifyRequest) (*VerifyResult, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("fortsignal: API key required")
	}
	var out VerifyResult
	status, err := c.doJSON(ctx, http.MethodPost, "/challenge/verify", req, &out)
	out.HTTPStatus = status
	if err != nil {
		if status == http.StatusForbidden && (out.Decision == "deny" || out.Reason != "") {
			if out.Decision == "" {
				out.Decision = "deny"
			}
			return &out, nil
		}
		return nil, err
	}
	if out.Decision == "" {
		return nil, fmt.Errorf("fortsignal: challenge/verify missing decision (status %d)", status)
	}
	return &out, nil
}

// RegisterAgent registers an Ed25519 public key.
func (c *Client) RegisterAgent(ctx context.Context, req AgentRegisterRequest) (*AgentRegisterResponse, error) {
	var out AgentRegisterResponse
	status, err := c.doJSON(ctx, http.MethodPost, "/agent/register", req, &out)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK && status != http.StatusCreated {
		return nil, fmt.Errorf("fortsignal: agent/register status %d", status)
	}
	return &out, nil
}

// AgentListResponse is GET /agent/list.
type AgentListResponse struct {
	Agents []AgentListItem `json:"agents"`
}

// AgentListItem is one registered agent summary.
type AgentListItem struct {
	AgentID      string `json:"agentId"`
	RegisteredAt string `json:"registeredAt,omitempty"`
	RotatedAt    string `json:"rotatedAt,omitempty"`
	Delegation   *struct {
		DelegationID string `json:"delegationId"`
		PolicyID     string `json:"policyId"`
		IssuedAt     string `json:"issuedAt,omitempty"`
		ExpiresAt    string `json:"expiresAt,omitempty"`
	} `json:"delegation,omitempty"`
}

// ListAgents calls GET /agent/list.
func (c *Client) ListAgents(ctx context.Context) (*AgentListResponse, error) {
	var out AgentListResponse
	status, err := c.doJSON(ctx, http.MethodGet, "/agent/list", nil, &out)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("fortsignal: agent/list status %d", status)
	}
	return &out, nil
}

// Ping does a lightweight authenticated call (agent list) to validate API key + reachability.
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.ListAgents(ctx)
	return err
}

func (c *Client) doJSON(ctx context.Context, method, path string, in any, out any) (int, error) {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return 0, err
		}
		body = bytes.NewReader(b)
	}
	url := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("fortsignal: request failed: %w", err)
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return res.StatusCode, err
	}

	if len(raw) > 0 && out != nil {
		// Tolerate empty body.
		if err := json.Unmarshal(raw, out); err != nil {
			// Non-JSON error page
			if res.StatusCode < 200 || res.StatusCode >= 300 {
				return res.StatusCode, fmt.Errorf("fortsignal: HTTP %d: %s", res.StatusCode, truncate(string(raw), 200))
			}
			return res.StatusCode, fmt.Errorf("fortsignal: decode response: %w", err)
		}
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		// Try to surface API error field.
		var er struct {
			Error  string `json:"error"`
			Reason string `json:"reason"`
		}
		_ = json.Unmarshal(raw, &er)
		msg := er.Error
		if msg == "" {
			msg = er.Reason
		}
		if msg == "" {
			msg = truncate(string(raw), 200)
		}
		return res.StatusCode, fmt.Errorf("fortsignal: HTTP %d: %s", res.StatusCode, msg)
	}
	return res.StatusCode, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
