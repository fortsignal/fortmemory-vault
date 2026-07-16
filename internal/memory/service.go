// Package memory is the application service coordinating vault, FortSignal,
// policy, index, and receipts for high-level memory operations.
package memory

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fortsignal/fortmemory-vault/internal/agent"
	"github.com/fortsignal/fortmemory-vault/internal/config"
	"github.com/fortsignal/fortmemory-vault/internal/fortsignal"
	"github.com/fortsignal/fortmemory-vault/internal/index"
	"github.com/fortsignal/fortmemory-vault/internal/policy"
	"github.com/fortsignal/fortmemory-vault/internal/receipts"
	"github.com/fortsignal/fortmemory-vault/internal/vault"
)

// Service is the memory use-case API.
type Service struct {
	Cfg        config.Config
	Vault      *vault.Store
	Index      index.Index // optional
	Receipts   receipts.Store
	Policy     *policy.Engine
	FortSignal fortsignal.Enforcer
	Signer     agent.Signer
	Signers    map[string]agent.Signer
}

// WriteInput is a governed write request.
type WriteInput struct {
	AgentID string
	Path    string
	Content []byte
	Mode    vault.WriteMode
}

// MutateResult mirrors FortSignal decision envelopes for API clients.
type MutateResult struct {
	Decision     string `json:"decision"`
	Reason       string `json:"reason,omitempty"`
	SignalID     string `json:"signalId,omitempty"`
	Path         string `json:"path,omitempty"`
	ContentHash  string `json:"contentHash,omitempty"`
	VerifiedBy   string `json:"verifiedBy,omitempty"`
	VerifiedAt   string `json:"verifiedAt,omitempty"`
	DelegationID string `json:"delegationId,omitempty"`
	PolicyID     string `json:"policyId,omitempty"`
}

func (s *Service) resolveSigner(agentID string) (agent.Signer, error) {
	if s.Signers != nil {
		if sig, ok := s.Signers[agentID]; ok && sig != nil {
			return sig, nil
		}
	}
	if s.Signer != nil {
		if agentID == "" || s.Signer.AgentID() == agentID {
			return s.Signer, nil
		}
		return nil, fmt.Errorf("signer agentId %q does not match %q", s.Signer.AgentID(), agentID)
	}
	return nil, fmt.Errorf("no signer for agent %q (register with --key)", agentID)
}

