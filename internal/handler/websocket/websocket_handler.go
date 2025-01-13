package websocket

import (
	"vongga_api/internal/domain"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type WebSocketMessage struct {
	Type    string      `json:"type"` // "message", "typing", "userStatus"
	Payload interface{} `json:"payload"`
}

type WebSocketHandler struct {
	hub *Hub
}

func NewWebSocketHandler() *WebSocketHandler {
	hub := NewHub()
	go hub.Run()
	return &WebSocketHandler{
		hub: NewHub(),
	}
}

func (h *WebSocketHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/socket", h.handleWebSocket())
	// other websocket routes
}

// WebSocket Handlers
func (h *WebSocketHandler) handleWebSocket() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		claims := c.Locals("user").(*domain.Claims)
		client := NewClient(
			claims.UserID,
			c,
			h.hub,
		)

		h.hub.Register <- client

		go client.writePump()
		client.readPump()
	})
}
