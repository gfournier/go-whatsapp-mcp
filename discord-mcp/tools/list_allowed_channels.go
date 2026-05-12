package tools

import (
	"context"
	"encoding/json"

	"github.com/gfournier/go-whatsapp-mcp/discord-mcp/config"
	"github.com/gfournier/go-whatsapp-mcp/discord-mcp/discord"
	"github.com/mark3labs/mcp-go/mcp"
)

func listAllowedChannelsTool() mcp.Tool {
	return mcp.NewTool("list_allowed_channels",
		mcp.WithDescription("List the Discord text channels this agent is permitted to read from and write to, as configured by DISCORD_ALLOWED_CHANNELS. Returns channel name, guild ID, and topic for each."),
	)
}

func listAllowedChannelsHandler(client *discord.Client, cfg *config.Config) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ids := cfg.AllowedChannelList()

		result := make([]*discord.Channel, 0, len(ids))
		for _, id := range ids {
			info, err := client.GetChannelInfo(ctx, id)
			if err != nil {
				// Return a partial entry so the caller knows the ID exists but lookup failed.
				result = append(result, &discord.Channel{ID: id, Name: "<error: " + err.Error() + ">"})
				continue
			}
			result = append(result, info)
		}

		data, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError("failed to encode result"), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
}
