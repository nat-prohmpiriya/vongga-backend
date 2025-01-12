package usecase

import (
	"context"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
)

type notificationUseCase struct {
	notificationRepo domain.NotificationRepository
	userRepo         domain.UserRepository
	tracer           trace.Tracer
}

func NewNotificationUseCase(
	notificationRepo domain.NotificationRepository,
	userRepo domain.UserRepository,
	tracer trace.Tracer,
) domain.NotificationUseCase {
	return &notificationUseCase{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		tracer:           tracer,
	}
}

func (n *notificationUseCase) CreateNotification(ctx context.Context, recipientID, senderID, refID primitive.ObjectID, nType domain.NotificationType, refType, message string) (*domain.Notification, error) {
	ctx, span := n.tracer.Start(ctx, "NotificationUseCase.CreateNotification")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"recipientID": recipientID.Hex(),
		"senderID":    senderID.Hex(),
		"refID":       refID.Hex(),
		"type":        nType,
		"refType":     refType,
		"message":     message,
	}
	logger.Input(input)

	notification := &domain.Notification{
		RecipientID: recipientID,
		SenderID:    senderID,
		Type:        nType,
		RefID:       refID,
		RefType:     refType,
		Message:     message,
		IsRead:      false,
	}

	err := n.notificationRepo.Create(ctx, notification)
	if err != nil {
		logger.Output("error creating notification 1", err)
		return nil, err
	}

	logger.Output(notification, nil)
	return notification, nil
}

func (n *notificationUseCase) FindNotification(ctx context.Context, notificationID primitive.ObjectID) (*domain.NotificationResponse, error) {
	ctx, span := n.tracer.Start(ctx, "NotificationUseCase.FindNotification")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	logger.Input(notificationID)

	notification, err := n.notificationRepo.FindByID(ctx, notificationID)
	if err != nil {
		logger.Output("error finding notification 1", err)
		return nil, err
	}

	// Find sender information
	sender, err := n.userRepo.FindByID(ctx, notification.SenderID.Hex())
	if err != nil {
		logger.Output("error finding sender 2", err)
		return nil, err
	}

	response := &domain.NotificationResponse{
		Notification: *notification,
	}
	response.Sender.UserID = sender.ID.Hex()
	response.Sender.Username = sender.Username
	response.Sender.DisplayName = sender.DisplayName
	response.Sender.PhotoProfile = sender.PhotoProfile
	response.Sender.FirstName = sender.FirstName
	response.Sender.LastName = sender.LastName

	logger.Output(response, nil)
	return response, nil
}

func (n *notificationUseCase) FindManyNotifications(ctx context.Context, recipientID primitive.ObjectID, limit, offset int) ([]domain.NotificationResponse, error) {
	ctx, span := n.tracer.Start(ctx, "NotificationUseCase.FindManyNotifications")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"recipientID": recipientID.Hex(),
		"limit":       limit,
		"offset":      offset,
	}
	logger.Input(input)

	notifications, err := n.notificationRepo.FindByRecipient(ctx, recipientID, limit, offset)
	if err != nil {
		logger.Output("error finding notifications 1", err)
		return nil, err
	}

	// Create response with user information
	response := make([]domain.NotificationResponse, len(notifications))
	for i, notification := range notifications {
		sender, err := n.userRepo.FindByID(ctx, notification.SenderID.Hex())
		if err != nil {
			logger.Output("error finding sender 2", err)
			continue
		}

		response[i] = domain.NotificationResponse{
			Notification: notification,
		}
		response[i].Sender.UserID = sender.ID.Hex()
		response[i].Sender.Username = sender.Username
		response[i].Sender.DisplayName = sender.DisplayName
		response[i].Sender.PhotoProfile = sender.PhotoProfile
		response[i].Sender.FirstName = sender.FirstName
		response[i].Sender.LastName = sender.LastName
	}

	logger.Output(response, nil)
	return response, nil
}

func (n *notificationUseCase) MarkAsRead(ctx context.Context, notificationID primitive.ObjectID) error {
	ctx, span := n.tracer.Start(ctx, "NotificationUseCase.MarkAsRead")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"notificationID": notificationID.Hex(),
	}
	logger.Input(input)

	err := n.notificationRepo.MarkAsRead(ctx, notificationID)
	if err != nil {
		logger.Output("error marking notification as read 1", err)
		return err
	}

	logger.Output("notification marked as read successfully", nil)
	return nil
}

func (n *notificationUseCase) MarkAllAsRead(ctx context.Context, recipientID primitive.ObjectID) error {
	ctx, span := n.tracer.Start(ctx, "NotificationUseCase.MarkAllAsRead")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"recipientID": recipientID.Hex(),
	}
	logger.Input(input)

	err := n.notificationRepo.MarkAllAsRead(ctx, recipientID)
	if err != nil {
		logger.Output("error marking all notifications as read 1", err)
		return err
	}

	logger.Output("all notifications marked as read successfully", nil)
	return nil
}

func (n *notificationUseCase) DeleteNotification(ctx context.Context, notificationID primitive.ObjectID) error {
	ctx, span := n.tracer.Start(ctx, "NotificationUseCase.DeleteNotification")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"notificationID": notificationID.Hex(),
	}
	logger.Input(input)

	err := n.notificationRepo.Delete(ctx, notificationID)
	if err != nil {
		logger.Output("error deleting notification 1", err)
		return err
	}

	logger.Output("notification deleted successfully", nil)
	return nil
}

func (n *notificationUseCase) FindUnreadCount(ctx context.Context, recipientID primitive.ObjectID) (int64, error) {
	ctx, span := n.tracer.Start(ctx, "NotificationUseCase.FindUnreadCount")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"recipientID": recipientID.Hex(),
	}
	logger.Input(input)

	count, err := n.notificationRepo.CountUnread(ctx, recipientID)
	if err != nil {
		logger.Output("error counting unread notifications 1", err)
		return 0, err
	}

	return count, nil
}
