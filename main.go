package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
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
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// All output except MCP JSON-RPC goes to stderr.
	level := zerolog.InfoLevel
	if cfg.LogLevel == "debug" {
		level = zerolog.DebugLevel
	} else if cfg.LogLevel == "warn" {
		level = zerolog.WarnLevel
	}
	zlog := zerolog.New(os.Stderr).With().Timestamp().Logger().Level(level)
	log := waLog.Zerolog(zlog)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	container, err := wastore.OpenSQLite(ctx, cfg.DBPath, log.Sub("db"))
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}

	waClient, err := whatsapp.NewClient(ctx, container, cfg, log.Sub("whatsapp"))
	if err != nil {
		return fmt.Errorf("create WhatsApp client: %w", err)
	}
	defer waClient.Disconnect()

	fmt.Fprintf(os.Stderr, "Connecting to WhatsApp...\n")
	if err := waClient.Connect(ctx); err != nil {
		return fmt.Errorf("connect to WhatsApp: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Waiting for WhatsApp connection...\n")
	if err := waClient.WaitReady(ctx, 60*time.Second); err != nil {
		return fmt.Errorf("WhatsApp not ready: %w", err)
	}
	fmt.Fprintf(os.Stderr, "WhatsApp connected. Starting MCP server.\n")

	mcpServer := server.NewMCPServer("whatsapp-mcp", "1.0.0")
	tools.RegisterAll(mcpServer, waClient, cfg)

	if err := server.ServeStdio(mcpServer); err != nil {
		return fmt.Errorf("MCP server error: %w", err)
	}
	return nil
}
