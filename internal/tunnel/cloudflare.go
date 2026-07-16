package tunnel

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// CloudflareOptions configures generated cloudflared config.
type CloudflareOptions struct {
	Hostname   string // e.g. memory.example.com
	URL        string // e.g. http://127.0.0.1:7432
	TunnelName string // logical name, default fortmemory
	TunnelUUID string // optional; placeholder if empty
	ConfigPath string // where to write YAML
}

// DetectCloudflared returns path to cloudflared or empty.
// Checks PATH then ~/.local/bin/cloudflared (user install location).
func DetectCloudflared() string {
	if p, err := exec.LookPath("cloudflared"); err == nil {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	candidate := filepath.Join(home, ".local", "bin", "cloudflared")
	if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
		return candidate
	}
	return ""
}

// CloudflareVersion runs cloudflared --version.
func CloudflareVersion(bin string) string {
	if bin == "" {
		bin = DetectCloudflared()
	}
	if bin == "" {
		return ""
	}
	out, err := exec.Command(bin, "--version").CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// CloudflareInstallGuide returns install steps for common platforms.
func CloudflareInstallGuide() string {
	return strings.TrimSpace(`
Install Cloudflare Tunnel (cloudflared) — FortMemory plugin
==========================================================

Fastest (FortMemory helper — Linux/macOS user install to ~/.local/bin):

  fortmemory cloudflare install

Official docs:
  https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/installation/

Linux (system package example):
  curl -L --output cloudflared.deb \
    https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
  sudo dpkg -i cloudflared.deb

macOS:
  brew install cloudflared

Windows:
  winget install --id Cloudflare.cloudflared

Then authenticate and create a named tunnel:
  cloudflared tunnel login
  cloudflared tunnel create fortmemory

Or dogfood with a temporary quick tunnel (no DNS):
  fortmemory cloudflare quick

Full guide: docs/CLOUDFLARE-TUNNEL.md
`) + "\n"
}

// releaseArch maps Go arch to cloudflared release arch.
func releaseArch() (string, error) {
	switch runtime.GOARCH {
	case "amd64":
		return "amd64", nil
	case "arm64":
		return "arm64", nil
	default:
		return "", fmt.Errorf("unsupported arch %s (install cloudflared manually)", runtime.GOARCH)
	}
}

// releaseOS maps Go OS to cloudflared release OS name.
func releaseOS() (string, error) {
	switch runtime.GOOS {
	case "linux":
		return "linux", nil
	case "darwin":
		return "darwin", nil
	case "windows":
		return "windows", nil
	default:
		return "", fmt.Errorf("unsupported OS %s", runtime.GOOS)
	}
}

// InstallCloudflared downloads the latest cloudflared into ~/.local/bin (or destDir).
// Does not require root. Idempotent if already on PATH with a working binary (unless force).
func InstallCloudflared(destDir string, force bool) (installPath string, err error) {
	if !force {
		if existing := DetectCloudflared(); existing != "" {
			return existing, nil
		}
	}
	goos, err := releaseOS()
	if err != nil {
		return "", err
	}
	arch, err := releaseArch()
	if err != nil {
		return "", err
	}
	if destDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		destDir = filepath.Join(home, ".local", "bin")
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", err
	}
	name := "cloudflared"
	if goos == "windows" {
		name = "cloudflared.exe"
	}
	url := fmt.Sprintf(
		"https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-%s-%s",
		goos, arch,
	)
	if goos == "windows" {
		url += ".exe"
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("download cloudflared: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download cloudflared: HTTP %d from %s", resp.StatusCode, url)
	}

	dest := filepath.Join(destDir, name)
	tmp := dest + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return "", err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	return dest, nil
}

