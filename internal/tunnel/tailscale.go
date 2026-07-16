package tunnel

import (
	"fmt"
	"os/exec"
	"strings"
)

// DetectTailscale returns path to tailscale CLI or empty.
func DetectTailscale() string {
	p, err := exec.LookPath("tailscale")
	if err != nil {
		return ""
	}
	return p
}

// TailscaleStatusReport human status for check.
func TailscaleStatusReport(bind string, port int) string {
	if bind == "" {
		bind = "127.0.0.1"
	}
	if port == 0 {
		port = 7432
	}
	var b strings.Builder
	b.WriteString("FortMemory Tailscale helper (supported remote)\n")
	b.WriteString("==============================================\n")
	path := DetectTailscale()
	if path == "" {
		b.WriteString("tailscale:  NOT FOUND on PATH\n")
		b.WriteString("Install:    https://tailscale.com/download\n")
	} else {
		b.WriteString("tailscale:  " + path + "\n")
		if out, err := exec.Command(path, "version").CombinedOutput(); err == nil {
			line := strings.TrimSpace(string(out))
			if i := strings.Index(line, "\n"); i > 0 {
				line = line[:i]
			}
			b.WriteString("version:    " + line + "\n")
		}
		if out, err := exec.Command(path, "status", "--self").CombinedOutput(); err == nil {
			// first line often has hostname
			line := strings.TrimSpace(string(out))
			if line != "" {
				if i := strings.Index(line, "\n"); i > 0 {
					line = line[:i]
				}
				if len(line) > 120 {
					line = line[:120] + "…"
				}
				b.WriteString("status:     " + line + "\n")
			}
		}
	}
	fmt.Fprintf(&b, "fortmemory: http://%s:%d (keep on loopback)\n", bind, port)
	b.WriteString("docs:       docs/TAILSCALE.md · docs/REMOTE-ACCESS.md\n")
	return b.String()
}

// TailscaleServeCommand returns the recommended serve invocation.
func TailscaleServeCommand(bind string, port int) string {
	if bind == "" {
		bind = "127.0.0.1"
	}
	if port == 0 {
		port = 7432
	}
	return fmt.Sprintf("tailscale serve --bg http://%s:%d", bind, port)
}

// TailscaleGuide is the supported alternative (full text).
func TailscaleGuide(bind string, port int) string {
	if bind == "" {
		bind = "127.0.0.1"
	}
	if port == 0 {
		port = 7432
	}
	cmd := TailscaleServeCommand(bind, port)
	return fmt.Sprintf(`Tailscale — supported remote access
===================================

Use when agents/devices share your Tailscale tailnet.
Primary hostname/Access path remains Cloudflare Tunnel (docs/CLOUDFLARE-TUNNEL.md).
Overview: docs/REMOTE-ACCESS.md · docs/TAILSCALE.md

1. Install + login: https://tailscale.com/download
2. Keep fortmemory on loopback: %s:%d
3. Expose on tailnet:

   %s

4. From another tailnet device:

   curl -sS http://<magicdns>:%d/v1/health
   curl -sS -H "Authorization: Bearer fm_…" http://<magicdns>:%d/v1/…

5. Restrict access with Tailscale ACLs.

CLI:
   fortmemory tailscale check
   fortmemory tailscale print-serve

Security: tokens + FortSignal still required. Tailscale is transport only.
`, bind, port, cmd, port, port)
}
