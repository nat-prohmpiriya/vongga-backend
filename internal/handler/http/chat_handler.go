package handler

import (
	"strconv"

	"vongga_api/internal/domain"
	"vongga_api/utils"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

type ChatHandler struct {
	chatUsecase domain.ChatUsecase
	tracer      trace.Tracer
}

func NewChatHandler(router fiber.Router, chatUsecase domain.ChatUsecase, tracer trace.Tracer) {
	handler := &ChatHandler{
		chatUsecase: chatUsecase,
		tracer:      tracer,
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
	ctx, span := h.tracer.Start(c.UserContext(), "ChatHandler.CreatePrivateChat")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	var req struct {
		UserID1 string `json:"userId1" binding:"required"`
		UserID2 string `json:"userId2" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		logger.Output("error parsing request 1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	logger.Input(map[string]string{
		"userID1": req.UserID1,
		"userID2": req.UserID2,
	})

	room, err := h.chatUsecase.CreatePrivateChat(ctx, req.UserID1, req.UserID2)
	if err != nil {
		logger.Output("error creating private chat 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(room, nil)
	return c.JSON(room)
}

func (h *ChatHandler) CreateGroupChat(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "ChatHandler.CreateGroupChat")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	var req struct {
		Name      string   `json:"name" binding:"required"`
		MemberIDs []string `json:"memberIds" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		logger.Output("error parsing request 1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	logger.Input(map[string]interface{}{
		"name":      req.Name,
		"memberIDs": req.MemberIDs,
	})

	room, err := h.chatUsecase.CreateGroupChat(ctx, req.Name, req.MemberIDs)
	if err != nil {
		logger.Output("error creating group chat 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(room, nil)
	return c.JSON(room)
}

func (h *ChatHandler) FindUserChats(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "ChatHandler.FindUserChats")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Input(userID.Hex())

	rooms, err := h.chatUsecase.FindUserChats(ctx, userID.Hex())
	if err != nil {
		logger.Output("error finding user chats 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(rooms, nil)
	return c.JSON(rooms)
}

func (h *ChatHandler) AddMemberToGroup(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "ChatHandler.AddMemberToGroup")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	roomID := c.Params("roomId")

	var req struct {
		UserID string `json:"userId" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		logger.Output("error parsing request 1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	logger.Input(map[string]string{
		"roomID": roomID,
		"userID": req.UserID,
	})

	if err := h.chatUsecase.AddMemberToGroup(ctx, roomID, req.UserID); err != nil {
		logger.Output("error adding member to group 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(nil, nil)
	return c.SendStatus(fiber.StatusOK)
}

func (h *ChatHandler) RemoveMemberFromGroup(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "ChatHandler.RemoveMemberFromGroup")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	roomID := c.Params("roomId")
	userID := c.Params("userId")

	logger.Input(map[string]string{
		"roomID": roomID,
		"userID": userID,
	})

	if err := h.chatUsecase.RemoveMemberFromGroup(ctx, roomID, userID); err != nil {
		logger.Output("error removing member from group 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(nil, nil)
	return c.SendStatus(fiber.StatusOK)
}

// Message handlers
func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "ChatHandler.SendMessage")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	senderID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding sender ID 1", err)
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
		logger.Output("error parsing request 1", err)
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

	message, err := h.chatUsecase.SendMessage(ctx, req.RoomID, senderID.Hex(), req.Type, req.Content)
	if err != nil {
		logger.Output("error sending message 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(message, nil)
	return c.JSON(message)
}

func (h *ChatHandler) SendFileMessage(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "ChatHandler.SendFileMessage")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	senderID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding sender ID 1", err)
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
		logger.Output("error parsing request 1", err)
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

	message, err := h.chatUsecase.SendFileMessage(ctx, req.RoomID, senderID.Hex(), req.FileType, req.FileSize, req.FileURL)
	if err != nil {
		logger.Output("error sending file message 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(message, nil)
	return c.JSON(message)
}

func (h *ChatHandler) FindChatMessages(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "ChatHandler.FindChatMessages")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	roomID := c.Params("roomId")
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	logger.Input(map[string]interface{}{
		"roomID": roomID,
		"limit":  limit,
		"offset": offset,
	})

	messages, err := h.chatUsecase.FindChatMessages(ctx, roomID, limit, offset)
	if err != nil {
		logger.Output("error finding chat messages 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(messages, nil)
	return c.JSON(messages)
}

func (h *ChatHandler) MarkMessageRead(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "ChatHandler.MarkMessageRead")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	messageID := c.Params("messageId")

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Input(map[string]interface{}{
		"messageID": messageID,
		"userID":    userID.Hex(),
	})

	if err := h.chatUsecase.MarkMessageRead(ctx, messageID, userID.Hex()); err != nil {
		logger.Output("error marking message read 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(nil, nil)
	return c.SendStatus(fiber.StatusOK)
}

// User status handlers
func (h *ChatHandler) UpdateUserStatus(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "ChatHandler.UpdateUserStatus")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var req struct {
		IsOnline bool `json:"isOnline" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		logger.Output("error parsing request 1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	logger.Input(map[string]interface{}{
		"userID":   userID.Hex(),
		"isOnline": req.IsOnline,
	})

	if err := h.chatUsecase.UpdateUserOnlineStatus(ctx, userID.Hex(), req.IsOnline); err != nil {
		logger.Output("error updating user status 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(nil, nil)
	return c.SendStatus(fiber.StatusOK)
}

func (h *ChatHandler) FindUserStatus(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "ChatHandler.FindUserStatus")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID := c.Params("userId")
	logger.Input(userID)

	status, err := h.chatUsecase.FindUserOnlineStatus(ctx, userID)
	if err != nil {
		logger.Output("error finding user status 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(status, nil)
	return c.JSON(status)
}

// Notification handlers
func (h *ChatHandler) FindUserNotifications(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "ChatHandler.FindUserNotifications")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Input(userID.Hex())

	notifications, err := h.chatUsecase.FindUserNotifications(ctx, userID.Hex())
	if err != nil {
		logger.Output("error finding user notifications 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(notifications, nil)
	return c.JSON(notifications)
}

func (h *ChatHandler) MarkNotificationRead(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "ChatHandler.MarkNotificationRead")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	notificationID := c.Params("notificationId")
	logger.Input(notificationID)

	if err := h.chatUsecase.MarkNotificationRead(ctx, notificationID); err != nil {
		logger.Output("error marking notification read 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(nil, nil)
	return c.SendStatus(fiber.StatusOK)
}
