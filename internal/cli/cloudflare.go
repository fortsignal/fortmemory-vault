package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/fortsignal/fortmemory-vault/internal/config"
	"github.com/fortsignal/fortmemory-vault/internal/tunnel"
)

// runCloudflare is the first-class Cloudflare Tunnel plugin entrypoint.
//
//	fortmemory cloudflare <install|check|config|quick|run|help>
//
// Alias: fortmemory tunnel cloudflare …
func runCloudflare(args []string) error {
	if len(args) == 0 {
		return cloudflareHelp()
	}
	// Allow legacy flag-only form: fortmemory cloudflare --check
	if stringsHasPrefixDash(args[0]) {
		return runTunnelCloudflare(args)
	}

	switch args[0] {
	case "help", "-h", "--help":
		return cloudflareHelp()
	case "install":
		return cloudflareInstall(args[1:])
	case "check", "status":
		return cloudflareCheck(args[1:])
	case "config", "write-config":
		return cloudflareConfig(args[1:])
	case "quick":
		return cloudflareQuick(args[1:])
	case "run":
		return cloudflareRun(args[1:])
	case "print-install":
		fmt.Print(tunnel.CloudflareInstallGuide())
		return nil
	default:
		// Fall back to flag-style for backward compat
		return runTunnelCloudflare(args)
	}
}

func cloudflareHelp() error {
	fmt.Print(`FortMemory Cloudflare Tunnel plugin (primary remote access)

Usage:
  fortmemory cloudflare install [--force]     Download cloudflared to ~/.local/bin
  fortmemory cloudflare check                 Detect binary + print status
  fortmemory cloudflare config [flags]        Write cloudflared ingress YAML
  fortmemory cloudflare quick [--url URL]     Temporary trycloud tunnel (dogfood)
  fortmemory cloudflare run [flags]           Run named tunnel
  fortmemory cloudflare print-install         Show manual install notes

Config flags:
  --hostname memory.example.com
  --url http://127.0.0.1:7432
  --name fortmemory
  --write-config ~/.cloudflared/config-fortmemory.yml
  --tunnel-id <UUID>
  --fortmemory-config path/to/config.toml

Run flags:
  --name fortmemory
  --cf-config path/to/cloudflared.yml

Security: keep fortmemory on 127.0.0.1; put Cloudflare Access / mTLS in front.
Docs: docs/CLOUDFLARE-TUNNEL.md

Also: fortmemory tunnel cloudflare … (alias)
`)
	return nil
}

func cloudflareInstall(args []string) error {
	fs := flag.NewFlagSet("cloudflare install", flag.ContinueOnError)
	force := fs.Bool("force", false, "re-download even if cloudflared exists")
	dest := fs.String("dir", "", "install directory (default ~/.local/bin)")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	path, err := tunnel.InstallCloudflared(*dest, *force)
	if err != nil {
		return err
	}
	fmt.Printf("cloudflared ready: %s\n", path)
	if v := tunnel.CloudflareVersion(path); v != "" {
		fmt.Println(v)
	}
	fmt.Println()
	fmt.Println("Ensure ~/.local/bin is on your PATH, then:")
	fmt.Println("  fortmemory cloudflare check")
	fmt.Println("  cloudflared tunnel login")
	fmt.Println("  fortmemory cloudflare config --hostname memory.example.com")
	fmt.Println("  fortmemory cloudflare quick   # temp URL for dogfood")
	return nil
}

func cloudflareCheck(args []string) error {
	fs := flag.NewFlagSet("cloudflare check", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "fortmemory config.toml")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	bind, port := resolveFMListen(*cfgPath)
	fmt.Print(tunnel.CloudflareStatusReport(bind, port))
	if tunnel.DetectCloudflared() == "" {
		fmt.Println()
		fmt.Println("Next: fortmemory cloudflare install")
	} else {
		fmt.Println()
		fmt.Println("Dogfood temp URL:  fortmemory cloudflare quick")
		fmt.Println("Named tunnel:      fortmemory cloudflare config --hostname yours.example.com")
	}
	return nil
}

