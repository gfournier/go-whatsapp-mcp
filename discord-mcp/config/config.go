package config

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken        string
	AllowedChannels map[string]bool
	LogLevel        string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		AllowedChannels: make(map[string]bool),
	}

	cfg.BotToken = os.Getenv("DISCORD_BOT_TOKEN")
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("DISCORD_BOT_TOKEN is required: set your Discord bot token from the developer portal")
	}

	raw := os.Getenv("DISCORD_ALLOWED_CHANNELS")
	if raw == "" {
		return nil, fmt.Errorf("DISCORD_ALLOWED_CHANNELS is required: set a comma-separated list of text channel IDs the agent is allowed to access")
	}
	for _, id := range strings.Split(raw, ",") {
		id = strings.TrimSpace(id)
		if id != "" {
			cfg.AllowedChannels[id] = true
		}
	}
	if len(cfg.AllowedChannels) == 0 {
		return nil, fmt.Errorf("DISCORD_ALLOWED_CHANNELS is set but contains no valid channel IDs")
	}

	return cfg, nil
}

func (c *Config) IsAllowed(channelID string) bool {
	return c.AllowedChannels[channelID]
}

func (c *Config) AllowedChannelList() []string {
	ids := make([]string, 0, len(c.AllowedChannels))
	for id := range c.AllowedChannels {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
