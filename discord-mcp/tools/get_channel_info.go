package tools

import (
	"context"
	"encoding/json"

	"github.com/gfournier/go-whatsapp-mcp/discord-mcp/discord"
	"github.com/mark3labs/mcp-go/mcp"
)

func getChannelInfoTool() mcp.Tool {
	return mcp.NewTool("get_channel_info",
		mcp.WithDescription("Get metadata about a Discord text channel: name, guild ID, and topic."),
		mcp.WithString("channel_id",
			mcp.Required(),
			mcp.Description("The Discord channel ID."),
		),
	)
}

func getChannelInfoHandler(client *discord.Client) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		channelID, err := req.RequireString("channel_id")
		if err != nil {
			return mcp.NewToolResultError("channel_id is required"), nil
		}

		info, err := client.GetChannelInfo(ctx, channelID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, err := json.Marshal(info)
		if err != nil {
			return mcp.NewToolResultError("failed to encode result"), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
}
