package tools

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/gfournier/go-whatsapp-mcp/config"
	"github.com/mark3labs/mcp-go/mcp"
)

func listAllowedJIDsTool() mcp.Tool {
	return mcp.NewTool("list_allowed_jids",
		mcp.WithDescription("List the JIDs this agent is permitted to read from and write to, as configured by WHATSAPP_ALLOWED_JIDS."),
	)
}

func listAllowedJIDsHandler(cfg *config.Config) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		jids := cfg.AllowedJIDList()
		sort.Strings(jids)
		data, err := json.Marshal(jids)
		if err != nil {
			return mcp.NewToolResultError("failed to encode result"), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
}
