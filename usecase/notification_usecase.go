package usecase

import (
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type notificationUseCase struct {
	notificationRepo domain.NotificationRepository
}

func NewNotificationUseCase(notificationRepo domain.NotificationRepository) domain.NotificationUseCase {
	return &notificationUseCase{
		notificationRepo: notificationRepo,
	}
}

func (n *notificationUseCase) CreateNotification(recipientID, senderID, refID primitive.ObjectID, nType domain.NotificationType, refType, message string) (*domain.Notification, error) {
	logger := utils.NewLogger("NotificationUseCase.CreateNotification")
	input := map[string]interface{}{
		"recipientID": recipientID.Hex(),
		"senderID":    senderID.Hex(),
		"refID":       refID.Hex(),
		"type":        nType,
		"refType":     refType,
		"message":     message,
	}
	logger.LogInput(input)

	notification := &domain.Notification{
		RecipientID: recipientID,
		SenderID:    senderID,
		Type:        nType,
		RefID:       refID,
		RefType:     refType,
		Message:     message,
		IsRead:      false,
	}

	err := n.notificationRepo.Create(notification)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(notification, nil)
	return notification, nil
}

func (n *notificationUseCase) GetNotification(notificationID primitive.ObjectID) (*domain.Notification, error) {
	logger := utils.NewLogger("NotificationUseCase.GetNotification")
	input := map[string]interface{}{
		"notificationID": notificationID.Hex(),
	}
	logger.LogInput(input)

	notification, err := n.notificationRepo.FindByID(notificationID)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(notification, nil)
	return notification, nil
}

func (n *notificationUseCase) ListNotifications(recipientID primitive.ObjectID, limit, offset int) ([]domain.Notification, error) {
	logger := utils.NewLogger("NotificationUseCase.ListNotifications")
	input := map[string]interface{}{
		"recipientID": recipientID.Hex(),
		"limit":       limit,
		"offset":      offset,
	}
	logger.LogInput(input)

	notifications, err := n.notificationRepo.FindByRecipient(recipientID, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(notifications, nil)
	return notifications, nil
}

func (n *notificationUseCase) MarkAsRead(notificationID primitive.ObjectID) error {
	logger := utils.NewLogger("NotificationUseCase.MarkAsRead")
	input := map[string]interface{}{
		"notificationID": notificationID.Hex(),
	}
	logger.LogInput(input)

	err := n.notificationRepo.MarkAsRead(notificationID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(map[string]interface{}{"success": true}, nil)
	return nil
}

func (n *notificationUseCase) MarkAllAsRead(recipientID primitive.ObjectID) error {
	logger := utils.NewLogger("NotificationUseCase.MarkAllAsRead")
	input := map[string]interface{}{
		"recipientID": recipientID.Hex(),
	}
	logger.LogInput(input)

	err := n.notificationRepo.MarkAllAsRead(recipientID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(map[string]interface{}{"success": true}, nil)
	return nil
}

func (n *notificationUseCase) DeleteNotification(notificationID primitive.ObjectID) error {
	logger := utils.NewLogger("NotificationUseCase.DeleteNotification")
	input := map[string]interface{}{
		"notificationID": notificationID.Hex(),
	}
	logger.LogInput(input)

	err := n.notificationRepo.Delete(notificationID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(map[string]interface{}{"success": true}, nil)
	return nil
}

func (n *notificationUseCase) GetUnreadCount(recipientID primitive.ObjectID) (int64, error) {
	logger := utils.NewLogger("NotificationUseCase.GetUnreadCount")
	input := map[string]interface{}{
		"recipientID": recipientID.Hex(),
	}
	logger.LogInput(input)

	count, err := n.notificationRepo.CountUnread(recipientID)
	if err != nil {
		logger.LogOutput(nil, err)
		return 0, err
	}

	logger.LogOutput(map[string]interface{}{"count": count}, nil)
	return count, nil
}
