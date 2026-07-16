package fortsignal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"strings"
	"unicode/utf8"
)

// Memory action names registered in FortSignal policies (lowercase match server-side).
const (
	ActionWrite  = "memory.write"
	ActionDelete = "memory.delete"
	ActionRead   = "memory.read"
	ActionSearch = "memory.search"
)

// MaxRecipientLen matches fortsignal-api validateRecipient.
const MaxRecipientLen = 256

// MaxMetadataJSONLen matches fortsignal-api validateMetadata serialized limit.
const MaxMetadataJSONLen = 2048

// ContentHash returns "sha256:" + hex of body bytes.
func ContentHash(body []byte) string {
	sum := sha256.Sum256(body)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// EncodeRecipient builds the FortSignal recipient field for a vault path.
//
// Canonical form: "{vaultId}/{relPath}" using forward slashes, cleaned.
// If over 256 runes/bytes budget, falls back to:
//
//	"{vaultId}/#/{sha256hex16}"
//
// and callers should put the real path in metadata["path"].
//
// Policy allowlists should use wildcards like "personal/Scratch/*".
func EncodeRecipient(vaultID, relPath string) (recipient string, metaPath string, err error) {
	vaultID = strings.TrimSpace(vaultID)
	if vaultID == "" {
		return "", "", fmt.Errorf("vaultId required")
	}
	raw := strings.ReplaceAll(relPath, "\\", "/")
	// Reject ".." before Clean — Clean("/../etc/passwd") becomes "/etc/passwd".
	for _, seg := range strings.Split(raw, "/") {
		if seg == ".." {
			return "", "", fmt.Errorf("path traversal")
		}
	}
	rel := path.Clean("/" + raw)
	rel = strings.TrimPrefix(rel, "/")
	if rel == "." || rel == "" {
		return "", "", fmt.Errorf("path required")
	}
	if strings.HasPrefix(rel, "/") {
		return "", "", fmt.Errorf("absolute path not allowed")
	}

	full := vaultID + "/" + rel
	if utf8.RuneCountInString(full) <= MaxRecipientLen && len(full) <= MaxRecipientLen {
		return full, "", nil
	}

	sum := sha256.Sum256([]byte(rel))
	short := vaultID + "/#/" + hex.EncodeToString(sum[:8])
	if len(short) > MaxRecipientLen {
		return "", "", fmt.Errorf("vaultId too long for recipient encoding")
	}
	return short, rel, nil
}

// WriteMetadata builds lean metadata for memory.write (must stay under 2048 JSON).
func WriteMetadata(vaultID, contentHash, mode, explicitPath string) map[string]any {
	m := map[string]any{
		"vaultId":     vaultID,
		"contentHash": contentHash,
		"mode":        mode,
	}
	if explicitPath != "" {
		m["path"] = explicitPath
	}
	return m
}
