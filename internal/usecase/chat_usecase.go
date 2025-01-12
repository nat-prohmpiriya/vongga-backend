package usecase

import (
	"fmt"
	"time"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type chatUsecase struct {
	chatRepo            domain.ChatRepository
	userRepo            domain.UserRepository
	notificationUsecase domain.NotificationUseCase
}

func NewChatUsecase(chatRepo domain.ChatRepository, userRepo domain.UserRepository, notificationUsecase domain.NotificationUseCase) domain.ChatUsecase {
	return &chatUsecase{
		chatRepo:            chatRepo,
		userRepo:            userRepo,
		notificationUsecase: notificationUsecase,
	}
}

// Room operations
func (u *chatUsecase) CreatePrivateChat(userID1 string, userID2 string) (*domain.ChatRoom, error) {
	logger := utils.NewTraceLogger("ChatUsecase.CreatePrivateChat")
	logger.Input(map[string]interface{}{
		"userID1": userID1,
		"userID2": userID2,
	})

	// Check if room already exists
	rooms, err := u.chatRepo.FindRoomsByUser(userID1)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	// Find existing private chat room with both users
	for _, room := range rooms {
		if room.Type == "private" && len(room.Members) == 2 {
			members := room.Members
			if (members[0] == userID1 && members[1] == userID2) ||
				(members[0] == userID2 && members[1] == userID1) {
				// Find user details
				var users []domain.User
				for _, memberID := range room.Members {
					user, err := u.userRepo.FindUserByID(memberID)
					if err != nil {
						logger.Output(nil, err)
						continue
					}
					users = append(users, *user)
				}
				room.Users = users
				logger.Output(room, nil)
				return room, nil
			}
		}
	}

	// Find user details for new room
	user1, err := u.userRepo.FindUserByID(userID1)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	user2, err := u.userRepo.FindUserByID(userID2)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	// Create new room if not exists
	room := &domain.ChatRoom{
		BaseModel: domain.BaseModel{
			ID:        primitive.NewObjectID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			IsActive:  true,
			Version:   1,
		},
		Type:    "private",
		Members: []string{userID1, userID2},
		Users:   []domain.User{*user1, *user2},
	}

	// Create room
	err = u.chatRepo.CreateRoom(room)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(room, nil)
	return room, nil
}

func (u *chatUsecase) CreateGroupChat(name string, memberIDs []string) (*domain.ChatRoom, error) {
	logger := utils.NewTraceLogger("ChatUsecase.CreateGroupChat")
	logger.Input(map[string]interface{}{
		"name":      name,
		"memberIDs": memberIDs,
	})

	room := &domain.ChatRoom{
		BaseModel: domain.BaseModel{
			ID:        primitive.NewObjectID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			IsActive:  true,
			Version:   1,
		},
		Name:    name,
		Type:    "group",
		Members: memberIDs,
	}

	// Create room
	err := u.chatRepo.CreateRoom(room)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(room, nil)
	return room, nil
}

func (u *chatUsecase) FindUserChats(userID string) ([]*domain.ChatRoom, error) {
	logger := utils.NewTraceLogger("ChatUsecase.FindUserChats")
	logger.Input(map[string]interface{}{
		"userID": userID,
	})

	// Find rooms
	rooms, err := u.chatRepo.FindRoomsByUser(userID)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	// Find user details for each room
	for _, room := range rooms {
		var users []domain.User
		for _, memberID := range room.Members {
			user, err := u.userRepo.FindUserByID(memberID)
			if err != nil {
				logger.Output(nil, err)
				continue
			}
			users = append(users, *user)
		}
		room.Users = users
	}

	logger.Output(rooms, nil)
	return rooms, nil
}

func (u *chatUsecase) AddMemberToGroup(roomID string, userID string) error {
	logger := utils.NewTraceLogger("ChatUsecase.AddMemberToGroup")
	logger.Input(map[string]interface{}{
		"roomID": roomID,
		"userID": userID,
	})

	room, err := u.chatRepo.FindRoom(roomID)
	if err != nil {
		logger.Output(nil, err)
		return err
	}

	if room.Type != "group" {
		err := fmt.Errorf("cannot add member to non-group chat")
		logger.Output(nil, err)
		return err
	}

	if err := u.AddMemberToRoom(roomID, userID); err != nil {
		logger.Output(nil, err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) RemoveMemberFromGroup(roomID string, userID string) error {
	logger := utils.NewTraceLogger("ChatUsecase.RemoveMemberFromGroup")
	logger.Input(map[string]interface{}{
		"roomID": roomID,
		"userID": userID,
	})

	room, err := u.chatRepo.FindRoom(roomID)
	if err != nil {
		logger.Output(nil, err)
		return err
	}

	if room.Type != "group" {
		err := fmt.Errorf("cannot remove member from non-group chat")
		logger.Output(nil, err)
		return err
	}

	if err := u.RemoveMemberFromRoom(roomID, userID); err != nil {
		logger.Output(nil, err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) DeleteRoom(roomID string) error {
	logger := utils.NewTraceLogger("ChatUsecase.DeleteRoom")
	logger.Input(roomID)

	// Check if room exists
	room, err := u.chatRepo.FindRoom(roomID)
	if err != nil {
		logger.Output(nil, err)
		return err
	}
	if room == nil {
		err := fmt.Errorf("room not found")
		logger.Output(nil, err)
		return err
	}

	// Delete room and all related data
	err = u.chatRepo.DeleteRoom(roomID)
	if err != nil {
		logger.Output(nil, err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) FindRoom(roomID string) (*domain.ChatRoom, error) {
	logger := utils.NewTraceLogger("ChatUsecase.FindRoom")
	logger.Input(roomID)

	room, err := u.chatRepo.FindRoom(roomID)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(room, nil)
	return room, nil
}

func (u *chatUsecase) UpdateRoom(room *domain.ChatRoom) error {
	logger := utils.NewTraceLogger("ChatUsecase.UpdateRoom")
	logger.Input(room)

	// Find existing room
	existingRoom, err := u.FindRoom(room.ID.Hex())
	if err != nil {
		logger.Output(nil, err)
		return err
	}
	if existingRoom == nil {
		err := fmt.Errorf("room not found")
		logger.Output(nil, err)
		return err
	}

	// Update room
	err = u.chatRepo.UpdateRoom(room)
	if err != nil {
		logger.Output(nil, err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

// Message operations
func (u *chatUsecase) SendMessage(roomID string, senderID string, messageType string, content string) (*domain.ChatMessage, error) {
	logger := utils.NewTraceLogger("ChatUsecase.SendMessage")
	logger.Input(map[string]interface{}{
		"roomID":      roomID,
		"senderID":    senderID,
		"messageType": messageType,
		"content":     content,
	})

	// Validate roomID
	if !primitive.IsValidObjectID(roomID) {
		err := fmt.Errorf("invalid room ID format")
		logger.Output(nil, err)
		return nil, err
	}

	// Find room to verify it exists and sender is a member
	room, err := u.chatRepo.FindRoom(roomID)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}
	if room == nil {
		err := fmt.Errorf("room not found")
		logger.Output(nil, err)
		return nil, err
	}

	// Verify sender is a member of the room
	isMember := false
	for _, memberID := range room.Members {
		if memberID == senderID {
			isMember = true
			break
		}
	}
	if !isMember {
		err := fmt.Errorf("sender is not a member of this room")
		logger.Output(nil, err)
		return nil, err
	}

	message := &domain.ChatMessage{
		BaseModel: domain.BaseModel{
			ID:        primitive.NewObjectID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			IsActive:  true,
			Version:   1,
		},
		RoomID:   roomID,
		SenderID: senderID,
		Type:     messageType,
		Content:  content,
		ReadBy:   []string{senderID},
	}

	if err := u.chatRepo.CreateMessage(message); err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	// Create notifications for all other members
	for _, memberID := range room.Members {
		if memberID == senderID {
			continue
		}

		notification, err := u.CreateNotification(memberID, "new_message", roomID, message.ID.Hex())
		if err != nil {
			logger.Output(nil, err)
			return nil, err
		}

		notification.Message = "New message received"

		if err := u.chatRepo.CreateNotification(notification); err != nil {
			logger.Output(nil, err)
			return nil, err
		}
	}

	logger.Output(message, nil)
	return message, nil
}

func (u *chatUsecase) SendFileMessage(roomID string, senderID string, fileType string, fileSize int64, fileURL string) (*domain.ChatMessage, error) {
	logger := utils.NewTraceLogger("ChatUsecase.SendFileMessage")
	logger.Input(map[string]interface{}{
		"roomID":   roomID,
		"senderID": senderID,
		"fileType": fileType,
		"fileSize": fileSize,
		"fileURL":  fileURL,
	})

	if fileSize > 10*1024*1024 { // 10MB limit
		err := fmt.Errorf("file size exceeds 10MB limit")
		logger.Output(nil, err)
		return nil, err
	}

	if fileType != "jpg" && fileType != "png" && fileType != "gif" {
		err := fmt.Errorf("unsupported file type: %s", fileType)
		logger.Output(nil, err)
		return nil, err
	}

	message := &domain.ChatMessage{
		BaseModel: domain.BaseModel{
			ID:        primitive.NewObjectID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			IsActive:  true,
			Version:   1,
		},
		RoomID:   roomID,
		SenderID: senderID,
		Type:     "file",
		FileURL:  fileURL,
		FileType: fileType,
		FileSize: fileSize,
		ReadBy:   []string{senderID},
	}

	if err := u.chatRepo.CreateMessage(message); err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	// Create notifications for other members (similar to text message)
	room, err := u.chatRepo.FindRoom(roomID)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	for _, memberID := range room.Members {
		if memberID == senderID {
			continue
		}

		notification, err := u.CreateNotification(memberID, "new_message", roomID, message.ID.Hex())
		if err != nil {
			logger.Output(nil, err)
			return nil, err
		}

		notification.Message = "New file received"

		if err := u.chatRepo.CreateNotification(notification); err != nil {
			logger.Output(nil, err)
			return nil, err
		}
	}

	logger.Output(message, nil)
	return message, nil
}

func (u *chatUsecase) FindChatMessages(roomID string, limit int, offset int) ([]*domain.ChatMessage, error) {
	logger := utils.NewTraceLogger("ChatUsecase.FindChatMessages")
	logger.Input(map[string]interface{}{
		"roomID": roomID,
		"limit":  limit,
		"offset": offset,
	})

	messages, err := u.chatRepo.FindRoomMessages(roomID, int64(limit), int64(offset))
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(messages, nil)
	return messages, nil
}

func (u *chatUsecase) MarkMessageRead(messageID string, userID string) error {
	logger := utils.NewTraceLogger("ChatUsecase.MarkMessageRead")
	logger.Input(map[string]interface{}{
		"messageID": messageID,
		"userID":    userID,
	})

	if err := u.chatRepo.MarkMessageAsRead(messageID, userID); err != nil {
		logger.Output(nil, err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) DeleteMessage(messageID string) error {
	logger := utils.NewTraceLogger("ChatUsecase.DeleteMessage")
	logger.Input(messageID)

	// Check if message exists
	message, err := u.chatRepo.FindMessage(messageID)
	if err != nil {
		logger.Output(nil, err)
		return err
	}
	if message == nil {
		err := fmt.Errorf("message not found")
		logger.Output(nil, err)
		return err
	}

	// Delete message
	err = u.chatRepo.DeleteMessage(messageID)
	if err != nil {
		logger.Output(nil, err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) FindUnreadMessages(roomID string, userID string) ([]*domain.ChatMessage, error) {
	logger := utils.NewTraceLogger("ChatUsecase.FindUnreadMessages")
	logger.Input(map[string]interface{}{
		"roomID": roomID,
		"userID": userID,
	})

	// Find unread messages from the room
	messages, err := u.chatRepo.FindUnreadMessages(roomID, userID)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(messages, nil)
	return messages, nil
}

func (u *chatUsecase) FindMessage(messageID string) (*domain.ChatMessage, error) {
	logger := utils.NewTraceLogger("ChatUsecase.FindMessage")
	logger.Input(messageID)

	message, err := u.chatRepo.FindMessage(messageID)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(message, nil)
	return message, nil
}

func (u *chatUsecase) FindNotification(notificationID string) (*domain.ChatNotification, error) {
	logger := utils.NewTraceLogger("ChatUsecase.FindNotification")
	logger.Input(notificationID)

	notification, err := u.chatRepo.FindNotification(notificationID)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(notification, nil)
	return notification, nil
}

// User status operations
func (u *chatUsecase) UpdateUserOnlineStatus(userID string, isOnline bool) error {
	logger := utils.NewTraceLogger("ChatUsecase.UpdateUserOnlineStatus")
	logger.Input(map[string]interface{}{
		"userID":   userID,
		"isOnline": isOnline,
	})

	status := &domain.ChatUserStatus{
		BaseModel: domain.BaseModel{
			ID:        primitive.NewObjectID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			IsActive:  true,
			Version:   1,
		},
		UserID:   userID,
		IsOnline: isOnline,
		LastSeen: time.Now(),
	}

	if err := u.chatRepo.UpdateUserStatus(status); err != nil {
		logger.Output(nil, err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) FindUserOnlineStatus(userID string) (*domain.ChatUserStatus, error) {
	logger := utils.NewTraceLogger("ChatUsecase.FindUserOnlineStatus")
	logger.Input(userID)

	status, err := u.chatRepo.FindUserStatus(userID)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(status, nil)
	return status, nil
}

func (u *chatUsecase) FindOnlineUsers(userIDs []string) ([]*domain.ChatUserStatus, error) {
	logger := utils.NewTraceLogger("ChatUsecase.FindOnlineUsers")
	logger.Input(userIDs)

	statuses := make([]*domain.ChatUserStatus, 0)
	for _, userID := range userIDs {
		status, err := u.chatRepo.FindUserStatus(userID)
		if err != nil {
			logger.Output(nil, err)
			continue
		}
		if status != nil {
			statuses = append(statuses, status)
		}
	}

	logger.Output(statuses, nil)
	return statuses, nil
}

// Notification operations
func (u *chatUsecase) CreateNotification(userID string, notificationType string, roomID string, messageID string) (*domain.ChatNotification, error) {
	logger := utils.NewTraceLogger("ChatUsecase.CreateNotification")
	logger.Input(map[string]interface{}{
		"userID":           userID,
		"notificationType": notificationType,
		"roomID":           roomID,
		"messageID":        messageID,
	})

	// Create notification
	notification := &domain.ChatNotification{
		BaseModel: domain.BaseModel{
			ID:        primitive.NewObjectID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			IsActive:  true,
			Version:   1,
		},
		UserID:    userID,
		Type:      notificationType,
		RoomID:    roomID,
		MessageID: messageID,
	}

	// Create notification
	err := u.chatRepo.CreateNotification(notification)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(notification, nil)
	return notification, nil
}

func (u *chatUsecase) FindUserNotifications(userID string) ([]*domain.ChatNotification, error) {
	logger := utils.NewTraceLogger("ChatUsecase.FindUserNotifications")
	logger.Input(userID)

	notifications, err := u.chatRepo.FindUserNotifications(userID)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(notifications, nil)
	return notifications, nil
}

func (u *chatUsecase) MarkNotificationRead(notificationID string) error {
	logger := utils.NewTraceLogger("ChatUsecase.MarkNotificationRead")
	logger.Input(notificationID)

	if err := u.chatRepo.MarkNotificationAsRead(notificationID); err != nil {
		logger.Output(nil, err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) DeleteNotification(notificationID string) error {
	logger := utils.NewTraceLogger("ChatUsecase.DeleteNotification")
	logger.Input(notificationID)

	// Check if notification exists
	notification, err := u.chatRepo.FindNotification(notificationID)
	if err != nil {
		logger.Output(nil, err)
		return err
	}
	if notification == nil {
		err := fmt.Errorf("notification not found")
		logger.Output(nil, err)
		return err
	}

	// Delete notification
	err = u.chatRepo.DeleteNotification(notificationID)
	if err != nil {
		logger.Output(nil, err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) SendNotification(notification *domain.ChatNotification) error {
	logger := utils.NewTraceLogger("ChatUsecase.SendNotification")
	logger.Input(notification)

	// Create notification
	err := u.chatRepo.CreateNotification(notification)
	if err != nil {
		logger.Output(nil, err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) AddMemberToRoom(roomID string, userID string) error {
	logger := utils.NewTraceLogger("ChatUsecase.AddMemberToRoom")
	logger.Input(map[string]interface{}{
		"roomID": roomID,
		"userID": userID,
	})

	// Find room
	room, err := u.FindRoom(roomID)
	if err != nil {
		logger.Output(nil, err)
		return err
	}
	if room == nil {
		err := fmt.Errorf("room not found")
		logger.Output(nil, err)
		return err
	}

	// Check if user is already a member
	for _, memberID := range room.Members {
		if memberID == userID {
			err := fmt.Errorf("user is already a member")
			logger.Output(nil, err)
			return err
		}
	}

	// Add member
	room.Members = append(room.Members, userID)
	if err := u.chatRepo.UpdateRoom(room); err != nil {
		logger.Output(nil, err)
		return err
	}

	// Create notification for the new member
	notification, err := u.CreateNotification(userID, "group_invite", roomID, "")
	if err != nil {
		logger.Output(nil, err)
		return err
	}

	notification.Message = fmt.Sprintf("You have been added to group: %s", room.Name)

	if err := u.chatRepo.CreateNotification(notification); err != nil {
		logger.Output(nil, err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) RemoveMemberFromRoom(roomID string, userID string) error {
	logger := utils.NewTraceLogger("ChatUsecase.RemoveMemberFromRoom")
	logger.Input(map[string]interface{}{
		"roomID": roomID,
		"userID": userID,
	})

	// Find room
	room, err := u.FindRoom(roomID)
	if err != nil {
		logger.Output(nil, err)
		return err
	}
	if room == nil {
		err := fmt.Errorf("room not found")
		logger.Output(nil, err)
		return err
	}

	// Check if user is a member
	found := false
	newMembers := make([]string, 0)
	for _, memberID := range room.Members {
		if memberID == userID {
			found = true
			continue
		}
		newMembers = append(newMembers, memberID)
	}
	if !found {
		err := fmt.Errorf("user is not a member")
		logger.Output(nil, err)
		return err
	}

	// Remove member
	room.Members = newMembers
	if err := u.chatRepo.UpdateRoom(room); err != nil {
		logger.Output(nil, err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) FindUserRooms(userID string) ([]*domain.ChatRoom, error) {
	logger := utils.NewTraceLogger("ChatUsecase.FindUserRooms")
	logger.Input(userID)

	rooms, err := u.chatRepo.FindRoomsByUser(userID)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(rooms, nil)
	return rooms, nil
}

func (u *chatUsecase) FindRoomsByUserID(userID string) ([]*domain.ChatRoom, error) {
	return u.chatRepo.FindRoomsByUser(userID)
}
