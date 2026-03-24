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
	UserID uuid.UUID

	send chan []byte
}
