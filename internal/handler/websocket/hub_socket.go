// internal/handler/websocket/hub_socket.go
package websocket

import (
	"sync"
)

type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				h.mu.Lock()
				delete(h.Clients, client)
				close(client.Send)
				h.mu.Unlock()
			}

		case message := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					h.mu.Lock()
					delete(h.Clients, client)
					close(client.Send)
					h.mu.Unlock()
				}
			}
		}
	}
}
