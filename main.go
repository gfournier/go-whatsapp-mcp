package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/gfournier/go-whatsapp-mcp/config"
	wastore "github.com/gfournier/go-whatsapp-mcp/store"
	"github.com/gfournier/go-whatsapp-mcp/tools"
	"github.com/gfournier/go-whatsapp-mcp/whatsapp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog"
	waLog "go.mau.fi/whatsmeow/util/log"
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
	case "list-chats":
		err = runListChats()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\nUsage: go-whatsapp-mcp [list-chats]\n", cmd)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func setup() (*config.Config, *whatsapp.Client, context.CancelFunc, error) {
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
	zlog := zerolog.New(os.Stderr).With().Timestamp().Logger().Level(level)
	log := waLog.Zerolog(zlog)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	container, err := wastore.OpenSQLite(ctx, cfg.DBPath, log.Sub("db"))
	if err != nil {
		cancel()
		return nil, nil, nil, fmt.Errorf("open sqlite: %w", err)
	}

	waClient, err := whatsapp.NewClient(ctx, container, cfg, log.Sub("whatsapp"))
	if err != nil {
		cancel()
		return nil, nil, nil, fmt.Errorf("create WhatsApp client: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Connecting to WhatsApp...")
	if err := waClient.Connect(ctx); err != nil {
		cancel()
		return nil, nil, nil, fmt.Errorf("connect to WhatsApp: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Waiting for WhatsApp connection...")
	if err := waClient.WaitReady(ctx, 60*time.Second); err != nil {
		cancel()
		return nil, nil, nil, fmt.Errorf("WhatsApp not ready: %w", err)
	}

	return cfg, waClient, cancel, nil
}

func runMCP() error {
	cfg, waClient, cancel, err := setup()
	if err != nil {
		return err
	}
	defer cancel()
	defer waClient.Disconnect()

	fmt.Fprintln(os.Stderr, "WhatsApp connected. Starting MCP server.")
	mcpServer := server.NewMCPServer("whatsapp-mcp", "1.0.0")
	tools.RegisterAll(mcpServer, waClient, cfg)

	if err := server.ServeStdio(mcpServer); err != nil {
		return fmt.Errorf("MCP server error: %w", err)
	}
	return nil
}

func runListChats() error {
	_, waClient, cancel, err := setup()
	if err != nil {
		return err
	}
	defer cancel()
	defer waClient.Disconnect()

	ctx := context.Background()
	chats, err := waClient.ListChats(ctx)
	if err != nil {
		return fmt.Errorf("list chats: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "JID\tNAME\tTYPE")
	fmt.Fprintln(w, "---\t----\t----")
	for _, c := range chats {
		kind := "dm"
		if c.IsGroup {
			kind = "group"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", c.JID, c.Name, kind)
	}
	w.Flush()
	return nil
}