// CloudflareConfigYAML generates a minimal ingress config.
func CloudflareConfigYAML(opt CloudflareOptions) string {
	if opt.URL == "" {
		opt.URL = "http://127.0.0.1:7432"
	}
	if opt.TunnelName == "" {
		opt.TunnelName = "fortmemory"
	}
	if opt.Hostname == "" {
		opt.Hostname = "memory.example.com"
	}
	tunnelVal := "<TUNNEL_UUID>"
	if opt.TunnelUUID != "" {
		tunnelVal = opt.TunnelUUID
	}
	cred := filepath.Join("<HOME>", ".cloudflared", "<TUNNEL_UUID>.json")
	if opt.TunnelUUID != "" {
		if home, err := os.UserHomeDir(); err == nil {
			cred = filepath.Join(home, ".cloudflared", opt.TunnelUUID+".json")
		}
	}
	return fmt.Sprintf(`# Generated by fortmemory cloudflare (Cloudflare Tunnel plugin)
# Named tunnel setup:
#   1. cloudflared tunnel login
#   2. cloudflared tunnel create %s
#   3. Set tunnel UUID below if still a placeholder
#   4. cloudflared tunnel route dns %s %s
#   5. fortmemory cloudflare run --config %s
#
# SECURITY: Put Cloudflare Access / mTLS in front of this hostname.
# Keep fortmemory bound to 127.0.0.1. See docs/CLOUDFLARE-TUNNEL.md

tunnel: %s
credentials-file: %s

ingress:
  - hostname: %s
    service: %s
  - service: http_status:404
`, opt.TunnelName, opt.TunnelName, opt.Hostname, opt.ConfigPath,
		tunnelVal, cred, opt.Hostname, opt.URL)
}

// WriteCloudflareConfig writes YAML to path (0600).
func WriteCloudflareConfig(path string, opt CloudflareOptions) error {
	if path == "" {
		return fmt.Errorf("config path required")
	}
	opt.ConfigPath = path
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(CloudflareConfigYAML(opt)), 0o600)
}

// CloudflareStatusReport is human output for check.
func CloudflareStatusReport(bind string, port int) string {
	var b strings.Builder
	b.WriteString("FortMemory Cloudflare Tunnel plugin\n")
	b.WriteString("===================================\n")
	path := DetectCloudflared()
	if path == "" {
		b.WriteString("cloudflared: NOT FOUND\n")
		b.WriteString("Install:     fortmemory cloudflare install\n")
	} else {
		b.WriteString("cloudflared: " + path + "\n")
		if v := CloudflareVersion(path); v != "" {
			b.WriteString("version:     " + v + "\n")
		}
	}
	if bind == "" {
		bind = "127.0.0.1"
	}
	if port == 0 {
		port = 7432
	}
	fmt.Fprintf(&b, "fortmemory:  http://%s:%d (keep on loopback)\n", bind, port)
	b.WriteString("mTLS/Access: Cloudflare Zero Trust in front of hostname\n")
	b.WriteString("docs:        docs/CLOUDFLARE-TUNNEL.md\n")
	return b.String()
}

// RunQuickTunnel starts a temporary trycloud tunnel to localURL (blocks).
// Useful for dogfood without DNS: cloudflared tunnel --url http://127.0.0.1:7432
func RunQuickTunnel(bin, localURL string) error {
	if bin == "" {
		bin = DetectCloudflared()
	}
	if bin == "" {
		return fmt.Errorf("cloudflared not installed — run: fortmemory cloudflare install")
	}
	if localURL == "" {
		localURL = "http://127.0.0.1:7432"
	}
	cmd := exec.Command(bin, "tunnel", "--url", localURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// RunNamedTunnel runs a named tunnel (blocks).
func RunNamedTunnel(bin, name, configPath string) error {
	if bin == "" {
		bin = DetectCloudflared()
	}
	if bin == "" {
		return fmt.Errorf("cloudflared not installed — run: fortmemory cloudflare install")
	}
	if name == "" {
		name = "fortmemory"
	}
	args := []string{"tunnel"}
	if configPath != "" {
		args = append(args, "--config", configPath)
	}
	args = append(args, "run", name)
	cmd := exec.Command(bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}


