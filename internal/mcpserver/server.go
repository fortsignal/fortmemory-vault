// Package mcpserver exposes FortMemory tools over the Model Context Protocol (stdio).
//
// Tools: memory_search, memory_read, memory_write, memory_delete
// All mutates go through memory.Service (FortSignal-gated).
package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/fortsignal/fortmemory-vault/internal/index"
	"github.com/fortsignal/fortmemory-vault/internal/memory"
	"github.com/fortsignal/fortmemory-vault/internal/vault"
)

// Deps for MCP tool handlers.
type Deps struct {
	Memory  *memory.Service
	AgentID string
	Version string
}

// RunStdio starts an MCP server on stdin/stdout until the client disconnects.
func RunStdio(ctx context.Context, d Deps) error {
	if d.Memory == nil {
		return fmt.Errorf("memory service required")
	}
	if d.AgentID == "" {
		return fmt.Errorf("agent id required")
	}
	ver := d.Version
	if ver == "" {
		ver = "0.0.0-dev"
	}

	s := server.NewMCPServer(
		"fortmemory",
		ver,
		server.WithToolCapabilities(true),
	)

	s.AddTool(mcp.NewTool("memory_search",
		mcp.WithDescription("Search the FortMemory vault (FTS). Returns paths, excerpts, scores."),
		mcp.WithString("q", mcp.Required(), mcp.Description("Search query")),
		mcp.WithNumber("topK", mcp.Description("Max results (default 8)")),
		mcp.WithString("pathPrefix", mcp.Description("Optional path prefix filter e.g. Scratch/")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q, err := req.RequireString("q")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		topK := 8
		if v := req.GetFloat("topK", 0); v > 0 {
			topK = int(v)
		}
		hits, err := d.Memory.Search(ctx, d.AgentID, index.SearchRequest{
			Query:      q,
			TopK:       topK,
			PathPrefix: req.GetString("pathPrefix", ""),
		})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		b, _ := json.MarshalIndent(hits, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	})

	s.AddTool(mcp.NewTool("memory_read",
		mcp.WithDescription("Read a Markdown note from the vault by relative path."),
		mcp.WithString("path", mcp.Required(), mcp.Description("Relative path e.g. Scratch/note.md")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		body, hash, err := d.Memory.Read(ctx, d.AgentID, path)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		out := map[string]any{"path": path, "content": string(body), "contentHash": hash}
		b, _ := json.MarshalIndent(out, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	})

	s.AddTool(mcp.NewTool("memory_write",
		mcp.WithDescription("Write a note via FortSignal-governed memory.write. Returns decision + signalId or deny reason."),
		mcp.WithString("path", mcp.Required(), mcp.Description("Relative path under vault")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Full Markdown body to write")),
		mcp.WithString("mode", mcp.Description("create|append|overwrite (default overwrite)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		content, err := req.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		mode := vault.WriteMode(req.GetString("mode", "overwrite"))
		if mode == "" {
			mode = vault.ModeOverwrite
		}
		res, err := d.Memory.Write(ctx, memory.WriteInput{
			AgentID: d.AgentID,
			Path:    path,
			Content: []byte(content),
			Mode:    mode,
		})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	})

	s.AddTool(mcp.NewTool("memory_delete",
		mcp.WithDescription("Delete a note via FortSignal-governed memory.delete."),
		mcp.WithString("path", mcp.Required(), mcp.Description("Relative path to delete")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		res, err := d.Memory.Delete(ctx, d.AgentID, path)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	})

	_ = ctx
	return server.ServeStdio(s)
}