// Write performs local policy + FortSignal enforce + vault mutation.
func (s *Service) Write(ctx context.Context, in WriteInput) (*MutateResult, error) {
	if s.Vault == nil || s.FortSignal == nil {
		return nil, fmt.Errorf("memory service not configured")
	}
	if in.Mode == "" {
		in.Mode = vault.ModeOverwrite
	}
	signer, err := s.resolveSigner(in.AgentID)
	if err != nil {
		return nil, err
	}
	agentID := signer.AgentID()
	if in.AgentID != "" && in.AgentID != agentID {
		return nil, fmt.Errorf("agentId mismatch")
	}

	if s.Policy != nil {
		d := s.Policy.CheckWrite(in.Path)
		if !d.Allow {
			res := &MutateResult{Decision: "deny", Reason: d.Reason, Path: in.Path}
			s.appendReceipt(ctx, res, agentID, in.Path, "", fortsignal.ActionWrite)
			return res, nil
		}
	}

	if _, clean, err := s.Vault.Resolve(in.Path, true); err != nil {
		return &MutateResult{Decision: "deny", Reason: err.Error(), Path: in.Path}, nil
	} else {
		in.Path = clean
	}

	contentHash := fortsignal.ContentHash(in.Content)
	recipient, metaPath, err := fortsignal.EncodeRecipient(s.Cfg.VaultID, in.Path)
	if err != nil {
		return nil, fmt.Errorf("recipient: %w", err)
	}
	meta := fortsignal.WriteMetadata(s.Cfg.VaultID, contentHash, string(in.Mode), metaPath)

	start, err := s.FortSignal.ChallengeStart(ctx, fortsignal.ChallengeStartRequest{
		AgentID:   agentID,
		Action:    fortsignal.ActionWrite,
		Amount:    float64(len(in.Content)),
		Recipient: recipient,
		Source:    agentID,
		Metadata:  meta,
	})
	if err != nil {
		if s.Cfg.Security.FailClosedOnFortSignal {
			return nil, fmt.Errorf("fortsignal unavailable: %w", err)
		}
		return nil, err
	}
	if start.Decision == "deny" {
		res := &MutateResult{Decision: "deny", Reason: start.Reason, Path: in.Path}
		s.appendReceipt(ctx, res, agentID, in.Path, contentHash, fortsignal.ActionWrite)
		return res, nil
	}

	sig, err := signer.SignChallenge(start.Challenge)
	if err != nil {
		return nil, fmt.Errorf("sign challenge: %w", err)
	}

	verify, err := s.FortSignal.ChallengeVerify(ctx, fortsignal.ChallengeVerifyRequest{
		AgentID:   agentID,
		Challenge: start.Challenge,
		Signature: sig,
	})
	if err != nil {
		if s.Cfg.Security.FailClosedOnFortSignal {
			return nil, fmt.Errorf("fortsignal unavailable: %w", err)
		}
		return nil, err
	}

	res := &MutateResult{
		Decision:     verify.Decision,
		Reason:       verify.Reason,
		SignalID:     verify.SignalID,
		Path:         in.Path,
		ContentHash:  contentHash,
		VerifiedBy:   verify.VerifiedBy,
		VerifiedAt:   verify.VerifiedAt,
		DelegationID: verify.DelegationID,
		PolicyID:     verify.PolicyID,
	}

	if verify.Decision != "allow" {
		s.appendReceipt(ctx, res, agentID, in.Path, contentHash, fortsignal.ActionWrite)
		return res, nil
	}

	// Disk annotation after allow (not part of FortSignal contentHash).
	toWrite := annotateFrontmatter(in.Content, res.SignalID, agentID)
	if err := s.Vault.Write(ctx, in.Path, toWrite, in.Mode); err != nil {
		return nil, fmt.Errorf("vault write after allow: %w", err)
	}
	if s.Index != nil {
		_ = s.Index.Upsert(ctx, in.Path, toWrite, res.SignalID)
	}
	s.appendReceipt(ctx, res, agentID, in.Path, contentHash, fortsignal.ActionWrite)
	return res, nil
}

// Delete is FortSignal-gated file removal.
func (s *Service) Delete(ctx context.Context, agentID, path string) (*MutateResult, error) {
	if s.Vault == nil || s.FortSignal == nil {
		return nil, fmt.Errorf("memory service not configured")
	}
	signer, err := s.resolveSigner(agentID)
	if err != nil {
		return nil, err
	}
	agentID = signer.AgentID()

	if s.Policy != nil {
		d := s.Policy.CheckWrite(path) // same write allow-list controls delete scope
		if !d.Allow {
			res := &MutateResult{Decision: "deny", Reason: d.Reason, Path: path}
			s.appendReceipt(ctx, res, agentID, path, "", fortsignal.ActionDelete)
			return res, nil
		}
	}

	if _, clean, err := s.Vault.Resolve(path, true); err != nil {
		return &MutateResult{Decision: "deny", Reason: err.Error(), Path: path}, nil
	} else {
		path = clean
	}

	recipient, metaPath, err := fortsignal.EncodeRecipient(s.Cfg.VaultID, path)
	if err != nil {
		return nil, fmt.Errorf("recipient: %w", err)
	}
	meta := map[string]any{"vaultId": s.Cfg.VaultID}
	if metaPath != "" {
		meta["path"] = metaPath
	}

	start, err := s.FortSignal.ChallengeStart(ctx, fortsignal.ChallengeStartRequest{
		AgentID:   agentID,
		Action:    fortsignal.ActionDelete,
		Amount:    0,
		Recipient: recipient,
		Source:    agentID,
		Metadata:  meta,
	})
	if err != nil {
		if s.Cfg.Security.FailClosedOnFortSignal {
			return nil, fmt.Errorf("fortsignal unavailable: %w", err)
		}
		return nil, err
	}
	if start.Decision == "deny" {
		res := &MutateResult{Decision: "deny", Reason: start.Reason, Path: path}
		s.appendReceipt(ctx, res, agentID, path, "", fortsignal.ActionDelete)
		return res, nil
	}

	sig, err := signer.SignChallenge(start.Challenge)
	if err != nil {
		return nil, fmt.Errorf("sign challenge: %w", err)
	}
	verify, err := s.FortSignal.ChallengeVerify(ctx, fortsignal.ChallengeVerifyRequest{
		AgentID:   agentID,
		Challenge: start.Challenge,
		Signature: sig,
	})
	if err != nil {
		if s.Cfg.Security.FailClosedOnFortSignal {
			return nil, fmt.Errorf("fortsignal unavailable: %w", err)
		}
		return nil, err
	}

	res := &MutateResult{
		Decision:     verify.Decision,
		Reason:       verify.Reason,
		SignalID:     verify.SignalID,
		Path:         path,
		VerifiedBy:   verify.VerifiedBy,
		VerifiedAt:   verify.VerifiedAt,
		DelegationID: verify.DelegationID,
		PolicyID:     verify.PolicyID,
	}
	if verify.Decision != "allow" {
		s.appendReceipt(ctx, res, agentID, path, "", fortsignal.ActionDelete)
		return res, nil
	}
	if err := s.Vault.Delete(ctx, path); err != nil {
		return nil, fmt.Errorf("vault delete after allow: %w", err)
	}
	if s.Index != nil {
		_ = s.Index.Remove(ctx, path)
	}
	s.appendReceipt(ctx, res, agentID, path, "", fortsignal.ActionDelete)
	return res, nil
}

