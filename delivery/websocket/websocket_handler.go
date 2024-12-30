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
		HandshakeTimeout: 10 * time.Second,
	}))
}

func (h *WebSocketHandler) handleWebSocket(ws *websocket.Conn) {
	logger := utils.NewLogger("WebSocketHandler.handleWebSocket")
	// Get token from query parameter
	token := ws.Query("token")
	logger.LogInput(token)
	if token == "" {
		logger.LogOutput(nil, fmt.Errorf("missing token"))
		ws.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(
				websocket.CloseInvalidFramePayloadData, // code 1007
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
				websocket.ClosePolicyViolation, // code 1008
				"Invalid or expired token",
			),
			time.Now().Add(time.Second),
		)

		ws.Close()
		return
	}

	logger.LogOutput(claims, nil)

	userID := claims.UserID

	// Create new client
	client := &Client{
		ID:      utils.GenerateID(),
		UserID:  userID,
		Conn:    ws,
		Send:    make(chan []byte, 256),
		Hub:     h.hub,
		RoomIDs: make(map[string]bool),
	}

	// ตรวจสอบ nil
	if client.Conn == nil {
		logger.LogOutput(nil, fmt.Errorf("Client connection is nil"))
		ws.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(
				websocket.CloseInternalServerErr, // code 1011
				"Client connection is nil",
			),
			time.Now().Add(time.Second),
		)
		ws.Close()
		return
	}

	// ตรวจสอบ nil ก่อนใช้ Hub
	if h.hub == nil {
		logger.LogOutput(nil, fmt.Errorf("Hub is nil"))
		ws.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(
				websocket.CloseInternalServerErr, // code 1011
				"Hub is nil",
			),
			time.Now().Add(time.Second),
		)
		ws.Close()
		return
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
