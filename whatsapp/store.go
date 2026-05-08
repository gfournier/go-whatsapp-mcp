package whatsapp

import "sync"

type MessageStore struct {
	mu    sync.RWMutex
	store map[string][]Message
	maxSize   int
}

func NewMessageStore(maxSize int) *MessageStore {
	return &MessageStore{
		store: make(map[string][]Message),
		maxSize:   maxSize,
	}
}

func (s *MessageStore) Append(jid string, msg Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	msgs := s.store[jid]
	msgs = append(msgs, msg)
	if len(msgs) > s.maxSize {
		msgs = msgs[len(msgs)-s.maxSize:]
	}
	s.store[jid] = msgs
}

// Get returns up to limit messages for a JID, newest first.
func (s *MessageStore) Get(jid string, limit int) []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msgs := s.store[jid]
	if len(msgs) == 0 {
		return nil
	}

	start := len(msgs) - limit
	if start < 0 {
		start = 0
	}
	slice := msgs[start:]

	result := make([]Message, len(slice))
	for i, m := range slice {
		result[len(slice)-1-i] = m
	}
	return result
}

func (s *MessageStore) KnownJIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jids := make([]string, 0, len(s.store))
	for jid := range s.store {
		jids = append(jids, jid)
	}
	return jids
}
