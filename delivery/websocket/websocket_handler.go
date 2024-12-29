package websocket

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
)

type WebSocketHandler struct {
	chatUsecase domain.ChatUsecase
	hub         *Hub
	authClient  domain.AuthClient
}

func NewWebSocketHandler(router fiber.Router, chatUsecase domain.ChatUsecase, authClient domain.AuthClient) {
	handler := &WebSocketHandler{
		chatUsecase: chatUsecase,
		hub:         NewHub(chatUsecase),
		authClient:  authClient,
	}

	// Start WebSocket hub
	go handler.hub.Run()

	// WebSocket endpoint with custom middleware for WebSocket authentication
	router.Get("/ws", websocket.New(handler.handleWebSocket, websocket.Config{
		HandshakeTimeout: 10 * time.Second,
	}))
}

func (h *WebSocketHandler) handleWebSocket(c *websocket.Conn) {
	// Get token from query parameter
	token := c.Query("token")
	if token == "" {
		c.Close()
		return
	}

	// Verify token
	claims, err := h.authClient.VerifyToken(token)
	if err != nil {
		c.Close()
		return
	}

	userID := claims.UserID

	// Create new client
	client := &Client{
		ID:      utils.GenerateID(),
		UserID:  userID,
		Conn:    c,
		Send:    make(chan []byte, 256),
		Hub:     h.hub,
		RoomIDs: make(map[string]bool),
	}

	// Register client
	h.hub.Register <- client

	// Start goroutines for reading and writing
	go client.WritePump()
	go client.ReadPump()
}
