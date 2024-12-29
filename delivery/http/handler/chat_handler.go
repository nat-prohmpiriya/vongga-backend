package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	ws "github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/delivery/websocket"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
)

type ChatHandler struct {
	chatUsecase domain.ChatUsecase
	hub         *ws.Hub
}

func NewChatHandler(router fiber.Router, chatUsecase domain.ChatUsecase) {
	handler := &ChatHandler{
		chatUsecase: chatUsecase,
		hub:         ws.NewHub(chatUsecase),
	}

	// Start WebSocket hub
	go handler.hub.Run()

	chatGroup := router.Group("/api/chat")
	{
		// WebSocket endpoint
		chatGroup.Get("/ws", websocket.New(handler.handleWebSocket))

		// Room endpoints
		chatGroup.Post("/rooms/private", handler.CreatePrivateChat)
		chatGroup.Post("/rooms/group", handler.CreateGroupChat)
		chatGroup.Get("/rooms", handler.GetUserChats)
		chatGroup.Post("/rooms/:roomId/members", handler.AddMemberToGroup)
		chatGroup.Delete("/rooms/:roomId/members/:userId", handler.RemoveMemberFromGroup)

		// Message endpoints
		chatGroup.Post("/messages", handler.SendMessage)
		chatGroup.Post("/messages/file", handler.SendFileMessage)
		chatGroup.Get("/rooms/:roomId/messages", handler.GetChatMessages)
		chatGroup.Put("/messages/:messageId/read", handler.MarkMessageRead)

		// User status endpoints
		chatGroup.Put("/status", handler.UpdateUserStatus)
		chatGroup.Get("/status/:userId", handler.GetUserStatus)

		// Notification endpoints
		chatGroup.Get("/notifications", handler.GetUserNotifications)
		chatGroup.Put("/notifications/:notificationId/read", handler.MarkNotificationRead)
	}
}

