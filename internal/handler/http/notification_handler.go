package handler

import (
	"vongga-api/internal/domain"
	"vongga-api/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationHandler struct {
	notificationUseCase domain.NotificationUseCase
}

func NewNotificationHandler(router fiber.Router, notificationUseCase domain.NotificationUseCase) *NotificationHandler {
	handler := &NotificationHandler{
		notificationUseCase: notificationUseCase,
	}

	router.Find("/", handler.FindManyNotifications)
	router.Find("/unread-count", handler.FindUnreadCount)
	router.Find("/:id", handler.FindNotification)
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
	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		return utils.HandleError(c, err)
	}

	limit := utils.FindQueryInt(c, "limit", 10)
	offset := utils.FindQueryInt(c, "offset", 0)

	notifications, err := h.notificationUseCase.FindManyNotifications(userID, limit, offset)
	if err != nil {
		return utils.HandleError(c, err)
	}

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
	notificationID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return utils.HandleError(c, domain.ErrInvalidID)
	}

	notification, err := h.notificationUseCase.FindNotification(notificationID)
	if err != nil {
		return utils.HandleError(c, err)
	}

	// Verify that the user owns this notification
	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		return utils.HandleError(c, err)
	}

	if notification.RecipientID != userID {
		return utils.HandleError(c, domain.ErrUnauthorized)
	}

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
	notificationID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return utils.HandleError(c, domain.ErrInvalidID)
	}

	// Verify ownership before marking as read
	notification, err := h.notificationUseCase.FindNotification(notificationID)
	if err != nil {
		return utils.HandleError(c, err)
	}

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		return utils.HandleError(c, err)
	}

	if notification.RecipientID != userID {
		return utils.HandleError(c, domain.ErrUnauthorized)
	}

	err = h.notificationUseCase.MarkAsRead(notificationID)
	if err != nil {
		return utils.HandleError(c, err)
	}

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
	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		return utils.HandleError(c, err)
	}

	err = h.notificationUseCase.MarkAllAsRead(userID)
	if err != nil {
		return utils.HandleError(c, err)
	}

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
	notificationID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return utils.HandleError(c, domain.ErrInvalidID)
	}

	// Verify ownership before deletion
	notification, err := h.notificationUseCase.FindNotification(notificationID)
	if err != nil {
		return utils.HandleError(c, err)
	}

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		return utils.HandleError(c, err)
	}

	if notification.RecipientID != userID {
		return utils.HandleError(c, domain.ErrUnauthorized)
	}

	err = h.notificationUseCase.DeleteNotification(notificationID)
	if err != nil {
		return utils.HandleError(c, err)
	}

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
	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		return utils.HandleError(c, err)
	}

	count, err := h.notificationUseCase.FindUnreadCount(userID)
	if err != nil {
		return utils.HandleError(c, err)
	}

	return c.JSON(fiber.Map{
		"count": count,
	})
}