// Read returns file bytes after local read policy.
func (s *Service) Read(ctx context.Context, agentID, path string) (content []byte, contentHash string, err error) {
	_ = agentID
	if s.Policy != nil {
		d := s.Policy.CheckRead(path)
		if !d.Allow {
			return nil, "", fmt.Errorf("%s", d.Reason)
		}
	}
	body, err := s.Vault.Read(ctx, path)
	if err != nil {
		return nil, "", err
	}
	return body, fortsignal.ContentHash(body), nil
}

// Search queries the index after local read policy filtering.
func (s *Service) Search(ctx context.Context, agentID string, req index.SearchRequest) ([]index.Hit, error) {
	_ = agentID
	if s.Index == nil {
		return nil, fmt.Errorf("search index not configured")
	}
	hits, err := s.Index.Search(ctx, req)
	if err != nil {
		return nil, err
	}
	if s.Policy == nil {
		return hits, nil
	}
	out := make([]index.Hit, 0, len(hits))
	for _, h := range hits {
		if s.Policy.CheckRead(h.Path).Allow {
			out = append(out, h)
		}
	}
	return out, nil
}

// Reindex rebuilds FTS from vault Markdown files.
func (s *Service) Reindex(ctx context.Context) (int, error) {
	if s.Index == nil {
		return 0, fmt.Errorf("index not configured")
	}
	n := 0
	err := s.Vault.WalkMarkdown(ctx, func(rel string, _ os.FileInfo) error {
		body, err := s.Vault.Read(ctx, rel)
		if err != nil {
			return err
		}
		if err := s.Index.Upsert(ctx, rel, body, ""); err != nil {
			return err
		}
		n++
		return nil
	})
	return n, err
}

func (s *Service) appendReceipt(ctx context.Context, res *MutateResult, agentID, path, contentHash, action string) {
	if s.Receipts == nil || res == nil {
		return
	}
	rec := receipts.Record{
		SignalID:     res.SignalID,
		Decision:     res.Decision,
		Reason:       res.Reason,
		Action:       action,
		Path:         path,
		ContentHash:  contentHash,
		AgentID:      agentID,
		DelegationID: res.DelegationID,
		PolicyID:     res.PolicyID,
		VerifiedBy:   res.VerifiedBy,
		VaultID:      s.Cfg.VaultID,
	}
	if res.VerifiedAt != "" {
		if t, err := time.Parse(time.RFC3339Nano, res.VerifiedAt); err == nil {
			rec.VerifiedAt = t
		} else if t, err := time.Parse(time.RFC3339, res.VerifiedAt); err == nil {
			rec.VerifiedAt = t
		}
	}
	_ = s.Receipts.Append(ctx, rec)
}
