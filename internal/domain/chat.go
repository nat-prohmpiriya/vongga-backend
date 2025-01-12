package domain

import (
	"time"

	"context"
)

type ChatRoom struct {
	BaseModel `bson:",inline"`
	Name      string   `bson:"name" json:"name"`
	Type      string   `bson:"type" json:"type"` // "private" or "group"
	Members   []string `bson:"members" json:"members"`
	Users     []User   `bson:"users,omitempty" json:"users,omitempty"`
}

type ChatMessage struct {
	BaseModel `bson:",inline"`
	RoomID    string   `bson:"roomId" json:"roomId"`
	SenderID  string   `bson:"senderId" json:"senderId"`
	Type      string   `bson:"type" json:"type"` // "text" or "file"
	Content   string   `bson:"content" json:"content"`
	FileURL   string   `bson:"fileUrl,omitempty" json:"fileUrl,omitempty"`
	FileType  string   `bson:"fileType,omitempty" json:"fileType,omitempty"`
	FileSize  int64    `bson:"fileSize,omitempty" json:"fileSize,omitempty"`
	ReadBy    []string `bson:"readBy" json:"readBy"`
}

type ChatUserStatus struct {
	BaseModel `bson:",inline"`
	UserID    string    `bson:"userId" json:"userId"`
	IsOnline  bool      `bson:"isOnline" json:"isOnline"`
	LastSeen  time.Time `bson:"lastSeen" json:"lastSeen"`
}

type ChatNotification struct {
	BaseModel `bson:",inline"`
	UserID    string `bson:"userId" json:"userId"`
	Type      string `bson:"type" json:"type"` // "new_message" or "group_invite"
	RoomID    string `bson:"roomId" json:"roomId"`
	MessageID string `bson:"messageId,omitempty" json:"messageId,omitempty"`
	Message   string `bson:"message" json:"message"`
	IsRead    bool   `bson:"isRead" json:"isRead"`
}

type ChatRepository interface {
	// Room operations
	CreateRoom(ctx context.Context, croom *ChatRoom) error
	FindRoom(ctx context.Context, roomID string) (*ChatRoom, error)
	FindRoomsByUser(ctx context.Context, userID string) ([]*ChatRoom, error)
	UpdateRoom(ctx context.Context, room *ChatRoom) error
	DeleteRoom(ctx context.Context, roomID string) error

	// Message operations
	CreateMessage(ctx context.Context, message *ChatMessage) error
	FindMessage(ctx context.Context, messageID string) (*ChatMessage, error)
	FindRoomMessages(ctx context.Context, roomID string, limit int64, offset int64) ([]*ChatMessage, error)
	DeleteMessage(ctx context.Context, messageID string) error
	MarkMessageAsRead(ctx context.Context, messageID string, userID string) error
	FindUnreadMessages(ctx context.Context, userID string, roomID string) ([]*ChatMessage, error)

	// Notification operations
	CreateNotification(ctx context.Context, notification *ChatNotification) error
	FindUserNotifications(ctx context.Context, userID string) ([]*ChatNotification, error)
	FindNotification(ctx context.Context, notificationID string) (*ChatNotification, error)
	DeleteNotification(ctx context.Context, notificationID string) error
	DeleteRoomNotifications(ctx context.Context, roomID string) error
	MarkNotificationAsRead(ctx context.Context, notificationID string) error

	// User status operations
	UpdateUserStatus(ctx context.Context, status *ChatUserStatus) error
	FindUserStatus(ctx context.Context, userID string) (*ChatUserStatus, error)
	FindOnlineUsers(ctx context.Context, userIDs []string) ([]*ChatUserStatus, error)
}

type ChatUsecase interface {
	// Room operations
	CreatePrivateChat(ctx context.Context, userID1, userID2 string) (*ChatRoom, error)
	CreateGroupChat(ctx context.Context, name string, memberIDs []string) (*ChatRoom, error)
	FindUserChats(ctx context.Context, userID string) ([]*ChatRoom, error)
	FindRoom(ctx context.Context, roomID string) (*ChatRoom, error)
	FindRoomsByUserID(ctx context.Context, userID string) ([]*ChatRoom, error)
	AddMemberToGroup(ctx context.Context, roomID, userID string) error
	RemoveMemberFromGroup(ctx context.Context, roomID, userID string) error
	UpdateRoom(ctx context.Context, room *ChatRoom) error
	DeleteRoom(ctx context.Context, roomID string) error

	// Message operations
	SendMessage(ctx context.Context, roomID, senderID, messageType, content string) (*ChatMessage, error)
	SendFileMessage(ctx context.Context, roomID, senderID string, fileType string, fileSize int64, fileURL string) (*ChatMessage, error)
	FindChatMessages(ctx context.Context, roomID string, limit, offset int) ([]*ChatMessage, error)
	MarkMessageRead(ctx context.Context, messageID, userID string) error
	FindUnreadMessages(ctx context.Context, userID string, roomID string) ([]*ChatMessage, error)
	DeleteMessage(ctx context.Context, messageID string) error

	// User status operations
	UpdateUserOnlineStatus(ctx context.Context, userID string, isOnline bool) error
	FindUserOnlineStatus(ctx context.Context, userID string) (*ChatUserStatus, error)
	FindOnlineUsers(ctx context.Context, userIDs []string) ([]*ChatUserStatus, error)

	// Notification operations
	CreateNotification(ctx context.Context, userID string, notificationType string, roomID string, messageID string) (*ChatNotification, error)
	SendNotification(ctx context.Context, notification *ChatNotification) error
	FindUserNotifications(ctx context.Context, userID string) ([]*ChatNotification, error)
	MarkNotificationRead(ctx context.Context, notificationID string) error
	DeleteNotification(ctx context.Context, notificationID string) error
}
