package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/fortsignal/fortmemory-vault/internal/tunnel"
)

// runTailscale is the first-class Tailscale helper (supported remote option).
//
//	fortmemory tailscale [check|print-serve|help]
func runTailscale(args []string) error {
	if len(args) == 0 {
		return tailscaleDefault(nil)
	}
	if stringsHasPrefixDash(args[0]) {
		return runTunnelTailscale(args)
	}
	switch args[0] {
	case "help", "-h", "--help":
		fmt.Print(`FortMemory Tailscale helper (supported remote access)

Usage:
  fortmemory tailscale                 Status + guide
  fortmemory tailscale check           Detect tailscale CLI
  fortmemory tailscale print-serve     Exact "tailscale serve" command
  fortmemory tailscale help

Flags:
  --config path    fortmemory config.toml (for bind/port)

Keep fortmemory on 127.0.0.1. Primary hostname/Access path: Cloudflare Tunnel.
Docs: docs/TAILSCALE.md · docs/REMOTE-ACCESS.md
`)
		return nil
	case "check", "status":
		return tailscaleCheck(args[1:])
	case "print-serve", "serve-cmd", "print-guide":
		return tailscalePrintServe(args[1:])
	default:
		return runTunnelTailscale(args)
	}
}

func tailscaleDefault(args []string) error {
	fs := flag.NewFlagSet("tailscale", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "fortmemory config.toml")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	bind, port := resolveFMListen(*cfgPath)
	fmt.Print(tunnel.TailscaleStatusReport(bind, port))
	fmt.Println()
	fmt.Print(tunnel.TailscaleGuide(bind, port))
	return nil
}

func tailscaleCheck(args []string) error {
	fs := flag.NewFlagSet("tailscale check", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "fortmemory config.toml")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	bind, port := resolveFMListen(*cfgPath)
	fmt.Print(tunnel.TailscaleStatusReport(bind, port))
	if tunnel.DetectTailscale() == "" {
		fmt.Println()
		fmt.Println("Next: install Tailscale from https://tailscale.com/download")
	} else {
		fmt.Println()
		fmt.Println("Next:", tunnel.TailscaleServeCommand(bind, port))
	}
	return nil
}

func tailscalePrintServe(args []string) error {
	fs := flag.NewFlagSet("tailscale print-serve", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "fortmemory config.toml")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	bind, port := resolveFMListen(*cfgPath)
	fmt.Println(tunnel.TailscaleServeCommand(bind, port))
	return nil
}