func cloudflareConfig(args []string) error {
	fs := flag.NewFlagSet("cloudflare config", flag.ContinueOnError)
	hostname := fs.String("hostname", "memory.example.com", "public hostname")
	url := fs.String("url", "", "local FortMemory URL")
	name := fs.String("name", "fortmemory", "tunnel name")
	tunnelID := fs.String("tunnel-id", "", "tunnel UUID (optional)")
	out := fs.String("write-config", "", "output YAML path")
	cfgPath := fs.String("config", "", "fortmemory config.toml")
	// alias used by legacy flag form
	legacyOut := fs.String("out", "", "alias for --write-config")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	path := *out
	if path == "" {
		path = *legacyOut
	}
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path = home + "/.cloudflared/config-fortmemory.yml"
	}
	localURL := *url
	if localURL == "" {
		bind, port := resolveFMListen(*cfgPath)
		localURL = "http://" + bind + ":" + strconv.Itoa(port)
	}
	opt := tunnel.CloudflareOptions{
		Hostname:   *hostname,
		URL:        localURL,
		TunnelName: *name,
		TunnelUUID: *tunnelID,
		ConfigPath: path,
	}
	if err := tunnel.WriteCloudflareConfig(path, opt); err != nil {
		return err
	}
	fmt.Printf("Wrote Cloudflare Tunnel config: %s\n", path)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. cloudflared tunnel login")
	fmt.Printf("  2. cloudflared tunnel create %s\n", *name)
	fmt.Println("  3. Put Tunnel UUID into the config (or re-run with --tunnel-id UUID)")
	fmt.Printf("  4. cloudflared tunnel route dns %s %s\n", *name, *hostname)
	fmt.Printf("  5. fortmemory cloudflare run --name %s --cf-config %s\n", *name, path)
	fmt.Println()
	fmt.Println("Enable Cloudflare Access / mTLS before production agents.")
	return nil
}

func cloudflareQuick(args []string) error {
	fs := flag.NewFlagSet("cloudflare quick", flag.ContinueOnError)
	url := fs.String("url", "", "local FortMemory URL")
	cfgPath := fs.String("config", "", "fortmemory config.toml")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	localURL := *url
	if localURL == "" {
		bind, port := resolveFMListen(*cfgPath)
		localURL = "http://" + bind + ":" + strconv.Itoa(port)
	}
	bin := tunnel.DetectCloudflared()
	if bin == "" {
		fmt.Fprintln(os.Stderr, "cloudflared missing — installing to ~/.local/bin …")
		p, err := tunnel.InstallCloudflared("", false)
		if err != nil {
			return err
		}
		bin = p
		fmt.Fprintf(os.Stderr, "installed: %s\n", bin)
	}
	fmt.Fprintf(os.Stderr, "Starting quick tunnel → %s\n", localURL)
	fmt.Fprintln(os.Stderr, "A trycloudflare.com URL will print below. Ctrl+C to stop.")
	fmt.Fprintln(os.Stderr, "NOTE: temporary; not for production. Use named tunnel + Access/mTLS for real agents.")
	return tunnel.RunQuickTunnel(bin, localURL)
}

func cloudflareRun(args []string) error {
	fs := flag.NewFlagSet("cloudflare run", flag.ContinueOnError)
	name := fs.String("name", "fortmemory", "tunnel name")
	cfConfig := fs.String("cf-config", "", "cloudflared config YAML")
	// also accept --config as cloudflared config for run (not fortmemory)
	cfg := fs.String("config", "", "alias for --cf-config when running tunnel")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	path := *cfConfig
	if path == "" {
		path = *cfg
	}
	bin := tunnel.DetectCloudflared()
	if bin == "" {
		return fmt.Errorf("cloudflared not found — run: fortmemory cloudflare install")
	}
	return tunnel.RunNamedTunnel(bin, *name, path)
}

func resolveFMListen(cfgPath string) (bind string, port int) {
	bind, port = config.DefaultBind, config.DefaultPort
	if p, err := config.Discover(cfgPath); err == nil {
		if cfg, err := config.Load(p); err == nil {
			return cfg.Bind, cfg.Port
		}
	}
	return bind, port
}

func stringsHasPrefixDash(s string) bool {
	return len(s) > 0 && s[0] == '-'
}
