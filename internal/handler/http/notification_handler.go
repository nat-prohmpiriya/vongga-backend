package handler

import (
	"vongga_api/internal/domain"
	"vongga_api/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
)

type NotificationHandler struct {
	notificationUseCase domain.NotificationUseCase
	tracer              trace.Tracer
}

func NewNotificationHandler(router fiber.Router, notificationUseCase domain.NotificationUseCase, tracer trace.Tracer) *NotificationHandler {
	handler := &NotificationHandler{
		notificationUseCase: notificationUseCase,
		tracer:              tracer,
	}

	router.Get("", handler.FindManyNotifications)
	router.Get("/unread-count", handler.FindUnreadCount)
	router.Get("/:id", handler.FindNotification)
	router.Post("/:id/read", handler.MarkAsRead)
	router.Post("/read-all", handler.MarkAllAsRead)
	router.Delete("/:id", handler.DeleteNotification)

	return handler
}

// FindManyNotifications godoc
// @Summary FindMany notifications for the authenticated user
// @Description Find a list of notifications with pagination
// @Tags notifications
// @Accept json
// @Produce json
// @Param limit query int false "Number of items to return (default 10)"
// @Param offset query int false "Number of items to skip (default 0)"
// @Success 200 {array} domain.Notification
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Router /notifications [get]
// @Security BearerAuth
func (h *NotificationHandler) FindManyNotifications(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "NotificationHandler.FindManyNotifications")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	limit := utils.FindQueryInt(c, "limit", 10)
	offset := utils.FindQueryInt(c, "offset", 0)
	logger.Input(map[string]interface{}{
		"userID": userID,
		"limit":  limit,
		"offset": offset,
	})

	notifications, err := h.notificationUseCase.FindManyNotifications(ctx, userID, limit, offset)
	if err != nil {
		logger.Output("error finding notifications 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(notifications, nil)
	return c.JSON(notifications)
}

// FindNotification godoc
// @Summary Find a specific notification
// @Description Find details of a specific notification
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} domain.Notification
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /notifications/{id} [get]
// @Security BearerAuth
func (h *NotificationHandler) FindNotification(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "NotificationHandler.FindNotification")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	notificationID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return utils.HandleError(c, domain.ErrInvalidID)
	}

	logger.Input(map[string]interface{}{
		"userID":         userID,
		"notificationID": notificationID,
	})

	notification, err := h.notificationUseCase.FindNotification(ctx, notificationID)
	if err != nil {
		logger.Output("error finding notification 3", err)
		return utils.HandleError(c, err)
	}

	// Verify that the user owns this notification
	if notification.RecipientID != userID {
		return utils.HandleError(c, domain.ErrUnauthorized)
	}

	logger.Output(notification, nil)
	return c.JSON(notification)
}

// MarkAsRead godoc
// @Summary Mark a notification as read
// @Description Mark a specific notification as read
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /notifications/{id}/read [post]
// @Security BearerAuth
func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "NotificationHandler.MarkAsRead")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	notificationID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return utils.HandleError(c, domain.ErrInvalidID)
	}

	logger.Input(map[string]interface{}{
		"userID":         userID,
		"notificationID": notificationID,
	})

	// Verify ownership before marking as read
	notification, err := h.notificationUseCase.FindNotification(ctx, notificationID)
	if err != nil {
		logger.Output("error finding notification 3", err)
		return utils.HandleError(c, err)
	}

	if notification.RecipientID != userID {
		return utils.HandleError(c, domain.ErrUnauthorized)
	}

	err = h.notificationUseCase.MarkAsRead(ctx, notificationID)
	if err != nil {
		logger.Output("error marking notification as read 4", err)
		return utils.HandleError(c, err)
	}

	logger.Output(nil, nil)
	return c.JSON(utils.SuccessResponse{
		Message: "Notification marked as read",
	})
}

// MarkAllAsRead godoc
// @Summary Mark all notifications as read
// @Description Mark all notifications for the authenticated user as read
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} utils.SuccessResponse
// @Failure 401 {object} utils.ErrorResponse
// @Router /notifications/read-all [post]
// @Security BearerAuth
func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "NotificationHandler.MarkAllAsRead")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Input(map[string]interface{}{
		"userID": userID,
	})

	err = h.notificationUseCase.MarkAllAsRead(ctx, userID)
	if err != nil {
		logger.Output("error marking all notifications as read 2", err)
		return utils.HandleError(c, err)
	}

	logger.Output(nil, nil)
	return c.JSON(utils.SuccessResponse{
		Message: "All notifications marked as read",
	})
}

// DeleteNotification godoc
// @Summary Delete a notification
// @Description Delete a specific notification
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /notifications/{id} [delete]
// @Security BearerAuth
func (h *NotificationHandler) DeleteNotification(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "NotificationHandler.DeleteNotification")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	notificationID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return utils.HandleError(c, domain.ErrInvalidID)
	}

	logger.Input(map[string]interface{}{
		"userID":         userID,
		"notificationID": notificationID,
	})

	// Verify ownership before deletion
	notification, err := h.notificationUseCase.FindNotification(ctx, notificationID)
	if err != nil {
		logger.Output("error finding notification 3", err)
		return utils.HandleError(c, err)
	}

	if notification.RecipientID != userID {
		return utils.HandleError(c, domain.ErrUnauthorized)
	}

	err = h.notificationUseCase.DeleteNotification(ctx, notificationID)
	if err != nil {
		logger.Output("error deleting notification 4", err)
		return utils.HandleError(c, err)
	}

	logger.Output(nil, nil)
	return c.JSON(utils.SuccessResponse{
		Message: "Notification deleted successfully",
	})
}

// FindUnreadCount godoc
// @Summary Find count of unread notifications
// @Description Find the number of unread notifications for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} map[string]int64
// @Failure 401 {object} utils.ErrorResponse
// @Router /notifications/unread-count [get]
// @Security BearerAuth
func (h *NotificationHandler) FindUnreadCount(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "NotificationHandler.FindUnreadCount")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Input(map[string]interface{}{
		"userID": userID,
	})

	count, err := h.notificationUseCase.FindUnreadCount(ctx, userID)
	if err != nil {
		logger.Output("error finding unread count 2", err)
		return utils.HandleError(c, err)
	}

	logger.Output(count, nil)
	return c.JSON(fiber.Map{
		"count": count,
	})
}
