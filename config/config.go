package config

import (
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

	if raw := os.Getenv("WHATSAPP_ALLOWED_JIDS"); raw != "" {
		for _, jid := range strings.Split(raw, ",") {
			jid = strings.TrimSpace(jid)
			if jid != "" {
				cfg.AllowedJIDs[jid] = true
			}
		}
	}

	return cfg, nil
}

func (c *Config) IsAllowed(jid string) bool {
	if len(c.AllowedJIDs) == 0 {
		return true
	}
	return c.AllowedJIDs[jid]
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
