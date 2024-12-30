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
	Type      string      `json:"type"`
	RoomID    string      `json:"roomId"`
	SenderID  string      `json:"senderId"`
	Content   string      `json:"content"`
	Data      interface{} `json:"data,omitempty"`
	CreatedAt time.Time   `json:"createdAt"`
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
				"status":       "registered",
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
				"status":       "unregistered",
			}, nil)

		case message := <-h.Broadcast:
			logger.LogInput(map[string]interface{}{
				"messageSize": len(message),
				"action":      "broadcast",
			})

			for client := range h.Clients {
				select {
				case client.Send <- message:
					logger.LogOutput(map[string]interface{}{
						"clientID": client.ID,
						"status":   "message_sent",
					}, nil)
				default:
					h.Mutex.Lock()
					delete(h.Clients, client)
					delete(h.UserMap, client.UserID)
					close(client.Send)
					h.Mutex.Unlock()

					logger.LogOutput(map[string]interface{}{
						"clientID": client.ID,
						"status":   "client_removed",
						"reason":   "send_failed",
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

	// แปลง message เป็น JSON
	messageBytes, err := json.Marshal(message)
	if err != nil {
		logger.LogOutput(nil, fmt.Errorf("error marshaling message: %v", err))
		return
	}

	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	// ส่งข้อความไปยังทุก client ที่อยู่ในห้อง
	for client := range h.Clients {
		if client.RoomIDs[roomID] {
			select {
			case client.Send <- messageBytes:
				logger.LogOutput(map[string]interface{}{
					"clientID": client.ID,
					"status":   "message_sent",
				}, nil)
			default:
				// ถ้าส่งไม่ได้ ให้ลบ client ออก
				delete(h.Clients, client)
				delete(h.UserMap, client.UserID)
				close(client.Send)
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
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.LogOutput(nil, fmt.Errorf("error: %v", err))
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			logger.LogOutput(nil, fmt.Errorf("error unmarshaling message: %v", err))
			continue
		}

		msg.SenderID = c.UserID
		msg.CreatedAt = time.Now()

		switch msg.Type {
		case "message":
			// จัดการข้อความปกติ
			chatMsg, err := c.Hub.ChatUsecase.SendMessage(
				msg.RoomID,
				msg.SenderID,
				"text",
				msg.Content,
			)
			if err != nil {
				logger.LogOutput(nil, fmt.Errorf("error sending message: %v", err))
				continue
			}

			// แปลง chatMsg เป็น Message สำหรับ broadcast
			broadcastMsg := Message{
				Type:      "message",
				RoomID:    chatMsg.RoomID,
				SenderID:  chatMsg.SenderID,
				Content:   chatMsg.Content,
				CreatedAt: chatMsg.CreatedAt,
			}
			c.Hub.BroadcastToRoom(msg.RoomID, broadcastMsg)

		case "typing":
			// สำหรับ typing status ไม่ต้องบันทึกลง database แค่ broadcast ไปยังสมาชิกในห้องเท่านั้น
			typingMsg := Message{
				Type:     "typing",
				RoomID:   msg.RoomID,
				SenderID: msg.SenderID,
				Content:  msg.Content, // "true" หรือ "false"
			}
			c.Hub.BroadcastToRoom(msg.RoomID, typingMsg)

		default:
			logger.LogOutput(nil, fmt.Errorf("unknown message type: %s", msg.Type))
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
				logger.LogInfo("channel closed")
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logger.LogOutput(nil, err)
				return
			}

			logger.LogOutput(map[string]interface{}{
				"messageSize": len(message),
				"action":      "message_sent",
			}, nil)

		case <-ticker.C:
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.LogOutput(nil, err)
				return
			}
		}
	}
}
