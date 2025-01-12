package websocket

import (
	"context"
	"time"

	"fmt"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"go.opentelemetry.io/otel/trace"
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
	tracer      trace.Tracer
}

func NewWebSocketHandler(router fiber.Router, chatUsecase domain.ChatUsecase, authClient domain.AuthClient, tracer trace.Tracer) {
	handler := &WebSocketHandler{
		chatUsecase: chatUsecase,
		hub:         NewHub(chatUsecase, tracer),
		authClient:  authClient,
		tracer:      tracer,
	}

	// Start WebSocket hub
	go handler.hub.Run(context.Background())

	// WebSocket endpoint with custom middleware for WebSocket authentication
	router.Get("/ws", websocket.New(handler.handleWebSocket, websocket.Config{
		HandshakeTimeout:  10 * time.Second,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
	}))
}

func (h *WebSocketHandler) handleWebSocket(ws *websocket.Conn) {
	ctx := context.Background()
	ctx, span := h.tracer.Start(ctx, "WebSocketHandler.handleWebSocket")
	logger := utils.NewTraceLogger(span)

	defer func() {
		if r := recover(); r != nil {
			logger.Output("panic recovered in handleWebSocket", fmt.Errorf("%v", r))
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

	// Find token from query parameter
	token := ws.Query("token")
	logger.Input(token)
	if token == "" {
		logger.Output(nil, fmt.Errorf("missing token"))
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
		logger.Output(nil, err)
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

	logger.Output(claims, nil)
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

	logger.Info(map[string]interface{}{
		"userID": userID,
		"status": "client_registered",
	})

	// Start ReadPump in main goroutine to keep ws.Conn alive
	go client.WritePump(ctx)
	client.ReadPump(ctx) // This blocks until connection is closed
}
