package whatsapp

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gfournier/go-whatsapp-mcp/config"
	qrterminal "github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/proto"
)

// sendLimiter allows 1 message/second with a burst of 5.
const sendLimitRate = 1
const sendLimitBurst = 5

type Client struct {
	wa      *whatsmeow.Client
	msgs    *MessageStore
	cfg     *config.Config
	log     waLog.Logger
	ready   chan struct{}
	once    sync.Once // ensures ready channel is closed only once
	limiter *rate.Limiter
}

func NewClient(ctx context.Context, container *sqlstore.Container, cfg *config.Config, log waLog.Logger) (*Client, error) {
	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, fmt.Errorf("get device: %w", err)
	}

	c := &Client{
		msgs:    NewMessageStore(cfg.MaxMessagesPerChat),
		cfg:     cfg,
		log:     log,
		ready:   make(chan struct{}),
		limiter: rate.NewLimiter(sendLimitRate, sendLimitBurst),
	}

	c.wa = whatsmeow.NewClient(device, log.Sub("wa"))
	c.wa.EnableAutoReconnect = true

	h := &eventHandler{client: c}
	c.wa.AddEventHandler(h.handle)

	return c, nil
}

// Connect authenticates and connects to WhatsApp.
// On first run it renders a QR code to stderr and waits for the user to scan it.
// On subsequent runs it reconnects with the saved session.
func (c *Client) Connect(ctx context.Context) error {
	if c.wa.Store.ID == nil {
		return c.connectWithQR(ctx)
	}
	return c.wa.ConnectContext(ctx)
}

func (c *Client) connectWithQR(ctx context.Context) error {
	qrChan, err := c.wa.GetQRChannel(ctx)
	if err != nil {
		return fmt.Errorf("get QR channel: %w", err)
	}

	if err := c.wa.ConnectContext(ctx); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	for evt := range qrChan {
		switch evt.Event {
		case whatsmeow.QRChannelEventCode:
			fmt.Fprintln(os.Stderr, "Scan this QR code in WhatsApp (Settings → Linked Devices → Link a Device):")
			// Render locally — never send the pairing code to an external service.
			qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stderr)
			fmt.Fprintln(os.Stderr)
		case "success":
			c.log.Infof("QR scan successful, session saved")
			return nil
		case "timeout":
			return fmt.Errorf("QR code timed out, restart the server to try again")
		default:
			if evt.Error != nil {
				return fmt.Errorf("QR pairing error: %w", evt.Error)
			}
		}
	}
	return fmt.Errorf("QR channel closed without successful pairing")
}

// WaitReady blocks until the WhatsApp connection is established or the context is cancelled.
func (c *Client) WaitReady(ctx context.Context, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-c.ready:
		return nil
	case <-timer.C:
		return fmt.Errorf("timed out waiting for WhatsApp connection")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) markReady() {
	c.once.Do(func() { close(c.ready) })
}

func (c *Client) Disconnect() {
	c.wa.Disconnect()
}

// ListChats returns a combined list: all joined WhatsApp groups plus any
// DM chats seen in the message stream.
func (c *Client) ListChats(ctx context.Context) ([]Chat, error) {
	groups, err := c.wa.GetJoinedGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("get joined groups: %w", err)
	}

	seen := make(map[string]bool)
	var chats []Chat
	for _, g := range groups {
		jidStr := g.JID.String()
		if !c.cfg.IsAllowed(jidStr) {
			continue
		}
		seen[jidStr] = true
		chats = append(chats, Chat{
			JID:     jidStr,
			Name:    g.Name,
			IsGroup: true,
		})
	}

	for _, jid := range c.msgs.KnownJIDs() {
		if seen[jid] {
			continue
		}
		if !c.cfg.IsAllowed(jid) {
			continue
		}
		chats = append(chats, Chat{
			JID:     jid,
			Name:    jid,
			IsGroup: false,
		})
	}

	return chats, nil
}

// GetMessages returns up to limit recent messages from a chat, newest first.
func (c *Client) GetMessages(_ context.Context, jid string, limit int) ([]Message, error) {
	if !c.cfg.IsAllowed(jid) {
		return nil, fmt.Errorf("chat %q is not in the allowed list", jid)
	}
	return c.msgs.Get(jid, limit), nil
}

// SendMessage sends a plain text message to the given JID.
// Rate-limited to 1 message/second (burst of 5) to protect the linked account.
func (c *Client) SendMessage(ctx context.Context, jid string, text string) (time.Time, error) {
	if !c.cfg.IsAllowed(jid) {
		return time.Time{}, fmt.Errorf("chat %q is not in the allowed list", jid)
	}
	if len(text) > 4096 {
		return time.Time{}, fmt.Errorf("message exceeds 4096 characters")
	}

	if err := c.limiter.Wait(ctx); err != nil {
		return time.Time{}, fmt.Errorf("rate limit: %w", err)
	}

	parsedJID, err := types.ParseJID(jid)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid JID %q: %w", jid, err)
	}

	resp, err := c.wa.SendMessage(ctx, parsedJID, &waE2E.Message{
		Conversation: proto.String(text),
	})
	if err != nil {
		return time.Time{}, fmt.Errorf("send message: %w", err)
	}
	return resp.Timestamp, nil
}

// GetGroupInfo returns metadata for a WhatsApp group.
// When includeParticipants is false, the Participants list is omitted to avoid
// exposing member phone numbers unnecessarily.
func (c *Client) GetGroupInfo(ctx context.Context, jid string, includeParticipants bool) (*GroupInfo, error) {
	if !c.cfg.IsAllowed(jid) {
		return nil, fmt.Errorf("chat %q is not in the allowed list", jid)
	}

	parsedJID, err := types.ParseJID(jid)
	if err != nil {
		return nil, fmt.Errorf("invalid JID %q: %w", jid, err)
	}

	info, err := c.wa.GetGroupInfo(ctx, parsedJID)
	if err != nil {
		return nil, fmt.Errorf("get group info: %w", err)
	}

	result := &GroupInfo{
		JID:              info.JID.String(),
		Name:             info.Name,
		Description:      info.Topic,
		ParticipantCount: len(info.Participants),
	}

	if includeParticipants {
		participants := make([]Participant, len(info.Participants))
		for i, p := range info.Participants {
			participants[i] = Participant{
				JID:          p.JID.String(),
				IsAdmin:      p.IsAdmin,
				IsSuperAdmin: p.IsSuperAdmin,
			}
		}
		result.Participants = participants
	}

	return result, nil
}
