package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
)

type ChatHandler struct {
	chatUsecase domain.ChatUsecase
}

func NewChatHandler(router fiber.Router, chatUsecase domain.ChatUsecase) {
	handler := &ChatHandler{
		chatUsecase: chatUsecase,
	}

	// Room endpoints
	router.Post("/rooms/private", handler.CreatePrivateChat)
	router.Post("/rooms/group", handler.CreateGroupChat)
	router.Get("/rooms", handler.GetUserChats)
	router.Post("/rooms/:roomId/members", handler.AddMemberToGroup)
	router.Delete("/rooms/:roomId/members/:userId", handler.RemoveMemberFromGroup)

	// Message endpoints
	router.Post("/messages", handler.SendMessage)
	router.Post("/messages/file", handler.SendFileMessage)
	router.Get("/rooms/:roomId/messages", handler.GetChatMessages)
	router.Put("/messages/:messageId/read", handler.MarkMessageRead)

	// User status endpoints
	router.Put("/status", handler.UpdateUserStatus)
	router.Get("/status/:userId", handler.GetUserStatus)

	// Notification endpoints
	router.Get("/notifications", handler.GetUserNotifications)
	router.Put("/notifications/:notificationId/read", handler.MarkNotificationRead)
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

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogInput(userID.Hex())

	rooms, err := h.chatUsecase.GetUserChats(userID.Hex())
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

	senderID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

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

	logger.LogInput(map[string]string{
		"roomID":   req.RoomID,
		"senderID": senderID.Hex(),
		"type":     req.Type,
		"content":  req.Content,
	})

	message, err := h.chatUsecase.SendMessage(req.RoomID, senderID.Hex(), req.Type, req.Content)
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

	senderID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

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

	logger.LogInput(map[string]interface{}{
		"roomID":   req.RoomID,
		"senderID": senderID.Hex(),
		"fileType": req.FileType,
		"fileSize": req.FileSize,
		"fileURL":  req.FileURL,
	})

	message, err := h.chatUsecase.SendFileMessage(req.RoomID, senderID.Hex(), req.FileType, req.FileSize, req.FileURL)
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

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogInput(map[string]interface{}{
		"messageID": messageID,
		"userID":    userID.Hex(),
	})

	if err := h.chatUsecase.MarkMessageRead(messageID, userID.Hex()); err != nil {
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

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var req struct {
		IsOnline bool `json:"isOnline" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	logger.LogInput(map[string]interface{}{
		"userID":   userID.Hex(),
		"isOnline": req.IsOnline,
	})

	if err := h.chatUsecase.UpdateUserOnlineStatus(userID.Hex(), req.IsOnline); err != nil {
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

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogInput(userID.Hex())

	notifications, err := h.chatUsecase.GetUserNotifications(userID.Hex())
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
