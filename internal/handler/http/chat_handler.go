package handler

import (
	"strconv"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"github.com/gofiber/fiber/v2"
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
	router.Get("/rooms", handler.FindUserChats)
	router.Post("/rooms/:roomId/members", handler.AddMemberToGroup)
	router.Delete("/rooms/:roomId/members/:userId", handler.RemoveMemberFromGroup)

	// Message endpoints
	router.Post("/messages", handler.SendMessage)
	router.Post("/messages/file", handler.SendFileMessage)
	router.Get("/rooms/:roomId/messages", handler.FindChatMessages)
	router.Put("/messages/:messageId/read", handler.MarkMessageRead)

	// User status endpoints
	router.Put("/status", handler.UpdateUserStatus)
	router.Get("/status/:userId", handler.FindUserStatus)

	// Notification endpoints
	router.Get("/notifications", handler.FindUserNotifications)
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

	logger := utils.NewTraceLogger("ChatHandler.CreatePrivateChat")
	logger.Input(map[string]string{
		"userID1": req.UserID1,
		"userID2": req.UserID2,
	})

	room, err := h.chatUsecase.CreatePrivateChat(c, req.UserID1, req.UserID2)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(room, nil)
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

	logger := utils.NewTraceLogger("ChatHandler.CreateGroupChat")
	logger.Input(map[string]interface{}{
		"name":      req.Name,
		"memberIDs": req.MemberIDs,
	})

	room, err := h.chatUsecase.CreateGroupChat(req.Name, req.MemberIDs)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(room, nil)
	return c.JSON(room)
}

func (h *ChatHandler) FindUserChats(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("ChatHandler.FindUserChats")

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Input(userID.Hex())

	rooms, err := h.chatUsecase.FindUserChats(userID.Hex())
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(rooms, nil)
	return c.JSON(rooms)
}

func (h *ChatHandler) AddMemberToGroup(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("ChatHandler.AddMemberToGroup")
	roomID := c.Params("roomId")

	var req struct {
		UserID string `json:"userId" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	logger.Input(map[string]string{
		"roomID": roomID,
		"userID": req.UserID,
	})

	if err := h.chatUsecase.AddMemberToGroup(roomID, req.UserID); err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(nil, nil)
	return c.SendStatus(fiber.StatusOK)
}

func (h *ChatHandler) RemoveMemberFromGroup(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("ChatHandler.RemoveMemberFromGroup")
	roomID := c.Params("roomId")
	userID := c.Params("userId")

	logger.Input(map[string]string{
		"roomID": roomID,
		"userID": userID,
	})

	if err := h.chatUsecase.RemoveMemberFromGroup(roomID, userID); err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(nil, nil)
	return c.SendStatus(fiber.StatusOK)
}

// Message handlers
func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("ChatHandler.SendMessage")

	senderID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output(nil, err)
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
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	logger.Input(map[string]string{
		"roomID":   req.RoomID,
		"senderID": senderID.Hex(),
		"type":     req.Type,
		"content":  req.Content,
	})

	message, err := h.chatUsecase.SendMessage(req.RoomID, senderID.Hex(), req.Type, req.Content)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(message, nil)
	return c.JSON(message)
}

func (h *ChatHandler) SendFileMessage(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("ChatHandler.SendFileMessage")

	senderID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output(nil, err)
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
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	logger.Input(map[string]interface{}{
		"roomID":   req.RoomID,
		"senderID": senderID.Hex(),
		"fileType": req.FileType,
		"fileSize": req.FileSize,
		"fileURL":  req.FileURL,
	})

	message, err := h.chatUsecase.SendFileMessage(req.RoomID, senderID.Hex(), req.FileType, req.FileSize, req.FileURL)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(message, nil)
	return c.JSON(message)
}

func (h *ChatHandler) FindChatMessages(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("ChatHandler.FindChatMessages")
	roomID := c.Params("roomId")
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	logger.Input(map[string]interface{}{
		"roomID": roomID,
		"limit":  limit,
		"offset": offset,
	})

	messages, err := h.chatUsecase.FindChatMessages(roomID, limit, offset)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(messages, nil)
	return c.JSON(messages)
}

func (h *ChatHandler) MarkMessageRead(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("ChatHandler.MarkMessageRead")
	messageID := c.Params("messageId")

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Input(map[string]interface{}{
		"messageID": messageID,
		"userID":    userID.Hex(),
	})

	if err := h.chatUsecase.MarkMessageRead(messageID, userID.Hex()); err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(nil, nil)
	return c.SendStatus(fiber.StatusOK)
}

// User status handlers
func (h *ChatHandler) UpdateUserStatus(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("ChatHandler.UpdateUserStatus")

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var req struct {
		IsOnline bool `json:"isOnline" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	logger.Input(map[string]interface{}{
		"userID":   userID.Hex(),
		"isOnline": req.IsOnline,
	})

	if err := h.chatUsecase.UpdateUserOnlineStatus(userID.Hex(), req.IsOnline); err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(nil, nil)
	return c.SendStatus(fiber.StatusOK)
}

func (h *ChatHandler) FindUserStatus(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("ChatHandler.FindUserStatus")
	userID := c.Params("userId")
	logger.Input(userID)

	status, err := h.chatUsecase.FindUserOnlineStatus(userID)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(status, nil)
	return c.JSON(status)
}

// Notification handlers
func (h *ChatHandler) FindUserNotifications(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("ChatHandler.FindUserNotifications")

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Input(userID.Hex())

	notifications, err := h.chatUsecase.FindUserNotifications(userID.Hex())
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(notifications, nil)
	return c.JSON(notifications)
}

func (h *ChatHandler) MarkNotificationRead(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("ChatHandler.MarkNotificationRead")
	notificationID := c.Params("notificationId")
	logger.Input(notificationID)

	if err := h.chatUsecase.MarkNotificationRead(notificationID); err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(nil, nil)
	return c.SendStatus(fiber.StatusOK)
}
