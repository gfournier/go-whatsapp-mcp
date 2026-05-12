package tools

import (
	"context"
	"fmt"

	"github.com/gfournier/go-whatsapp-mcp/whatsapp"
	"github.com/mark3labs/mcp-go/mcp"
)

func sendMessageTool() mcp.Tool {
	return mcp.NewTool("send_message",
		mcp.WithDescription("Send a plain text message to a WhatsApp chat or group."),
		mcp.WithString("jid",
			mcp.Required(),
			mcp.Description("The WhatsApp JID of the target chat."),
		),
		mcp.WithString("text",
			mcp.Required(),
			mcp.Description("The message text to send (max 4096 characters)."),
		),
	)
}

func sendMessageHandler(client *whatsapp.Client) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		jid, err := req.RequireString("jid")
		if err != nil {
			return mcp.NewToolResultError("jid is required"), nil
		}

		text, err := req.RequireString("text")
		if err != nil {
			return mcp.NewToolResultError("text is required"), nil
		}

		ts, err := client.SendMessage(ctx, jid, text)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("message sent at %s", ts.UTC().Format("2006-01-02T15:04:05Z"))), nil
	}
}
