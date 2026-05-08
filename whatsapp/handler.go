package whatsapp

import (
	"time"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
)

type eventHandler struct {
	client *Client
}

func (h *eventHandler) handle(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		h.handleMessage(v)
	case *events.HistorySync:
		h.handleHistorySync(v)
	case *events.Connected:
		h.client.log.Infof("WhatsApp connected")
		h.client.markReady()
	case *events.Disconnected:
		h.client.log.Warnf("WhatsApp disconnected, reconnecting...")
		go h.client.reconnectLoop()
	}
}

func (h *eventHandler) handleMessage(evt *events.Message) {
	body := extractText(evt.Message)
	if body == "" {
		return
	}
	msg := Message{
		ID:         evt.Info.ID,
		JID:        evt.Info.Chat.String(),
		Sender:     evt.Info.Sender.String(),
		SenderName: evt.Info.PushName,
		Body:       body,
		Timestamp:  evt.Info.Timestamp,
		IsFromMe:   evt.Info.IsFromMe,
	}
	h.client.msgs.Append(msg.JID, msg)
}

func (h *eventHandler) handleHistorySync(evt *events.HistorySync) {
	for _, conv := range evt.Data.GetConversations() {
		jid := conv.GetID()
		for _, wrapper := range conv.GetMessages() {
			webMsg := wrapper.GetMessage()
			if webMsg == nil {
				continue
			}
			body := extractText(webMsg.GetMessage())
			if body == "" {
				continue
			}
			key := webMsg.GetKey()
			sender := key.GetParticipant()
			if sender == "" {
				sender = key.GetRemoteJID()
			}
			msg := Message{
				ID:         key.GetID(),
				JID:        jid,
				Sender:     sender,
				SenderName: webMsg.GetPushName(),
				Body:       body,
				Timestamp:  time.Unix(int64(webMsg.GetMessageTimestamp()), 0),
				IsFromMe:   key.GetFromMe(),
			}
			h.client.msgs.Append(jid, msg)
		}
	}
}

func extractText(msg *waE2E.Message) string {
	if msg == nil {
		return ""
	}
	if t := msg.GetConversation(); t != "" {
		return t
	}
	if t := msg.GetExtendedTextMessage().GetText(); t != "" {
		return t
	}
	if t := msg.GetEphemeralMessage().GetMessage().GetConversation(); t != "" {
		return t
	}
	return ""
}
