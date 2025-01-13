// internal/dto/notification.go
package dto

type NotificationType string

const (
	NotificationTypeLike      NotificationType = "like"
	NotificationTypeComment   NotificationType = "comment"
	NotificationTypeFollow    NotificationType = "follow"
	NotificationTypeFriendReq NotificationType = "friend_request"
	NotificationTypeMention   NotificationType = "mention"
)

// FindManyNotificationsRequest represents the request parameters for finding multiple notifications
type FindManyNotificationsRequest struct {
	Limit  int `validate:"min=1,max=100" query:"limit"` // default 10
	Offset int `validate:"min=0" query:"offset"`        // default 0
}

// FindNotificationRequest represents the request parameters for finding a specific notification
type FindNotificationRequest struct {
	ID string `validate:"required" param:"id"`
}

// MarkAsReadRequest represents the request parameters for marking a notification as read
type MarkAsReadRequest struct {
	ID string `validate:"required" param:"id"`
}

// DeleteNotificationRequest represents the request parameters for deleting a notification
type DeleteNotificationRequest struct {
	ID string `validate:"required" param:"id"`
}

// CreateNotificationRequest represents the request parameters for creating a new notification
type CreateNotificationRequest struct {
	RecipientID string           `validate:"required" json:"recipientId"`
	Type        NotificationType `validate:"required" json:"type"`
	Message     string           `validate:"required,min=1,max=500" json:"message"`
	RefID       string           `validate:"required" json:"refId"`
	RefType     string           `validate:"required" json:"refType"`
}
