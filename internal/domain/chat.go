package domain

import (
	"time"
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
	SaveRoom(room *ChatRoom) error
	GetRoom(roomID string) (*ChatRoom, error)
	GetRoomsByUser(userID string) ([]*ChatRoom, error)
	UpdateRoom(room *ChatRoom) error
	DeleteRoom(roomID string) error

	// Message operations
	SaveMessage(message *ChatMessage) error
	GetMessage(messageID string) (*ChatMessage, error)
	GetRoomMessages(roomID string, limit int64, offset int64) ([]*ChatMessage, error)
	DeleteMessage(messageID string) error
	MarkMessageAsRead(messageID string, userID string) error
	GetUnreadMessages(userID string, roomID string) ([]*ChatMessage, error)

	// Notification operations
	CreateNotification(notification *ChatNotification) error
	SaveNotification(notification *ChatNotification) error
	GetUserNotifications(userID string) ([]*ChatNotification, error)
	GetNotification(notificationID string) (*ChatNotification, error)
	DeleteNotification(notificationID string) error
	DeleteRoomNotifications(roomID string) error
	MarkNotificationAsRead(notificationID string) error

	// User status operations
	UpdateUserStatus(status *ChatUserStatus) error
	GetUserStatus(userID string) (*ChatUserStatus, error)
	GetOnlineUsers(userIDs []string) ([]*ChatUserStatus, error)
}

type ChatUsecase interface {
	// Room operations
	CreatePrivateChat(userID1, userID2 string) (*ChatRoom, error)
	CreateGroupChat(name string, memberIDs []string) (*ChatRoom, error)
	GetUserChats(userID string) ([]*ChatRoom, error)
	GetRoom(roomID string) (*ChatRoom, error)
	GetRoomsByUserID(userID string) ([]*ChatRoom, error)
	AddMemberToGroup(roomID, userID string) error
	RemoveMemberFromGroup(roomID, userID string) error
	UpdateRoom(room *ChatRoom) error
	DeleteRoom(roomID string) error

	// Message operations
	SendMessage(roomID, senderID, messageType, content string) (*ChatMessage, error)
	SendFileMessage(roomID, senderID string, fileType string, fileSize int64, fileURL string) (*ChatMessage, error)
	GetChatMessages(roomID string, limit, offset int) ([]*ChatMessage, error)
	MarkMessageRead(messageID, userID string) error
	GetUnreadMessages(userID string, roomID string) ([]*ChatMessage, error)
	DeleteMessage(messageID string) error

	// User status operations
	UpdateUserOnlineStatus(userID string, isOnline bool) error
	GetUserOnlineStatus(userID string) (*ChatUserStatus, error)
	GetOnlineUsers(userIDs []string) ([]*ChatUserStatus, error)

	// Notification operations
	SendNotification(notification *ChatNotification) error
	GetUserNotifications(userID string) ([]*ChatNotification, error)
	MarkNotificationRead(notificationID string) error
	DeleteNotification(notificationID string) error
}
