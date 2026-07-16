package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/fortsignal/fortmemory-vault/internal/config"
	"github.com/fortsignal/fortmemory-vault/internal/tunnel"
)

func runTunnel(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: fortmemory tunnel <cloudflare|tailscale> …\nPrefer: fortmemory cloudflare …")
	}
	provider := args[0]
	switch provider {
	case "cloudflare", "cf":
		// Delegate to first-class plugin (supports subcommands + legacy flags).
		return runCloudflare(args[1:])
	case "tailscale", "ts":
		return runTunnelTailscale(args[1:])
	default:
		return fmt.Errorf("unknown tunnel provider %q (cloudflare|tailscale)", provider)
	}
}

// runTunnelCloudflare keeps legacy flag-style: fortmemory tunnel cloudflare --check
func runTunnelCloudflare(args []string) error {
	fs := flag.NewFlagSet("tunnel cloudflare", flag.ContinueOnError)
	check := fs.Bool("check", false, "detect cloudflared and print status")
	printInstall := fs.Bool("print-install", false, "print install instructions")
	writeConfig := fs.String("write-config", "", "write sample cloudflared YAML to path")
	hostname := fs.String("hostname", "memory.example.com", "public hostname for ingress")
	url := fs.String("url", "", "local FortMemory URL")
	tunnelName := fs.String("name", "fortmemory", "tunnel name")
	cfgPath := fs.String("config", "", "fortmemory config.toml (optional)")
	run := fs.Bool("run", false, "exec cloudflared tunnel run")
	quick := fs.Bool("quick", false, "temporary trycloud tunnel")
	install := fs.Bool("install", false, "install cloudflared to ~/.local/bin")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}

	bind, port := config.DefaultBind, config.DefaultPort
	localURL := *url
	if *cfgPath != "" || os.Getenv("FORTMEMORY_CONFIG") != "" {
		if p, err := config.Discover(*cfgPath); err == nil {
			if cfg, err := config.Load(p); err == nil {
				bind, port = cfg.Bind, cfg.Port
			}
		}
	}
	if localURL == "" {
		localURL = "http://" + bind + ":" + strconv.Itoa(port)
	}

	if !*check && !*printInstall && *writeConfig == "" && !*run && !*quick && !*install {
		*check = true
	}

	if *install {
		return cloudflareInstall(nil)
	}
	if *printInstall {
		fmt.Print(tunnel.CloudflareInstallGuide())
	}
	if *check {
		fmt.Print(tunnel.CloudflareStatusReport(bind, port))
		fmt.Println()
		fmt.Println("Prefer subcommands: fortmemory cloudflare install|check|config|quick|run")
	}
	if *writeConfig != "" {
		opt := tunnel.CloudflareOptions{
			Hostname:   *hostname,
			URL:        localURL,
			TunnelName: *tunnelName,
			ConfigPath: *writeConfig,
		}
		if err := tunnel.WriteCloudflareConfig(*writeConfig, opt); err != nil {
			return err
		}
		fmt.Printf("Wrote %s\n", *writeConfig)
	}
	if *quick {
		return tunnel.RunQuickTunnel(tunnel.DetectCloudflared(), localURL)
	}
	if *run {
		return tunnel.RunNamedTunnel(tunnel.DetectCloudflared(), *tunnelName, *writeConfig)
	}
	return nil
}

func runTunnelTailscale(args []string) error {
	// Delegate to first-class tailscale command.
	return runTailscale(args)
}
