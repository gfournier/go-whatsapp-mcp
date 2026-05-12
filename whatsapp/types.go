package whatsapp

import "time"

type Message struct {
	ID         string    `json:"id"`
	JID        string    `json:"jid"`
	Sender     string    `json:"sender"`
	SenderName string    `json:"sender_name"`
	Body       string    `json:"body"`
	Timestamp  time.Time `json:"timestamp"`
	IsFromMe   bool      `json:"is_from_me"`
}

type Chat struct {
	JID     string `json:"jid"`
	Name    string `json:"name"`
	IsGroup bool   `json:"is_group"`
}

type Participant struct {
	JID          string `json:"jid"`
	IsAdmin      bool   `json:"is_admin"`
	IsSuperAdmin bool   `json:"is_super_admin"`
}

type GroupInfo struct {
	JID              string        `json:"jid"`
	Name             string        `json:"name"`
	Description      string        `json:"description"`
	ParticipantCount int           `json:"participant_count"`
	Participants     []Participant `json:"participants,omitempty"`
}
