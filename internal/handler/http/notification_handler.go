package handler

import (
	"vongga_api/internal/domain"
	"vongga_api/internal/dto"
	"vongga_api/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

type NotificationHandler struct {
	notificationUseCase domain.NotificationUseCase
	tracer              trace.Tracer
	validate            *validator.Validate
}

func NewNotificationHandler(router fiber.Router, notificationUseCase domain.NotificationUseCase, tracer trace.Tracer) *NotificationHandler {
	handler := &NotificationHandler{
		notificationUseCase: notificationUseCase,
		tracer:              tracer,
		validate:            validator.New(),
	}

	router.Get("", handler.FindManyNotifications)
	router.Get("/unread-count", handler.FindUnreadCount)
	router.Get("/:id", handler.FindNotification)
	router.Post("/:id/read", handler.MarkAsRead)
	router.Post("/read-all", handler.MarkAllAsRead)
	router.Delete("/:id", handler.DeleteNotification)

	return handler
}

func (h *NotificationHandler) FindManyNotifications(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "NotificationHandler.FindManyNotifications")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID := c.Locals("userId").(string)

	var req dto.FindManyNotificationsRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid query parameters",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "validation failed",
		})
	}

	logger.Input(map[string]interface{}{
		"userID": userID,
		"limit":  req.Limit,
		"offset": req.Offset,
	})

	notifications, err := h.notificationUseCase.FindManyNotifications(ctx, userID, req.Limit, req.Offset)
	if err != nil {
		logger.Output("error finding notifications", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(notifications, nil)
	return c.JSON(notifications)
}

func (h *NotificationHandler) FindNotification(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "NotificationHandler.FindNotification")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID := c.Locals("userID").(string)

	var req dto.FindNotificationRequest
	if err := c.ParamsParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid notification id",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "validation failed",
		})
	}

	logger.Input(map[string]interface{}{
		"userID":         userID,
		"notificationID": req.ID,
	})

	notification, err := h.notificationUseCase.FindNotification(ctx, req.ID)
	if err != nil {
		logger.Output("error finding notification", err)
		return utils.HandleError(c, err)
	}

	// Verify that the user owns this notification
	if notification.RecipientID != userID {
		return utils.HandleError(c, domain.ErrUnauthorized)
	}

	logger.Output(notification, nil)
	return c.JSON(notification)
}

func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "NotificationHandler.MarkAsRead")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID := c.Locals("userID").(string)

	var req dto.MarkAsReadRequest
	if err := c.ParamsParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid notification id",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "validation failed",
		})
	}

	logger.Input(map[string]interface{}{
		"userID":         userID,
		"notificationID": req.ID,
	})

	// Verify ownership before marking as read
	notification, err := h.notificationUseCase.FindNotification(ctx, req.ID)
	if err != nil {
		logger.Output("error finding notification", err)
		return utils.HandleError(c, err)
	}

	if notification.RecipientID != userID {
		return utils.HandleError(c, domain.ErrUnauthorized)
	}

	err = h.notificationUseCase.MarkAsRead(ctx, req.ID)
	if err != nil {
		logger.Output("error marking notification as read", err)
		return utils.HandleError(c, err)
	}

	logger.Output(nil, nil)
	return c.JSON(utils.SuccessResponse{
		Message: "Notification marked as read",
	})
}

func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "NotificationHandler.MarkAllAsRead")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID := c.Locals("userID").(string)

	logger.Input(map[string]interface{}{
		"userID": userID,
	})

	err := h.notificationUseCase.MarkAllAsRead(ctx, userID)
	if err != nil {
		logger.Output("error marking all notifications as read", err)
		return utils.HandleError(c, err)
	}

	logger.Output(nil, nil)
	return c.JSON(utils.SuccessResponse{
		Message: "All notifications marked as read",
	})
}

func (h *NotificationHandler) DeleteNotification(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "NotificationHandler.DeleteNotification")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID := c.Locals("userID").(string)

	var req dto.DeleteNotificationRequest
	if err := c.ParamsParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid notification id",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "validation failed",
		})
	}

	logger.Input(map[string]interface{}{
		"userID":         userID,
		"notificationID": req.ID,
	})

	// Verify ownership before deletion
	notification, err := h.notificationUseCase.FindNotification(ctx, req.ID)
	if err != nil {
		logger.Output("error finding notification", err)
		return utils.HandleError(c, err)
	}

	if notification.RecipientID != userID {
		return utils.HandleError(c, domain.ErrUnauthorized)
	}

	err = h.notificationUseCase.DeleteNotification(ctx, req.ID)
	if err != nil {
		logger.Output("error deleting notification", err)
		return utils.HandleError(c, err)
	}

	logger.Output(nil, nil)
	return c.JSON(utils.SuccessResponse{
		Message: "Notification deleted successfully",
	})
}

func (h *NotificationHandler) FindUnreadCount(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "NotificationHandler.FindUnreadCount")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID := c.Locals("userID").(string)

	logger.Input(map[string]interface{}{
		"userID": userID,
	})

	count, err := h.notificationUseCase.FindUnreadCount(ctx, userID)
	if err != nil {
		logger.Output("error finding unread count", err)
		return utils.HandleError(c, err)
	}

	logger.Output(count, nil)
	return c.JSON(fiber.Map{
		"count": count,
	})
}
