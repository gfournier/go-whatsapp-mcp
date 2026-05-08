package tools

import (
	"context"
	"encoding/json"

	"github.com/gfournier/go-whatsapp-mcp/whatsapp"
	"github.com/mark3labs/mcp-go/mcp"
)

func getMessagesTool() mcp.Tool {
	return mcp.NewTool("get_messages",
		mcp.WithDescription("Fetch recent messages from a WhatsApp chat or group. Returns messages newest-first."),
		mcp.WithString("jid",
			mcp.Required(),
			mcp.Description("The WhatsApp JID of the chat (e.g. 120363000000000001@g.us for a group, 15551234567@s.whatsapp.net for a DM)."),
		),
		mcp.WithInteger("limit",
			mcp.Description("Maximum number of messages to return (default 50, max 200)."),
			mcp.DefaultNumber(50),
		),
	)
}

func getMessagesHandler(client *whatsapp.Client) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		jid, err := req.RequireString("jid")
		if err != nil {
			return mcp.NewToolResultError("jid is required"), nil
		}

		limit := req.GetInt("limit", 50)
		if limit <= 0 || limit > 200 {
			limit = 50
		}

		msgs, err := client.GetMessages(ctx, jid, limit)
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
