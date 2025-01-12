package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"github.com/gofiber/websocket/v2"
	"go.opentelemetry.io/otel/trace"
)

// Message types
const (
	MessageTypeMessage    = "message"
	MessageTypeTyping     = "typing"
	MessageTypePing       = "ping"
	MessageTypePong       = "pong"
	MessageTypeUserStatus = "userStatus"
)

// WebSocketMessage represents the message structure for WebSocket communication
type WebSocketMessage struct {
	Type      string      `json:"type"`                // message, typing, ping, pong
	RoomID    string      `json:"roomId"`              // room identifier
	SenderID  string      `json:"senderId,omitempty"`  // set by server
	Content   string      `json:"content"`             // message content or typing status (true/false)
	Data      interface{} `json:"data,omitempty"`      // additional data if needed
	CreatedAt string      `json:"createdAt,omitempty"` // set by server in RFC3339 format
}

// Client represents a WebSocket client connection
type Client struct {
	ID      string
	UserID  string
	Conn    *websocket.Conn
	Send    chan []byte
	Hub     *Hub
	RoomIDs map[string]bool
	mu      sync.Mutex
}

type Hub struct {
	Clients     map[*Client]bool
	UserMap     map[string]*Client // maps userID to client
	Broadcast   chan []byte
	Register    chan *Client
	Unregister  chan *Client
	Mutex       sync.Mutex
	ChatUsecase domain.ChatUsecase
	tracer      trace.Tracer
}

func NewHub(chatUsecase domain.ChatUsecase, tracer trace.Tracer) *Hub {
	return &Hub{
		Clients:     make(map[*Client]bool),
		UserMap:     make(map[string]*Client),
		Broadcast:   make(chan []byte),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		ChatUsecase: chatUsecase,
		tracer:      tracer,
	}
}

