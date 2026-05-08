package tools

import (
	"context"
	"encoding/json"

	"github.com/gfournier/go-whatsapp-mcp/whatsapp"
	"github.com/mark3labs/mcp-go/mcp"
)

func listChatsTool() mcp.Tool {
	return mcp.NewTool("list_chats",
		mcp.WithDescription("List available WhatsApp chats and groups that the bot has access to."),
	)
}

func listChatsHandler(client *whatsapp.Client) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		chats, err := client.ListChats(ctx)
		if err != nil {
			return mcp.NewToolResultError("failed to list chats: " + err.Error()), nil
		}
		data, err := json.Marshal(chats)
		if err != nil {
			return mcp.NewToolResultError("failed to encode result"), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
}
