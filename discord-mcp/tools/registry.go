package tools

import (
	"github.com/gfournier/go-whatsapp-mcp/discord-mcp/config"
	"github.com/gfournier/go-whatsapp-mcp/discord-mcp/discord"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterAll(s *server.MCPServer, client *discord.Client, cfg *config.Config) {
	s.AddTool(listAllowedChannelsTool(), listAllowedChannelsHandler(client, cfg))
	s.AddTool(getMessagesTool(), getMessagesHandler(client))
	s.AddTool(sendMessageTool(), sendMessageHandler(client))
	s.AddTool(getChannelInfoTool(), getChannelInfoHandler(client))
}
