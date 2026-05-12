package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"

	"github.com/gfournier/go-whatsapp-mcp/discord-mcp/config"
	"github.com/gfournier/go-whatsapp-mcp/discord-mcp/discord"
	"github.com/gfournier/go-whatsapp-mcp/discord-mcp/tools"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog"
)

func main() {
	cmd := ""
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	var err error
	switch cmd {
	case "":
		err = runMCP()
	case "list-channels":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: discord-mcp list-channels <guild-id>")
			os.Exit(1)
		}
		err = runListChannels(os.Args[2])
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\nUsage: discord-mcp [list-channels <guild-id>]\n", cmd)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

// setup loads config, initialises the logger, and creates the Discord client.
// Unlike WhatsApp, there is no SQLite store, QR code, or WebSocket to open —
// NewClient makes a single REST call to validate the bot token.
func setup() (*config.Config, *discord.Client, context.CancelFunc, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("load config: %w", err)
	}

	level := zerolog.InfoLevel
	if cfg.LogLevel == "debug" {
		level = zerolog.DebugLevel
	} else if cfg.LogLevel == "warn" {
		level = zerolog.WarnLevel
	}
	log := zerolog.New(os.Stderr).With().Timestamp().Logger().Level(level)

	_, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	client, err := discord.NewClient(cfg, log)
	if err != nil {
		cancel()
		return nil, nil, nil, fmt.Errorf("create Discord client: %w", err)
	}

	return cfg, client, cancel, nil
}

func runMCP() error {
	cfg, client, cancel, err := setup()
	if err != nil {
		return err
	}
	defer cancel()
	defer client.Disconnect()

	fmt.Fprintln(os.Stderr, "Discord bot authenticated. Starting MCP server.")
	mcpServer := server.NewMCPServer("discord-mcp", "1.0.0")
	tools.RegisterAll(mcpServer, client, cfg)

	if err := server.ServeStdio(mcpServer); err != nil {
		return fmt.Errorf("MCP server error: %w", err)
	}
	return nil
}

func runListChannels(guildID string) error {
	_, client, cancel, err := setup()
	if err != nil {
		return err
	}
	defer cancel()
	defer client.Disconnect()

	channels, err := client.ListGuildChannels(context.Background(), guildID)
	if err != nil {
		return fmt.Errorf("list channels: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tTOPIC")
	fmt.Fprintln(w, "--\t----\t-----")
	for _, ch := range channels {
		topic := ch.Topic
		if topic == "" {
			topic = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", ch.ID, ch.Name, topic)
	}
	w.Flush()
	return nil
}
