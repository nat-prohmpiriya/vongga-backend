package websocket

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
)

type Client struct {
	ID      string
	UserID  string
	Conn    *websocket.Conn
	Send    chan []byte
	Hub     *Hub
	RoomIDs map[string]bool
}

type Message struct {
	Type      string          `json:"type"`
	RoomID    string          `json:"roomId"`
	SenderID  string          `json:"senderId"`
	Content   string          `json:"content"`
	Data      interface{}     `json:"data,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
}

type Hub struct {
	Clients     map[*Client]bool
	UserMap     map[string]*Client // maps userID to client
	Broadcast   chan []byte
	Register    chan *Client
	Unregister  chan *Client
	Mutex       sync.Mutex
	ChatUsecase domain.ChatUsecase
}

func NewHub(chatUsecase domain.ChatUsecase) *Hub {
	return &Hub{
		Clients:     make(map[*Client]bool),
		UserMap:     make(map[string]*Client),
		Broadcast:   make(chan []byte),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		ChatUsecase: chatUsecase,
	}
}

func (h *Hub) Run() {
	logger := utils.NewLogger("Hub.Run")

	for {
		select {
		case client := <-h.Register:
			h.Mutex.Lock()
			h.Clients[client] = true
			h.UserMap[client.UserID] = client
			h.Mutex.Unlock()

			logger.LogOutput(map[string]interface{}{
				"totalClients": len(h.Clients),
				"status":      "registered",
			}, nil)

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				h.Mutex.Lock()
				delete(h.Clients, client)
				delete(h.UserMap, client.UserID)
				close(client.Send)
				h.Mutex.Unlock()
			}

			logger.LogOutput(map[string]interface{}{
				"totalClients": len(h.Clients),
				"status":      "unregistered",
			}, nil)

		case message := <-h.Broadcast:
			logger.LogInput(map[string]interface{}{
				"messageSize": len(message),
				"action":     "broadcast",
			})

			for client := range h.Clients {
				select {
				case client.Send <- message:
					logger.LogOutput(map[string]interface{}{
						"clientID": client.ID,
						"status":  "message_sent",
					}, nil)
				default:
					h.Mutex.Lock()
					delete(h.Clients, client)
					delete(h.UserMap, client.UserID)
					close(client.Send)
					h.Mutex.Unlock()

					logger.LogOutput(map[string]interface{}{
						"clientID": client.ID,
						"status":  "client_removed",
						"reason":  "send_failed",
					}, nil)
				}
			}
		}
	}
}

func (h *Hub) BroadcastToRoom(roomID string, message interface{}) {
	logger := utils.NewLogger("Hub.BroadcastToRoom")
	logger.LogInput(map[string]interface{}{
		"roomID":  roomID,
		"message": message,
	})

	// Get room members
	room, err := h.ChatUsecase.GetRoom(roomID)
	if err != nil {
		logger.LogOutput(nil, err)
		return
	}

	// Convert message to JSON
	msgBytes, err := json.Marshal(message)
	if err != nil {
		logger.LogOutput(nil, err)
		return
	}

	// Send to all room members
	for _, memberID := range room.Members {
		if client, ok := h.UserMap[memberID]; ok {
			select {
			case client.Send <- msgBytes:
				logger.LogOutput(map[string]interface{}{
					"clientID": client.ID,
					"status":  "message_sent",
				}, nil)
			default:
				h.Mutex.Lock()
				delete(h.Clients, client)
				delete(h.UserMap, client.UserID)
				close(client.Send)
				h.Mutex.Unlock()

				logger.LogOutput(map[string]interface{}{
					"clientID": client.ID,
					"status":  "client_removed",
					"reason":  "send_failed",
				}, nil)
			}
		}
	}
}

func (h *Hub) BroadcastUserStatus(userID string, status string) {
	logger := utils.NewLogger("Hub.BroadcastUserStatus")
	logger.LogInput(map[string]interface{}{
		"userID": userID,
		"status": status,
	})

	msg := Message{
		Type:      "user_status",
		SenderID:  userID,
		Content:   status,
		CreatedAt: time.Now(),
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		logger.LogOutput(nil, err)
		return
	}

	h.Broadcast <- msgBytes
}

func (c *Client) ReadPump() {
	logger := utils.NewLogger("Client.ReadPump")
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			logger.LogOutput(nil, err)
			break
		}

		logger.LogInput(map[string]interface{}{
			"messageSize": len(message),
			"action":     "message_received",
		})

		// Parse message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			logger.LogOutput(nil, err)
			continue
		}

		// Handle message based on type
		switch msg.Type {
		case "join_room":
			roomID := msg.RoomID
			c.RoomIDs[roomID] = true

			// Notify room members
			c.Hub.BroadcastToRoom(roomID, Message{
				Type:      "user_joined",
				RoomID:    roomID,
				SenderID:  c.UserID,
				Content:   "User joined the room",
				CreatedAt: time.Now(),
			})

		case "leave_room":
			roomID := msg.RoomID
			delete(c.RoomIDs, roomID)

			// Notify room members
			c.Hub.BroadcastToRoom(roomID, Message{
				Type:      "user_left",
				RoomID:    roomID,
				SenderID:  c.UserID,
				Content:   "User left the room",
				CreatedAt: time.Now(),
			})

		case "chat_message":
			// Save message to database
			chatMsg := &domain.ChatMessage{
				RoomID:   msg.RoomID,
				SenderID: msg.SenderID,
				Content:  msg.Content,
				Type:     "text",
			}
			if _, err := c.Hub.ChatUsecase.SendMessage(chatMsg.RoomID, chatMsg.SenderID, chatMsg.Type, chatMsg.Content); err != nil {
				logger.LogOutput(nil, err)
				continue
			}

			// Broadcast message to room
			c.Hub.BroadcastToRoom(msg.RoomID, msg)
		}
	}
}

func (c *Client) WritePump() {
	logger := utils.NewLogger("Client.WritePump")
	ticker := time.NewTicker(60 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				logger.LogOutput(nil, fmt.Errorf("send channel closed"))
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logger.LogOutput(nil, err)
				return
			}

			logger.LogOutput(map[string]interface{}{
				"messageSize": len(message),
				"action":     "message_sent",
			}, nil)

		case <-ticker.C:
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.LogOutput(nil, err)
				return
			}
		}
	}
}
