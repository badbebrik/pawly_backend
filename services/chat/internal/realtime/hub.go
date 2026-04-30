package realtime

import (
	"sync"

	"github.com/google/uuid"
)

type Hub struct {
	mu sync.RWMutex

	clientsByUser map[uuid.UUID]map[*Client]struct{}
	inboxSubs     map[*Client]struct{}
	convSubs      map[uuid.UUID]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clientsByUser: make(map[uuid.UUID]map[*Client]struct{}),
		inboxSubs:     make(map[*Client]struct{}),
		convSubs:      make(map[uuid.UUID]map[*Client]struct{}),
	}
}

type Client struct {
	ID     uuid.UUID
	UserID uuid.UUID

	send chan []byte
	done chan struct{}
	once sync.Once
	mu   sync.RWMutex
	subs map[uuid.UUID]struct{}
}

func NewClient(userID uuid.UUID, send chan []byte) *Client {
	return &Client{
		ID:     uuid.New(),
		UserID: userID,
		send:   send,
		done:   make(chan struct{}),
		subs:   make(map[uuid.UUID]struct{}),
	}
}

func (h *Hub) AddClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clientsByUser[client.UserID] == nil {
		h.clientsByUser[client.UserID] = make(map[*Client]struct{})
	}
	h.clientsByUser[client.UserID][client] = struct{}{}
}

func (h *Hub) RemoveClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients := h.clientsByUser[client.UserID]; clients != nil {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.clientsByUser, client.UserID)
		}
	}

	delete(h.inboxSubs, client)

	for conversationID, clients := range h.convSubs {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.convSubs, conversationID)
		}
	}

	client.close()
}

func (h *Hub) SubscribeInbox(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.inboxSubs[client] = struct{}{}
}

func (h *Hub) SubscribeConversation(conversationID uuid.UUID, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.convSubs[conversationID] == nil {
		h.convSubs[conversationID] = make(map[*Client]struct{})
	}
	h.convSubs[conversationID][client] = struct{}{}
}

func (h *Hub) UnsubscribeConversation(conversationID uuid.UUID, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients := h.convSubs[conversationID]
	if clients == nil {
		return
	}

	delete(clients, client)
	if len(clients) == 0 {
		delete(h.convSubs, conversationID)
	}
}

func (h *Hub) PublishToClient(client *Client, message []byte) {
	client.publish(message)
}

func (h *Hub) PublishToUserInbox(userID uuid.UUID, message []byte) {
	h.mu.RLock()
	clients := h.clientsByUser[userID]
	targets := make([]*Client, 0, len(clients))
	for client := range clients {
		if _, ok := h.inboxSubs[client]; ok {
			targets = append(targets, client)
		}
	}
	h.mu.RUnlock()

	for _, client := range targets {
		client.publish(message)
	}
}

func (h *Hub) PublishToConversation(conversationID uuid.UUID, message []byte) {
	h.mu.RLock()
	clients := h.convSubs[conversationID]
	targets := make([]*Client, 0, len(clients))
	for client := range clients {
		targets = append(targets, client)
	}
	h.mu.RUnlock()

	for _, client := range targets {
		client.publish(message)
	}
}

func (h *Hub) PublishToConversationExceptClientID(conversationID, excludedClientID uuid.UUID, message []byte) {
	h.mu.RLock()
	clients := h.convSubs[conversationID]
	targets := make([]*Client, 0, len(clients))
	for client := range clients {
		if client.ID == excludedClientID {
			continue
		}
		targets = append(targets, client)
	}
	h.mu.RUnlock()

	for _, client := range targets {
		client.publish(message)
	}
}

func (h *Hub) HasConversationSubscribers(conversationID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.convSubs[conversationID]) > 0
}

func (h *Hub) HasUserInboxSubscribers(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients := h.clientsByUser[userID]
	for client := range clients {
		if _, ok := h.inboxSubs[client]; ok {
			return true
		}
	}
	return false
}

func (c *Client) publish(message []byte) {
	select {
	case <-c.done:
		return
	case c.send <- message:
	default:
	}
}

func (c *Client) Done() <-chan struct{} {
	return c.done
}

func (c *Client) Close() {
	c.close()
}

func (c *Client) close() {
	c.once.Do(func() {
		close(c.done)
	})
}

func (c *Client) AddConversationSubscription(conversationID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subs[conversationID] = struct{}{}
}

func (c *Client) RemoveConversationSubscription(conversationID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subs, conversationID)
}

func (c *Client) ConversationSubscriptions() []uuid.UUID {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ids := make([]uuid.UUID, 0, len(c.subs))
	for conversationID := range c.subs {
		ids = append(ids, conversationID)
	}
	return ids
}