func (h *Hub) Run(ctx context.Context) {
	ctx, span := h.tracer.Start(ctx, "Hub.Run")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	for {
		select {
		case client := <-h.Register:
			h.Mutex.Lock()
			h.Clients[client] = true
			h.UserMap[client.UserID] = client
			h.Mutex.Unlock()

			logger.Output(map[string]interface{}{
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

			logger.Output(map[string]interface{}{
				"totalClients": len(h.Clients),
				"status":       "unregistered",
			}, nil)

		case message := <-h.Broadcast:
			logger.Input(map[string]interface{}{
				"messageSize": len(message),
				"action":      "broadcast",
			})

			for client := range h.Clients {
				select {
				case client.Send <- message:
					logger.Output(map[string]interface{}{
						"clientID": client.ID,
						"status":   "messageSent",
					}, nil)
				default:
					h.Mutex.Lock()
					delete(h.Clients, client)
					delete(h.UserMap, client.UserID)
					close(client.Send)
					h.Mutex.Unlock()

					logger.Output(map[string]interface{}{
						"clientID": client.ID,
						"status":   "clientRemoved",
						"reason":   "sendFailed",
					}, nil)
				}
			}
		}
	}
}

func (h *Hub) BroadcastToRoom(ctx context.Context, roomID string, message interface{}) {
	ctx, span := h.tracer.Start(ctx, "Hub.BroadcastToRoom")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"roomID":  roomID,
		"message": message,
	})

	// แปลง message เป็น JSON
	messageBytes, err := json.Marshal(message)
	if err != nil {
		logger.Output(nil, fmt.Errorf("error marshaling message: %v", err))
		return
	}

	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	// ส่งข้อความไปยังทุก client ที่อยู่ในห้อง
	for client := range h.Clients {
		if client.RoomIDs[roomID] {
			select {
			case client.Send <- messageBytes:
				logger.Output(map[string]interface{}{
					"clientID": client.ID,
					"status":   "messageSent",
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

func (h *Hub) BroadcastUserStatus(ctx context.Context, userID string, status string) {
	ctx, span := h.tracer.Start(ctx, "Hub.BroadcastUserStatus")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"userID": userID,
		"status": status,
	})

	msg := WebSocketMessage{
		Type:      MessageTypeUserStatus,
		SenderID:  userID,
		Content:   status,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		logger.Output(nil, err)
		return
	}

	h.Broadcast <- msgBytes
}

func (c *Client) JoinRoom(roomID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.RoomIDs == nil {
		c.RoomIDs = make(map[string]bool)
	}
	c.RoomIDs[roomID] = true
}

func (c *Client) ReadPump(ctx context.Context) {
	ctx, span := c.Hub.tracer.Start(ctx, "Client.ReadPump")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			logger.Output(nil, fmt.Errorf("panic recovered in ReadPump: %v", r))
		}
		logger.Info("closing connection and unregistering client")
		if c.Hub != nil {
			c.Hub.Unregister <- c
		}
		if c.Conn != nil {
			c.Conn.Close()
		}
	}()

	// Check if connection is valid
	if c.Conn == nil {
		logger.Output(nil, fmt.Errorf("connection is nil"))
		return
	}

	// ดึงข้อมูลห้องแชทที่ user เป็นสมาชิกและเพิ่มเข้าไปใน RoomIDs
	rooms, err := c.Hub.ChatUsecase.FindRoomsByUserID(ctx, c.UserID)
	if err != nil {
		logger.Output(nil, fmt.Errorf("error getting user rooms: %v", err))
	} else {
		for _, room := range rooms {
			c.JoinRoom(room.ID.String())
		}
	}

	// Set read deadline and pong handler
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		messageType, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Output(nil, fmt.Errorf("unexpected close error: %v", err))
			} else {
				logger.Output(nil, fmt.Errorf("read error: %v", err))
			}
			break
		}

		// Handle ping message
		if messageType == websocket.PingMessage {
			if err := c.Conn.WriteMessage(websocket.PongMessage, nil); err != nil {
				logger.Output(nil, fmt.Errorf("error sending pong: %v", err))
				return
			}
			continue
		}

		// Handle text message
		if messageType != websocket.TextMessage {
			logger.Output(nil, fmt.Errorf("unexpected message type: %v", messageType))
			continue
		}

		// Validate message is not empty
		if len(message) == 0 {
			logger.Output(nil, fmt.Errorf("received empty message"))
			continue
		}

		var msg WebSocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			logger.Output(nil, fmt.Errorf("error unmarshaling message: %v", err))
			continue
		}

		// Handle ping type message from client
		if msg.Type == MessageTypePing {
			pongMsg := WebSocketMessage{
				Type:      MessageTypePong,
				CreatedAt: time.Now().Format(time.RFC3339),
			}
			pongBytes, err := json.Marshal(pongMsg)
			if err != nil {
				logger.Output(nil, fmt.Errorf("error marshaling pong message: %v", err))
				continue
			}
			select {
			case c.Send <- pongBytes:
			default:
				logger.Output(nil, fmt.Errorf("send channel full"))
				return
			}
			continue
		}

		msg.SenderID = c.UserID
		msg.CreatedAt = time.Now().Format(time.RFC3339)

		switch msg.Type {
		case MessageTypeMessage:
			if msg.RoomID == "" || msg.Content == "" {
				logger.Output(nil, fmt.Errorf("roomID and content are required for message type"))
				continue
			}

			// จัดการข้อความปกติ
			chatMsg, err := c.Hub.ChatUsecase.SendMessage(
				ctx,
				msg.RoomID,
				msg.SenderID,
				"text",
				msg.Content,
			)
			if err != nil {
				logger.Output(nil, fmt.Errorf("error sending message: %v", err))
				continue
			}

			// เพิ่ม client เข้าห้องแชทถ้ายังไม่ได้อยู่ในห้อง
			c.JoinRoom(msg.RoomID)

			// แปลง chatMsg เป็น Message สำหรับ broadcast
			broadcastMsg := WebSocketMessage{
				Type:      MessageTypeMessage,
				RoomID:    chatMsg.RoomID,
				SenderID:  chatMsg.SenderID,
				Content:   chatMsg.Content,
				CreatedAt: chatMsg.CreatedAt.Format(time.RFC3339),
			}

			// Safely broadcast message
			func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Output(nil, fmt.Errorf("panic recovered in broadcast: %v", r))
					}
				}()
				if c.Hub != nil {
					c.Hub.BroadcastToRoom(ctx, msg.RoomID, broadcastMsg)
				}
			}()

		case MessageTypeTyping:
			if msg.RoomID == "" {
				logger.Output(nil, fmt.Errorf("roomID is required for typing status"))
				continue
			}

			typingMsg := WebSocketMessage{
				Type:      MessageTypeTyping,
				RoomID:    msg.RoomID,
				SenderID:  msg.SenderID,
				Content:   msg.Content,
				CreatedAt: time.Now().Format(time.RFC3339),
			}

			func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Output(nil, fmt.Errorf("panic recovered in broadcast: %v", r))
					}
				}()
				if c.Hub != nil {
					c.Hub.BroadcastToRoom(ctx, msg.RoomID, typingMsg)
				}
			}()

		case MessageTypeUserStatus:
			statusMsg := WebSocketMessage{
				Type:      MessageTypeUserStatus,
				SenderID:  msg.SenderID,
				Content:   msg.Content,
				CreatedAt: time.Now().Format(time.RFC3339),
			}

			// Broadcast user status to all connected clients
			func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Output(nil, fmt.Errorf("panic recovered in broadcast: %v", r))
					}
				}()
				if c.Hub != nil {
					statusBytes, err := json.Marshal(statusMsg)
					if err != nil {
						logger.Output(nil, fmt.Errorf("error marshaling status message: %v", err))
						return
					}
					c.Hub.Broadcast <- statusBytes
				}
			}()

		default:
			logger.Output(nil, fmt.Errorf("unknown message type: %s", msg.Type))
		}
	}
}

func (c *Client) WritePump(ctx context.Context) {
	ctx, span := c.Hub.tracer.Start(ctx, "Client.WritePump")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	ticker := time.NewTicker(30 * time.Second)

	defer func() {
		if r := recover(); r != nil {
			logger.Output(nil, fmt.Errorf("panic recovered in WritePump: %v", r))
		}
		ticker.Stop()
		if c.Conn != nil {
			c.Conn.Close()
		}
	}()

	// Check if connection is valid
	if c.Conn == nil {
		logger.Output(nil, fmt.Errorf("connection is nil"))
		return
	}

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// Send channel was closed
				if c.Conn != nil {
					if err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
						logger.Output(nil, fmt.Errorf("error sending close message: %v", err))
					}
				}
				return
			}

			if c.Conn != nil {
				c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
					logger.Output(nil, fmt.Errorf("error writing message: %v", err))
					return
				}
			}

		case <-ticker.C:
			if c.Conn != nil {
				c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					logger.Output(nil, fmt.Errorf("error writing ping message: %v", err))
					return
				}
			}
		}
	}
}
