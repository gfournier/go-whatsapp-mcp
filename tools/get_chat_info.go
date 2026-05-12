package tools

import (
	"context"
	"encoding/json"

	"github.com/gfournier/go-whatsapp-mcp/whatsapp"
	"github.com/mark3labs/mcp-go/mcp"
)

func getChatInfoTool() mcp.Tool {
	return mcp.NewTool("get_chat_info",
		mcp.WithDescription("Get metadata about a WhatsApp group: name, description, and optionally the participant list."),
		mcp.WithString("jid",
			mcp.Required(),
			mcp.Description("The WhatsApp group JID (e.g. 120363000000000001@g.us)."),
		),
		mcp.WithBoolean("include_participants",
			mcp.Description("Whether to include the full participant list with phone numbers (default false)."),
			mcp.DefaultBool(false),
		),
	)
}

func getChatInfoHandler(client *whatsapp.Client) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		jid, err := req.RequireString("jid")
		if err != nil {
			return mcp.NewToolResultError("jid is required"), nil
		}

		includeParticipants := req.GetBool("include_participants", false)

		info, err := client.GetGroupInfo(ctx, jid, includeParticipants)
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
