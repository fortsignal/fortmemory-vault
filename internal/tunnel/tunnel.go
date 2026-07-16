// Package tunnel provides OSS helpers for remote access.
//
// Founder preference (docs/FOUNDING-CONTEXT.md):
//   - Cloudflare Tunnel + mTLS: primary
//   - Tailscale: supported alternative
//
// We shell out to / guide cloudflared rather than embedding Cloudflare credentials
// in the FortMemory process (better security boundary).
package tunnel

import "fmt"

// Provider names.
const (
	ProviderCloudflare = "cloudflare" // primary
	ProviderTailscale  = "tailscale"  // supported
)

// Guide returns a short pointer; prefer CloudflareInstallGuide / TailscaleGuide.
func Guide(provider string, bind string, port int) (string, error) {
	switch provider {
	case ProviderCloudflare:
		return CloudflareStatusReport(bind, port) + "\n" + CloudflareInstallGuide(), nil
	case ProviderTailscale:
		return TailscaleGuide(bind, port), nil
	default:
		return "", fmt.Errorf("unknown tunnel provider %q (use cloudflare|tailscale)", provider)
	}
}
