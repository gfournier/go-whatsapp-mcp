package tools

import (
	"context"
	"fmt"

	"github.com/gfournier/go-whatsapp-mcp/discord-mcp/discord"
	"github.com/mark3labs/mcp-go/mcp"
)

func sendMessageTool() mcp.Tool {
	return mcp.NewTool("send_message",
		mcp.WithDescription("Send a plain text message to a Discord channel."),
		mcp.WithString("channel_id",
			mcp.Required(),
			mcp.Description("The Discord channel ID to send to."),
		),
		mcp.WithString("text",
			mcp.Required(),
			mcp.Description("The message text to send (max 2000 characters, Discord limit)."),
		),
	)
}

func sendMessageHandler(client *discord.Client) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		channelID, err := req.RequireString("channel_id")
		if err != nil {
			return mcp.NewToolResultError("channel_id is required"), nil
		}

		text, err := req.RequireString("text")
		if err != nil {
			return mcp.NewToolResultError("text is required"), nil
		}

		ts, err := client.SendMessage(ctx, channelID, text)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("message sent at %s", ts.UTC().Format("2006-01-02T15:04:05Z"))), nil
	}
}
