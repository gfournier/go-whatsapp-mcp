package discord

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gfournier/go-whatsapp-mcp/discord-mcp/config"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

const sendLimitRate = 1
const sendLimitBurst = 5

type Client struct {
	session *discordgo.Session
	cfg     *config.Config
	log     zerolog.Logger
	selfID  string
	limiter *rate.Limiter
}

func NewClient(cfg *config.Config, log zerolog.Logger) (*Client, error) {
	session, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("create discord session: %w", err)
	}

	self, err := session.User("@me")
	if err != nil {
		return nil, fmt.Errorf("verify bot token: %w", err)
	}

	return &Client{
		session: session,
		cfg:     cfg,
		log:     log,
		selfID:  self.ID,
		limiter: rate.NewLimiter(sendLimitRate, sendLimitBurst),
	}, nil
}

func (c *Client) Disconnect() {
	_ = c.session.Close()
}

// GetMessages returns up to limit recent messages from a channel, newest first.
func (c *Client) GetMessages(_ context.Context, channelID string, limit int) ([]Message, error) {
	if !c.cfg.IsAllowed(channelID) {
		return nil, fmt.Errorf("channel %q is not in the allowed list", channelID)
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// Empty before/after/around returns the most recent messages, newest first.
	raw, err := c.session.ChannelMessages(channelID, limit, "", "", "")
	if err != nil {
		return nil, fmt.Errorf("fetch messages from channel %q: %w", channelID, err)
	}

	msgs := make([]Message, len(raw))
	for i, m := range raw {
		msgs[i] = Message{
			ID:         m.ID,
			ChannelID:  m.ChannelID,
			AuthorID:   m.Author.ID,
			AuthorName: m.Author.Username,
			Content:    m.Content,
			Timestamp:  m.Timestamp,
			IsFromMe:   m.Author.ID == c.selfID,
		}
	}
	return msgs, nil
}

// SendMessage sends a plain text message to a Discord channel.
// Rate-limited to 1 message/second (burst 5) on top of discordgo's built-in HTTP rate limiter.
func (c *Client) SendMessage(ctx context.Context, channelID string, text string) (time.Time, error) {
	if !c.cfg.IsAllowed(channelID) {
		return time.Time{}, fmt.Errorf("channel %q is not in the allowed list", channelID)
	}
	if len(text) > 2000 {
		return time.Time{}, fmt.Errorf("message exceeds 2000 characters (Discord limit)")
	}

	if err := c.limiter.Wait(ctx); err != nil {
		return time.Time{}, fmt.Errorf("rate limit: %w", err)
	}

	msg, err := c.session.ChannelMessageSend(channelID, text)
	if err != nil {
		return time.Time{}, fmt.Errorf("send message to channel %q: %w", channelID, err)
	}

	return msg.Timestamp, nil
}

// GetChannelInfo returns metadata for a Discord text channel.
func (c *Client) GetChannelInfo(_ context.Context, channelID string) (*Channel, error) {
	if !c.cfg.IsAllowed(channelID) {
		return nil, fmt.Errorf("channel %q is not in the allowed list", channelID)
	}

	ch, err := c.session.Channel(channelID)
	if err != nil {
		return nil, fmt.Errorf("get channel %q: %w", channelID, err)
	}

	return &Channel{
		ID:      ch.ID,
		GuildID: ch.GuildID,
		Name:    ch.Name,
		Topic:   ch.Topic,
	}, nil
}

// ListGuildChannels returns all text channels in a guild.
// Used by the list-channels CLI subcommand; no allowlist check.
func (c *Client) ListGuildChannels(_ context.Context, guildID string) ([]Channel, error) {
	all, err := c.session.GuildChannels(guildID)
	if err != nil {
		return nil, fmt.Errorf("list channels for guild %q: %w", guildID, err)
	}

	var result []Channel
	for _, ch := range all {
		if ch.Type != discordgo.ChannelTypeGuildText {
			continue
		}
		result = append(result, Channel{
			ID:      ch.ID,
			GuildID: ch.GuildID,
			Name:    ch.Name,
			Topic:   ch.Topic,
		})
	}
	return result, nil
}
