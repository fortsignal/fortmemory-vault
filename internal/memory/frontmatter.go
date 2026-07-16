package memory

import (
	"fmt"
	"strings"
	"time"
)

// annotateFrontmatter merges governance fields into Markdown after allow.
// The FortSignal contentHash is computed on the agent payload *before* this runs.
func annotateFrontmatter(body []byte, signalID, agentID string) []byte {
	if signalID == "" {
		return body
	}
	now := time.Now().UTC().Format(time.RFC3339)
	text := string(body)
	// Normalize newlines
	text = strings.ReplaceAll(text, "\r\n", "\n")

	fields := map[string]string{
		"last_signal_id": signalID,
		"updated":        now,
	}
	if agentID != "" {
		fields["last_agent_id"] = agentID
	}

	if strings.HasPrefix(text, "---\n") {
		end := strings.Index(text[4:], "\n---")
		if end >= 0 {
			end += 4 // position of \n--- relative to start... actually end is index in text[4:]
			// full close index:
			closeIdx := 4 + end // points at \n of \n---
			// find end of closing ---
			rest := text[closeIdx:]
			// rest starts with \n---
			lineEnd := strings.Index(rest[1:], "\n")
			var after string
			var fmBlock string
			if lineEnd < 0 {
				fmBlock = text[:len(text)]
				after = ""
			} else {
				// closeIdx is at '\n' before ---
				// better parse properly
				fmBlock, after = splitExistingFrontmatter(text)
			}
			_ = lineEnd
			if fmBlock != "" {
				merged := mergeYAMLFrontmatter(fmBlock, fields)
				if after == "" && !strings.HasSuffix(merged, "\n") {
					return []byte(merged + "\n")
				}
				return []byte(merged + after)
			}
		}
	}

	// No frontmatter — prepend
	var b strings.Builder
	b.WriteString("---\n")
	for _, k := range []string{"last_signal_id", "last_agent_id", "updated"} {
		if v, ok := fields[k]; ok && v != "" {
			fmt.Fprintf(&b, "%s: %s\n", k, v)
		}
	}
	b.WriteString("---\n")
	if text != "" && !strings.HasPrefix(text, "\n") {
		// keep body as-is
	}
	b.WriteString(text)
	if !strings.HasSuffix(text, "\n") && text != "" {
		b.WriteByte('\n')
	}
	return []byte(b.String())
}

func splitExistingFrontmatter(text string) (fmWithDelimiters string, after string) {
	if !strings.HasPrefix(text, "---\n") {
		return "", text
	}
	// find closing \n---\n or \n--- at EOF
	rest := text[4:]
	idx := strings.Index(rest, "\n---\n")
	if idx >= 0 {
		// fm is --- + rest[:idx] + \n---\n
		fm := text[:4+idx+5] // --- + content + \n---\n
		after = text[4+idx+5:]
		return fm, after
	}
	idx = strings.Index(rest, "\n---")
	if idx >= 0 && idx+4 == len(rest) {
		return text, ""
	}
	if idx >= 0 && strings.HasPrefix(rest[idx:], "\n---") {
		// \n--- at end without trailing newline
		fm := text[:4+idx+4]
		after = text[4+idx+4:]
		return fm, after
	}
	return "", text
}

func mergeYAMLFrontmatter(fmBlock string, fields map[string]string) string {
	// fmBlock includes --- wrappers
	inner := fmBlock
	if strings.HasPrefix(inner, "---\n") {
		inner = inner[4:]
	}
	inner = strings.TrimSuffix(inner, "\n---\n")
	inner = strings.TrimSuffix(inner, "\n---")
	inner = strings.TrimSuffix(inner, "---")

	lines := strings.Split(inner, "\n")
	seen := map[string]bool{}
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			out = append(out, line)
			continue
		}
		key := strings.SplitN(trimmed, ":", 2)[0]
		key = strings.TrimSpace(key)
		if v, ok := fields[key]; ok {
			out = append(out, fmt.Sprintf("%s: %s", key, v))
			seen[key] = true
			continue
		}
		out = append(out, line)
	}
	for _, k := range []string{"last_signal_id", "last_agent_id", "updated"} {
		if seen[k] {
			continue
		}
		if v, ok := fields[k]; ok && v != "" {
			out = append(out, fmt.Sprintf("%s: %s", k, v))
		}
	}
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(strings.Join(out, "\n"))
	if len(out) > 0 && out[len(out)-1] != "" {
		b.WriteByte('\n')
	}
	b.WriteString("---\n")
	return b.String()
}
