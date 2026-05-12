package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBPath             string
	AllowedJIDs        map[string]bool
	MaxMessagesPerChat int
	LogLevel           string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		DBPath:             getEnv("WHATSAPP_DB_PATH", "whatsapp.db"),
		MaxMessagesPerChat: getEnvInt("WHATSAPP_MAX_MESSAGES", 500),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		AllowedJIDs:        make(map[string]bool),
	}

	if cfg.MaxMessagesPerChat <= 0 {
		return nil, fmt.Errorf("WHATSAPP_MAX_MESSAGES must be a positive integer, got %d", cfg.MaxMessagesPerChat)
	}

	raw := os.Getenv("WHATSAPP_ALLOWED_JIDS")
	if raw == "" {
		return nil, fmt.Errorf("WHATSAPP_ALLOWED_JIDS is required: set a comma-separated list of JIDs the agent is allowed to access")
	}
	for _, jid := range strings.Split(raw, ",") {
		jid = strings.TrimSpace(jid)
		if jid != "" {
			cfg.AllowedJIDs[jid] = true
		}
	}
	if len(cfg.AllowedJIDs) == 0 {
		return nil, fmt.Errorf("WHATSAPP_ALLOWED_JIDS is set but contains no valid JIDs")
	}

	return cfg, nil
}

func (c *Config) IsAllowed(jid string) bool {
	return c.AllowedJIDs[jid]
}

func (c *Config) AllowedJIDList() []string {
	jids := make([]string, 0, len(c.AllowedJIDs))
	for jid := range c.AllowedJIDs {
		jids = append(jids, jid)
	}
	return jids
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
