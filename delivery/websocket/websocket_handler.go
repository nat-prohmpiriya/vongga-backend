package websocket

import (
	"time"

	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
)

const (
	// Token missing
	CloseInvalidFramePayloadData = 1007 // ข้อมูลไม่ถูกต้อง

	// Token invalid/expired
	ClosePolicyViolation = 1008 // ผิด policy (token)

	// Normal closure
	CloseNormalClosure = 1000 // ปิดปกติ

	// Server error
	CloseInternalServerErr = 1011 // server error
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
		HandshakeTimeout:    10 * time.Second,
		ReadBufferSize:     1024,
		WriteBufferSize:    1024,
		EnableCompression:  true,
	}))
}

func (h *WebSocketHandler) handleWebSocket(ws *websocket.Conn) {
	logger := utils.NewLogger("WebSocketHandler.handleWebSocket")
	
	defer func() {
		if r := recover(); r != nil {
			logger.LogOutput(nil, fmt.Errorf("panic recovered in handleWebSocket: %v", r))
			ws.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(
					websocket.CloseInternalServerErr,
					fmt.Sprintf("Internal server error: %v", r),
				),
				time.Now().Add(time.Second),
			)
			ws.Close()
		}
	}()

	// Get token from query parameter
	token := ws.Query("token")
	logger.LogInput(token)
	if token == "" {
		logger.LogOutput(nil, fmt.Errorf("missing token"))
		ws.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(
				websocket.CloseInvalidFramePayloadData,
				"Missing token",
			), time.Now().Add(time.Second),
		)
		ws.Close()
		return
	}

	// Verify token
	claims, err := h.authClient.VerifyToken(token)
	if err != nil {
		logger.LogOutput(nil, err)
		ws.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(
				websocket.ClosePolicyViolation,
				"Invalid or expired token",
			),
			time.Now().Add(time.Second),
		)
		ws.Close()
		return
	}

	logger.LogOutput(claims, nil)
	userID := claims.UserID

	// Create new client with mutex
	client := &Client{
		ID:      utils.GenerateID(),
		UserID:  userID,
		Conn:    ws,
		Send:    make(chan []byte, 256),
		Hub:     h.hub,
		RoomIDs: make(map[string]bool),
	}

	// Register client before starting pumps
	h.hub.Register <- client

	logger.LogInfo(map[string]interface{}{
		"userID": userID,
		"status": "client_registered",
	})

	// Start ReadPump in main goroutine to keep ws.Conn alive
	go client.WritePump()
	client.ReadPump() // This blocks until connection is closed
}
