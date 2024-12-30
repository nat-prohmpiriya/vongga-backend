package websocket

import (
	"time"

	"fmt"

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
	logger := utils.NewLogger("WebSocketHandler.handleWebSocket")
	// Get token from query parameter
	token := c.Query("token")
	logger.LogInput(token)
	if token == "" {
		logger.LogOutput(nil, fmt.Errorf("missing token"))
		c.Close()
		return
	}

	// Verify token
	claims, err := h.authClient.VerifyToken(token)
	if err != nil {
		logger.LogOutput(nil, err)
		c.Close()
		return
	}

	logger.LogOutput(claims, nil)

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

	logger.LogInfo(map[string]interface{}{
		"userID": userID,
		"status": "connected",
	})

	// Register client
	h.hub.Register <- client
	logger.LogInfo(map[string]interface{}{
		"userID": userID,
		"status": "registered",
	})

	// Start goroutines for reading and writing
	go client.WritePump()
	go client.ReadPump()

	logger.LogInfo(map[string]interface{}{
		"userID": userID,
		"status": "pump_started",
	})
}
