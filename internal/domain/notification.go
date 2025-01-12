package domain

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeLike      NotificationType = "like"
	NotificationTypeComment   NotificationType = "comment"
	NotificationTypeFollow    NotificationType = "follow"
	NotificationTypeFriendReq NotificationType = "friend_request"
	NotificationTypeMention   NotificationType = "mention"
)

// Notification represents a notification entity
type Notification struct {
	BaseModel   `bson:",inline"`
	RecipientID primitive.ObjectID `bson:"recipientId" json:"recipientId"`
	SenderID    primitive.ObjectID `bson:"senderId" json:"senderId"`
	Type        NotificationType   `bson:"type" json:"type"`
	RefID       primitive.ObjectID `bson:"refId" json:"refId"`     // Reference ID (e.g., post ID, comment ID)
	RefType     string             `bson:"refType" json:"refType"` // Reference type (e.g., "post", "comment")
	Message     string             `bson:"message" json:"message"`
	IsRead      bool               `bson:"isRead" json:"isRead"`
}

// NotificationResponse represents a notification with sender information
type NotificationResponse struct {
	Notification `bson:",inline"`
	Sender       struct {
		UserID       string `json:"userId"`
		Username     string `json:"username"`
		DisplayName  string `json:"displayName"`
		PhotoProfile string `json:"photoProfile"`
		FirstName    string `json:"firstName"`
		LastName     string `json:"lastName"`
	} `json:"sender"`
}

// NotificationRepository interface
type NotificationRepository interface {
	Create(ctx context.Context, notification *Notification) error
	Update(ctx context.Context, notification *Notification) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*Notification, error)
	FindByRecipient(ctx context.Context, recipientID primitive.ObjectID, limit, offset int) ([]Notification, error)
	MarkAsRead(ctx context.Context, notificationID primitive.ObjectID) error
	MarkAllAsRead(ctx context.Context, recipientID primitive.ObjectID) error
	CountUnread(ctx context.Context, recipientID primitive.ObjectID) (int64, error)
}

// NotificationUseCase interface
type NotificationUseCase interface {
	CreateNotification(ctx context.Context, recipientID, senderID, refID primitive.ObjectID, nType NotificationType, refType, message string) (*Notification, error)
	FindNotification(ctx context.Context, notificationID primitive.ObjectID) (*NotificationResponse, error)
	FindManyNotifications(ctx context.Context, recipientID primitive.ObjectID, limit, offset int) ([]NotificationResponse, error)
	MarkAsRead(ctx context.Context, notificationID primitive.ObjectID) error
	MarkAllAsRead(ctx context.Context, recipientID primitive.ObjectID) error
	DeleteNotification(ctx context.Context, notificationID primitive.ObjectID) error
	FindUnreadCount(ctx context.Context, recipientID primitive.ObjectID) (int64, error)
}
