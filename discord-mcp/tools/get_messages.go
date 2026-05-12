package tools

import (
	"context"
	"encoding/json"

	"github.com/gfournier/go-whatsapp-mcp/discord-mcp/discord"
	"github.com/mark3labs/mcp-go/mcp"
)

func getMessagesTool() mcp.Tool {
	return mcp.NewTool("get_messages",
		mcp.WithDescription("Fetch recent messages from a Discord text channel. Returns messages newest-first. Maximum 100 per call (Discord API limit)."),
		mcp.WithString("channel_id",
			mcp.Required(),
			mcp.Description("The Discord channel ID (e.g. 1234567890123456789)."),
		),
		mcp.WithInteger("limit",
			mcp.Description("Maximum number of messages to return (default 50, max 100)."),
			mcp.DefaultNumber(50),
		),
	)
}

func getMessagesHandler(client *discord.Client) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		channelID, err := req.RequireString("channel_id")
		if err != nil {
			return mcp.NewToolResultError("channel_id is required"), nil
		}

		limit := req.GetInt("limit", 50)
		if limit <= 0 || limit > 100 {
			limit = 50
		}

		msgs, err := client.GetMessages(ctx, channelID, limit)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, err := json.Marshal(msgs)
		if err != nil {
			return mcp.NewToolResultError("failed to encode result"), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
}
