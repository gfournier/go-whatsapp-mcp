package tools

import (
	"github.com/gfournier/go-whatsapp-mcp/config"
	"github.com/gfournier/go-whatsapp-mcp/whatsapp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterAll(s *server.MCPServer, client *whatsapp.Client, cfg *config.Config) {
	s.AddTool(listAllowedJIDsTool(), listAllowedJIDsHandler(cfg))
	s.AddTool(getMessagesTool(), getMessagesHandler(client))
	s.AddTool(sendMessageTool(), sendMessageHandler(client))
	s.AddTool(getChatInfoTool(), getChatInfoHandler(client))
}
