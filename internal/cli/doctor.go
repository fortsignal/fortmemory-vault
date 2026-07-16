package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fortsignal/fortmemory-vault/internal/agent"
	"github.com/fortsignal/fortmemory-vault/internal/app"
	"github.com/fortsignal/fortmemory-vault/internal/config"
	"github.com/fortsignal/fortmemory-vault/internal/fortsignal"
	"github.com/fortsignal/fortmemory-vault/internal/memory"
	"github.com/fortsignal/fortmemory-vault/internal/vault"
)

func runDoctor(args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "path to config.toml")
	keyPath := fs.String("key", "", "agent key JSON for live probe")
	agentID := fs.String("agent", "", "agentId for live probe")
	live := fs.Bool("live", true, "call FortSignal (API key + optional probe)")
	writeProbe := fs.Bool("write-probe", false, "full memory.write to Scratch/_doctor_probe.md")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}

	fmt.Println("FortMemory integration doctor")
	fmt.Println("============================")
	ok := true
	pass := func(name, detail string) { fmt.Printf("  OK   %s — %s\n", name, detail) }
	warn := func(name, detail string) { fmt.Printf("  WARN %s — %s\n", name, detail) }
	fail := func(name, detail string) {
		fmt.Printf("  FAIL %s — %s\n", name, detail)
		ok = false
	}

	cfgFile, err := config.Discover(*cfgPath)
	if err != nil {
		fail("config", err.Error())
		printDoctorNext()
		return fmt.Errorf("doctor failed")
	}
	cfg, err := config.Load(cfgFile)
	if err != nil {
		fail("config", err.Error())
		return fmt.Errorf("doctor failed")
	}
	pass("config", cfgFile)

	if st, err := os.Stat(cfg.VaultPath); err != nil || !st.IsDir() {
		fail("vault", fmt.Sprintf("%s (%v)", cfg.VaultPath, err))
	} else {
		pass("vault", cfg.VaultPath)
	}

	rt, err := app.Open(cfgFile, false)
	if err != nil {
		fail("runtime", err.Error())
		return fmt.Errorf("doctor failed")
	}
	defer rt.Close()
	pass("runtime", fmt.Sprintf("vaultId=%s bind=%s:%d", rt.Cfg.VaultID, rt.Cfg.Bind, rt.Cfg.Port))

	list, err := rt.Agents.List(context.Background())
	if err != nil {
		warn("local-agents", err.Error())
	} else if len(list) == 0 {
		warn("local-agents", "none — run: fortmemory agent add <id> --key agent-key.json")
	} else {
		var ids []string
		for _, a := range list {
			tag := a.AgentID
			if a.KeyPath == "" {
				tag += "(no-key)"
			}
			ids = append(ids, tag)
		}
		pass("local-agents", strings.Join(ids, ", "))
	}

	apiKey, keyErr := cfg.APIKey()
	if keyErr != nil {
		if *live {
			fail("fortsignal-api-key", keyErr.Error())
		} else {
			warn("fortsignal-api-key", keyErr.Error())
		}
	} else {
		pass("fortsignal-api-key", "set via "+cfg.FortSignal.APIKeyEnv)
	}

	if !*live || keyErr != nil {
		printDoctorSummary(ok)
		printDoctorNext()
		if !ok {
			return fmt.Errorf("doctor failed")
		}
		return nil
	}

	client := fortsignal.New(apiKey, cfg.FortSignal.APIBase)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	agents, err := client.ListAgents(ctx)
	if err != nil {
		fail("fortsignal-api", err.Error()+" (base="+cfg.FortSignal.APIBase+")")
	} else {
		pass("fortsignal-api", fmt.Sprintf("reachable; %d agent(s) on tenant", len(agents.Agents)))
		for _, a := range agents.Agents {
			del := "no-delegation"
			if a.Delegation != nil {
				del = fmt.Sprintf("delegation=%s policy=%s exp=%s",
					a.Delegation.DelegationID, a.Delegation.PolicyID, a.Delegation.ExpiresAt)
			}
			fmt.Printf("         · %s — %s\n", a.AgentID, del)
		}
	}

	keyFile := *keyPath
	if keyFile == "" {
		keyFile = cfg.Agent.KeyFile
	}
	aid := *agentID
	if aid == "" {
		aid = cfg.Agent.ID
	}

	var signer agent.Signer
	if keyFile != "" {
		sig, err := agent.LoadSigner(keyFile)
		if err != nil {
			fail("agent-key", err.Error())
		} else {
			signer = sig
			if aid == "" {
				aid = sig.AgentID()
			}
			pass("agent-key", fmt.Sprintf("%s (agentId=%s)", keyFile, sig.AgentID()))
		}
	} else {
		warn("agent-key", "not provided — pass --key for challenge/write probe")
	}

	if signer != nil {
		if *writeProbe {
			rt.Memory.Signer = signer
			if rt.Memory.Signers == nil {
				rt.Memory.Signers = map[string]agent.Signer{}
			}
			rt.Memory.Signers[signer.AgentID()] = signer
			rt.Memory.FortSignal = client

			res, err := rt.Memory.Write(ctx, memory.WriteInput{
				AgentID: signer.AgentID(),
				Path:    "Scratch/_doctor_probe.md",
				Content: []byte("# FortMemory doctor probe\n\nSafe to delete.\n"),
				Mode:    vault.ModeOverwrite,
			})
			if err != nil {
				fail("write-probe", err.Error())
			} else if res.Decision == "allow" {
				pass("write-probe", "allow signalId="+res.SignalID+" path="+res.Path)
			} else {
				fail("write-probe", "deny reason="+res.Reason+
					" (need action memory.write, recipients "+cfg.VaultID+"/Scratch/*, active delegation)")
			}
		} else {
			// challenge start + verify without writing vault
			contentHash := fortsignal.ContentHash([]byte("doctor-probe"))
			recipient := cfg.VaultID + "/Scratch/_doctor_probe.md"
			start, err := client.ChallengeStart(ctx, fortsignal.ChallengeStartRequest{
				AgentID:   signer.AgentID(),
				Action:    fortsignal.ActionWrite,
				Amount:    12,
				Recipient: recipient,
				Source:    signer.AgentID(),
				Metadata:  fortsignal.WriteMetadata(cfg.VaultID, contentHash, "overwrite", ""),
			})
			if err != nil {
				fail("challenge-start", err.Error())
			} else if start.Decision == "deny" {
				fail("challenge-start", "deny reason="+start.Reason+
					" — fix policy/delegation for agent "+signer.AgentID())
			} else if start.Challenge == "" {
				fail("challenge-start", "empty challenge")
			} else {
				pass("challenge-start", fmt.Sprintf("challenge ok (delegation=%s expiresIn=%ds)",
					start.DelegationID, start.ExpiresIn))
				sigB64, err := signer.SignChallenge(start.Challenge)
				if err != nil {
					fail("sign", err.Error())
				} else {
					vr, err := client.ChallengeVerify(ctx, fortsignal.ChallengeVerifyRequest{
						AgentID:   signer.AgentID(),
						Challenge: start.Challenge,
						Signature: sigB64,
					})
					if err != nil {
						fail("challenge-verify", err.Error())
					} else if vr.Decision == "allow" {
						pass("challenge-verify", "allow signalId="+vr.SignalID+
							" (no file write; use --write-probe to write Scratch/_doctor_probe.md)")
					} else {
						fail("challenge-verify", "deny reason="+vr.Reason)
					}
				}
			}
		}
	}

	printDoctorSummary(ok)
	printDoctorNext()
	if !ok {
		return fmt.Errorf("doctor found failures — see docs/INTEGRATION.md")
	}
	fmt.Println("\nIntegration looks ready. Next: fortmemory serve")
	return nil
}

func printDoctorSummary(ok bool) {
	fmt.Println()
	if ok {
		fmt.Println("Result: PASS (warnings may still apply)")
	} else {
		fmt.Println("Result: FAIL — fix items above")
	}
}

func printDoctorNext() {
	fmt.Println()
	fmt.Println("Setup checklist: docs/INTEGRATION.md")
	fmt.Println("  1. export FORTSIGNAL_API_KEY=…")
	fmt.Println("  2. Dashboard: passport + policy + delegation")
	fmt.Println("  3. fortmemory agent add <id> --key agent-key.json")
	fmt.Println("  4. fortmemory doctor --key agent-key.json")
	fmt.Println("  5. fortmemory doctor --key agent-key.json --write-probe")
}