// Room handlers
func (h *ChatHandler) CreatePrivateChat(c *fiber.Ctx) error {
	var req struct {
		UserID1 string `json:"userId1" binding:"required"`
		UserID2 string `json:"userId2" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	logger := utils.NewLogger("ChatHandler.CreatePrivateChat")
	logger.LogInput(map[string]string{
		"userID1": req.UserID1,
		"userID2": req.UserID2,
	})

	room, err := h.chatUsecase.CreatePrivateChat(req.UserID1, req.UserID2)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(room, nil)
	return c.JSON(room)
}

func (h *ChatHandler) CreateGroupChat(c *fiber.Ctx) error {
	var req struct {
		Name      string   `json:"name" binding:"required"`
		MemberIDs []string `json:"memberIds" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	logger := utils.NewLogger("ChatHandler.CreateGroupChat")
	logger.LogInput(map[string]interface{}{
		"name":      req.Name,
		"memberIDs": req.MemberIDs,
	})

	room, err := h.chatUsecase.CreateGroupChat(req.Name, req.MemberIDs)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(room, nil)
	return c.JSON(room)
}

func (h *ChatHandler) GetUserChats(c *fiber.Ctx) error {
	logger := utils.NewLogger("ChatHandler.GetUserChats")
	userID := c.Locals("userId").(string)
	logger.LogInput(userID)

	rooms, err := h.chatUsecase.GetUserChats(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(rooms, nil)
	return c.JSON(rooms)
}

func (h *ChatHandler) AddMemberToGroup(c *fiber.Ctx) error {
	logger := utils.NewLogger("ChatHandler.AddMemberToGroup")
	roomID := c.Params("roomId")

	var req struct {
		UserID string `json:"userId" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	logger.LogInput(map[string]string{
		"roomID": roomID,
		"userID": req.UserID,
	})

	if err := h.chatUsecase.AddMemberToGroup(roomID, req.UserID); err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(nil, nil)
	return c.SendStatus(fiber.StatusOK)
}

func (h *ChatHandler) RemoveMemberFromGroup(c *fiber.Ctx) error {
	logger := utils.NewLogger("ChatHandler.RemoveMemberFromGroup")
	roomID := c.Params("roomId")
	userID := c.Params("userId")

	logger.LogInput(map[string]string{
		"roomID": roomID,
		"userID": userID,
	})

	if err := h.chatUsecase.RemoveMemberFromGroup(roomID, userID); err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(nil, nil)
	return c.SendStatus(fiber.StatusOK)
}

// Message handlers
func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	logger := utils.NewLogger("ChatHandler.SendMessage")
	var req struct {
		RoomID  string `json:"roomId" binding:"required"`
		Content string `json:"content" binding:"required"`
		Type    string `json:"type" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	senderID := c.Locals("userId").(string)
	logger.LogInput(map[string]string{
		"roomID":   req.RoomID,
		"senderID": senderID,
		"type":     req.Type,
		"content":  req.Content,
	})

	message, err := h.chatUsecase.SendMessage(req.RoomID, senderID, req.Type, req.Content)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(message, nil)
	return c.JSON(message)
}

func (h *ChatHandler) SendFileMessage(c *fiber.Ctx) error {
	logger := utils.NewLogger("ChatHandler.SendFileMessage")
	var req struct {
		RoomID   string `json:"roomId" binding:"required"`
		FileType string `json:"fileType" binding:"required"`
		FileSize int64  `json:"fileSize" binding:"required"`
		FileURL  string `json:"fileUrl" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	senderID := c.Locals("userId").(string)
	logger.LogInput(map[string]interface{}{
		"roomID":   req.RoomID,
		"senderID": senderID,
		"fileType": req.FileType,
		"fileSize": req.FileSize,
		"fileURL":  req.FileURL,
	})

	message, err := h.chatUsecase.SendFileMessage(req.RoomID, senderID, req.FileType, req.FileSize, req.FileURL)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(message, nil)
	return c.JSON(message)
}

func (h *ChatHandler) GetChatMessages(c *fiber.Ctx) error {
	logger := utils.NewLogger("ChatHandler.GetChatMessages")
	roomID := c.Params("roomId")
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	logger.LogInput(map[string]interface{}{
		"roomID": roomID,
		"limit":  limit,
		"offset": offset,
	})

	messages, err := h.chatUsecase.GetChatMessages(roomID, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(messages, nil)
	return c.JSON(messages)
}

func (h *ChatHandler) MarkMessageRead(c *fiber.Ctx) error {
	logger := utils.NewLogger("ChatHandler.MarkMessageRead")
	messageID := c.Params("messageId")
	userID := c.Locals("userId").(string)

	logger.LogInput(map[string]string{
		"messageID": messageID,
		"userID":    userID,
	})

	if err := h.chatUsecase.MarkMessageRead(messageID, userID); err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(nil, nil)
	return c.SendStatus(fiber.StatusOK)
}

// User status handlers
func (h *ChatHandler) UpdateUserStatus(c *fiber.Ctx) error {
	logger := utils.NewLogger("ChatHandler.UpdateUserStatus")
	var req struct {
		IsOnline bool `json:"isOnline" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	userID := c.Locals("userId").(string)
	logger.LogInput(map[string]interface{}{
		"userID":   userID,
		"isOnline": req.IsOnline,
	})

	if err := h.chatUsecase.UpdateUserOnlineStatus(userID, req.IsOnline); err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(nil, nil)
	return c.SendStatus(fiber.StatusOK)
}

func (h *ChatHandler) GetUserStatus(c *fiber.Ctx) error {
	logger := utils.NewLogger("ChatHandler.GetUserStatus")
	userID := c.Params("userId")
	logger.LogInput(userID)

	status, err := h.chatUsecase.GetUserOnlineStatus(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(status, nil)
	return c.JSON(status)
}

// Notification handlers
func (h *ChatHandler) GetUserNotifications(c *fiber.Ctx) error {
	logger := utils.NewLogger("ChatHandler.GetUserNotifications")
	userID := c.Locals("userId").(string)
	logger.LogInput(userID)

	notifications, err := h.chatUsecase.GetUserNotifications(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(notifications, nil)
	return c.JSON(notifications)
}

func (h *ChatHandler) MarkNotificationRead(c *fiber.Ctx) error {
	logger := utils.NewLogger("ChatHandler.MarkNotificationRead")
	notificationID := c.Params("notificationId")
	logger.LogInput(notificationID)

	if err := h.chatUsecase.MarkNotificationRead(notificationID); err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(nil, nil)
	return c.SendStatus(fiber.StatusOK)
}

func (h *ChatHandler) handleWebSocket(c *websocket.Conn) {
	// Get user ID from context
	userID := c.Locals("userId").(string)

	// Create new client
	client := &ws.Client{
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
