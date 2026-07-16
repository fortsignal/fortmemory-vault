// Command fortmemory is the FortMemory CLI and local memory server entrypoint.
//
// Subcommands (target):
//
//	fortmemory version
//	fortmemory init [path]
//	fortmemory serve
//	fortmemory reindex
//	fortmemory agent add <agentId>
//
// See docs/CLI.md and docs/PROJECT-LAYOUT.md.
package main

import (
	"fmt"
	"os"

	"github.com/fortsignal/fortmemory-vault/internal/cli"
)

func main() {
	if err := cli.Execute(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "fortmemory: %v\n", err)
		os.Exit(1)
	}
}
