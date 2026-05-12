package discord

import "time"

type Message struct {
	ID         string    `json:"id"`
	ChannelID  string    `json:"channel_id"`
	AuthorID   string    `json:"author_id"`
	AuthorName string    `json:"author_name"`
	Content    string    `json:"content"`
	Timestamp  time.Time `json:"timestamp"`
	IsFromMe   bool      `json:"is_from_me"`
}

type Channel struct {
	ID      string `json:"id"`
	GuildID string `json:"guild_id"`
	Name    string `json:"name"`
	Topic   string `json:"topic,omitempty"`
}
