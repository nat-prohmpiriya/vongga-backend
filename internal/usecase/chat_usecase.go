package usecase

import (
	"context"
	"fmt"
	"time"

	"vongga_api/internal/domain"
	"vongga_api/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
)

type chatUsecase struct {
	chatRepo            domain.ChatRepository
	userRepo            domain.UserRepository
	notificationUsecase domain.NotificationUseCase
	tracer              trace.Tracer
}

func NewChatUsecase(
	chatRepo domain.ChatRepository,
	userRepo domain.UserRepository,
	notificationUsecase domain.NotificationUseCase,
	tracer trace.Tracer,
) domain.ChatUsecase {
	return &chatUsecase{
		chatRepo:            chatRepo,
		userRepo:            userRepo,
		notificationUsecase: notificationUsecase,
		tracer:              tracer,
	}
}

// Room operations
func (u *chatUsecase) CreatePrivateChat(ctx context.Context, userID1 string, userID2 string) (*domain.ChatRoom, error) {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.CreatePrivateChat")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"userID1": userID1,
		"userID2": userID2,
	})

	// Check if room already exists
	rooms, err := u.chatRepo.FindRoomsByUser(ctx, userID1)
	if err != nil {
		logger.Output("error finding rooms 1", err)
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
					user, err := u.userRepo.FindUserByID(ctx, memberID)
					if err != nil {
						logger.Output("error finding user details 2", err)
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
	user1, err := u.userRepo.FindUserByID(ctx, userID1)
	if err != nil {
		logger.Output("error finding user1 3", err)
		return nil, err
	}

	user2, err := u.userRepo.FindUserByID(ctx, userID2)
	if err != nil {
		logger.Output("error finding user2 4", err)
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
	err = u.chatRepo.CreateRoom(ctx, room)
	if err != nil {
		logger.Output("error creating room 5", err)
		return nil, err
	}

	logger.Output(room, nil)
	return room, nil
}

func (u *chatUsecase) CreateGroupChat(ctx context.Context, name string, memberIDs []string) (*domain.ChatRoom, error) {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.CreateGroupChat")
	defer span.End()
	logger := utils.NewTraceLogger(span)
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
	err := u.chatRepo.CreateRoom(ctx, room)
	if err != nil {
		logger.Output("error creating group chat 1", err)
		return nil, err
	}

	logger.Output(room, nil)
	return room, nil
}

func (u *chatUsecase) FindUserChats(ctx context.Context, userID string) ([]*domain.ChatRoom, error) {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.FindUserChats")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"userID": userID,
	})

	// Find rooms
	rooms, err := u.chatRepo.FindRoomsByUser(ctx, userID)
	if err != nil {
		logger.Output("error finding rooms 1", err)
		return nil, err
	}

	// Find user details for each room
	for _, room := range rooms {
		var users []domain.User
		for _, memberID := range room.Members {
			user, err := u.userRepo.FindUserByID(ctx, memberID)
			if err != nil {
				logger.Output("error finding user details 2", err)
				continue
			}
			users = append(users, *user)
		}
		room.Users = users
	}

	logger.Output(rooms, nil)
	return rooms, nil
}

func (u *chatUsecase) AddMemberToGroup(ctx context.Context, roomID string, userID string) error {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.AddMemberToGroup")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"roomID": roomID,
		"userID": userID,
	})

	room, err := u.chatRepo.FindRoom(ctx, roomID)
	if err != nil {
		logger.Output("error finding room 1", err)
		return err
	}

	if room.Type != "group" {
		err := fmt.Errorf("room is not group chat 2")
		logger.Output(err, nil)
		return err
	}

	if err := u.AddMemberToRoom(ctx, roomID, userID); err != nil {
		logger.Output("error adding member 3", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) RemoveMemberFromGroup(ctx context.Context, roomID string, userID string) error {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.RemoveMemberFromGroup")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"roomID": roomID,
		"userID": userID,
	})

	room, err := u.chatRepo.FindRoom(ctx, roomID)
	if err != nil {
		logger.Output("error finding room 1", err)
		return err
	}

	if room.Type != "group" {
		err := fmt.Errorf("room is not group chat 2")
		logger.Output(err, nil)
		return err
	}

	if err := u.RemoveMemberFromRoom(ctx, roomID, userID); err != nil {
		logger.Output("error removing member 3", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) DeleteRoom(ctx context.Context, roomID string) error {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.DeleteRoom")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(roomID)

	// Check if room exists
	room, err := u.chatRepo.FindRoom(ctx, roomID)
	if err != nil {
		logger.Output("error finding room 1", err)
		return err
	}
	if room == nil {
		err := fmt.Errorf("room not found")
		logger.Output(err, nil)
		return err
	}

	// Delete room and all related data
	err = u.chatRepo.DeleteRoom(ctx, roomID)
	if err != nil {
		logger.Output("error deleting room 2", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) FindRoom(ctx context.Context, roomID string) (*domain.ChatRoom, error) {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.FindRoom")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(roomID)

	room, err := u.chatRepo.FindRoom(ctx, roomID)
	if err != nil {
		logger.Output("error finding room 1", err)
		return nil, err
	}

	logger.Output(room, nil)
	return room, nil
}

func (u *chatUsecase) UpdateRoom(ctx context.Context, room *domain.ChatRoom) error {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.UpdateRoom")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(room)

	// Find existing room
	existingRoom, err := u.FindRoom(ctx, room.ID.Hex())
	if err != nil {
		logger.Output("error finding existing room 1", err)
		return err
	}
	if existingRoom == nil {
		err := fmt.Errorf("room not found")
		logger.Output(err, nil)
		return err
	}

	// Update room
	err = u.chatRepo.UpdateRoom(ctx, room)
	if err != nil {
		logger.Output("error updating room 2", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

// Message operations
func (u *chatUsecase) SendMessage(ctx context.Context, roomID string, senderID string, messageType string, content string) (*domain.ChatMessage, error) {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.SendMessage")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"roomID":      roomID,
		"senderID":    senderID,
		"messageType": messageType,
		"content":     content,
	})

	// Validate roomID
	if !primitive.IsValidObjectID(roomID) {
		err := fmt.Errorf("invalid room ID format")
		logger.Output(err, nil)
		return nil, err
	}

	// Find room to verify it exists and sender is a member
	room, err := u.chatRepo.FindRoom(ctx, roomID)
	if err != nil {
		logger.Output("error finding room 1", err)
		return nil, err
	}
	if room == nil {
		err := fmt.Errorf("room not found")
		logger.Output(err, nil)
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
		logger.Output(err, nil)
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

	if err := u.chatRepo.CreateMessage(ctx, message); err != nil {
		logger.Output("error creating message 2", err)
		return nil, err
	}

	// Create notifications for all other members
	for _, memberID := range room.Members {
		if memberID == senderID {
			continue
		}

		notification, err := u.CreateNotification(ctx, memberID, "new_message", roomID, message.ID.Hex())
		if err != nil {
			logger.Output("error creating notification 3", err)
			return nil, err
		}

		notification.Message = "New message received"

		if err := u.chatRepo.CreateNotification(ctx, notification); err != nil {
			logger.Output("error creating notification 4", err)
			return nil, err
		}
	}

	logger.Output(message, nil)
	return message, nil
}

func (u *chatUsecase) SendFileMessage(ctx context.Context, roomID string, senderID string, fileType string, fileSize int64, fileURL string) (*domain.ChatMessage, error) {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.SendFileMessage")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"roomID":   roomID,
		"senderID": senderID,
		"fileType": fileType,
		"fileSize": fileSize,
		"fileURL":  fileURL,
	})

	if fileSize > 10*1024*1024 { // 10MB limit
		err := fmt.Errorf("file size exceeds 10MB limit")
		logger.Output(err, nil)
		return nil, err
	}

	if fileType != "jpg" && fileType != "png" && fileType != "gif" {
		err := fmt.Errorf("unsupported file type: %s", fileType)
		logger.Output(err, nil)
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

	if err := u.chatRepo.CreateMessage(ctx, message); err != nil {
		logger.Output("error creating message 2", err)
		return nil, err
	}

	// Create notifications for other members (similar to text message)
	room, err := u.chatRepo.FindRoom(ctx, roomID)
	if err != nil {
		logger.Output("error finding room 3", err)
		return nil, err
	}

	for _, memberID := range room.Members {
		if memberID == senderID {
			continue
		}

		notification, err := u.CreateNotification(ctx, memberID, "new_message", roomID, message.ID.Hex())
		if err != nil {
			logger.Output("error creating notification 4", err)
			return nil, err
		}

		notification.Message = "New file received"

		if err := u.chatRepo.CreateNotification(ctx, notification); err != nil {
			logger.Output("error creating notification 5", err)
			return nil, err
		}
	}

	logger.Output(message, nil)
	return message, nil
}

func (u *chatUsecase) FindChatMessages(ctx context.Context, roomID string, limit int, offset int) ([]*domain.ChatMessage, error) {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.FindChatMessages")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"roomID": roomID,
		"limit":  limit,
		"offset": offset,
	})

	messages, err := u.chatRepo.FindRoomMessages(ctx, roomID, int64(limit), int64(offset))
	if err != nil {
		logger.Output("error finding messages 1", err)
		return nil, err
	}

	logger.Output(messages, nil)
	return messages, nil
}

func (u *chatUsecase) MarkMessageRead(ctx context.Context, messageID string, userID string) error {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.MarkMessageRead")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"messageID": messageID,
		"userID":    userID,
	})

	if err := u.chatRepo.MarkMessageAsRead(ctx, messageID, userID); err != nil {
		logger.Output("error marking message as read 1", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) DeleteMessage(ctx context.Context, messageID string) error {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.DeleteMessage")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(messageID)

	// Check if message exists
	message, err := u.chatRepo.FindMessage(ctx, messageID)
	if err != nil {
		logger.Output("error finding message 1", err)
		return err
	}
	if message == nil {
		err := fmt.Errorf("message not found")
		logger.Output(err, nil)
		return err
	}

	// Delete message
	err = u.chatRepo.DeleteMessage(ctx, messageID)
	if err != nil {
		logger.Output("error deleting message 2", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) FindUnreadMessages(ctx context.Context, roomID string, userID string) ([]*domain.ChatMessage, error) {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.FindUnreadMessages")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"roomID": roomID,
		"userID": userID,
	})

	// Find unread messages from the room
	messages, err := u.chatRepo.FindUnreadMessages(ctx, roomID, userID)
	if err != nil {
		logger.Output("error finding unread messages 1", err)
		return nil, err
	}

	logger.Output(messages, nil)
	return messages, nil
}

func (u *chatUsecase) FindMessage(ctx context.Context, messageID string) (*domain.ChatMessage, error) {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.FindMessage")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(messageID)

	message, err := u.chatRepo.FindMessage(ctx, messageID)
	if err != nil {
		logger.Output("error finding message 1", err)
		return nil, err
	}

	logger.Output(message, nil)
	return message, nil
}

func (u *chatUsecase) FindNotification(ctx context.Context, notificationID string) (*domain.ChatNotification, error) {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.FindNotification")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(notificationID)

	notification, err := u.chatRepo.FindNotification(ctx, notificationID)
	if err != nil {
		logger.Output("error finding notification 1", err)
		return nil, err
	}

	logger.Output(notification, nil)
	return notification, nil
}

// User status operations
func (u *chatUsecase) UpdateUserOnlineStatus(ctx context.Context, userID string, isOnline bool) error {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.UpdateUserOnlineStatus")
	defer span.End()
	logger := utils.NewTraceLogger(span)
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

	if err := u.chatRepo.UpdateUserStatus(ctx, status); err != nil {
		logger.Output("error updating user status 1", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) FindUserOnlineStatus(ctx context.Context, userID string) (*domain.ChatUserStatus, error) {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.FindUserOnlineStatus")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(userID)

	status, err := u.chatRepo.FindUserStatus(ctx, userID)
	if err != nil {
		logger.Output("error finding user status 1", err)
		return nil, err
	}

	logger.Output(status, nil)
	return status, nil
}

func (u *chatUsecase) FindOnlineUsers(ctx context.Context, userIDs []string) ([]*domain.ChatUserStatus, error) {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.FindOnlineUsers")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(userIDs)

	statuses := make([]*domain.ChatUserStatus, 0)
	for _, userID := range userIDs {
		status, err := u.chatRepo.FindUserStatus(ctx, userID)
		if err != nil {
			logger.Output("error finding user status 1", err)
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
func (u *chatUsecase) CreateNotification(ctx context.Context, userID string, notificationType string, roomID string, messageID string) (*domain.ChatNotification, error) {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.CreateNotification")
	defer span.End()
	logger := utils.NewTraceLogger(span)
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
	err := u.chatRepo.CreateNotification(ctx, notification)
	if err != nil {
		logger.Output("error creating notification 1", err)
		return nil, err
	}

	logger.Output(notification, nil)
	return notification, nil
}

func (u *chatUsecase) FindUserNotifications(ctx context.Context, userID string) ([]*domain.ChatNotification, error) {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.FindUserNotifications")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(userID)

	notifications, err := u.chatRepo.FindUserNotifications(ctx, userID)
	if err != nil {
		logger.Output("error finding notifications 1", err)
		return nil, err
	}

	logger.Output(notifications, nil)
	return notifications, nil
}

func (u *chatUsecase) MarkNotificationRead(ctx context.Context, notificationID string) error {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.MarkNotificationRead")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(notificationID)

	if err := u.chatRepo.MarkNotificationAsRead(ctx, notificationID); err != nil {
		logger.Output("error marking notification as read 1", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) DeleteNotification(ctx context.Context, notificationID string) error {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.DeleteNotification")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(notificationID)

	// Check if notification exists
	notification, err := u.chatRepo.FindNotification(ctx, notificationID)
	if err != nil {
		logger.Output("error finding notification 1", err)
		return err
	}
	if notification == nil {
		err := fmt.Errorf("notification not found")
		logger.Output(err, nil)
		return err
	}

	// Delete notification
	err = u.chatRepo.DeleteNotification(ctx, notificationID)
	if err != nil {
		logger.Output("error deleting notification 2", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) SendNotification(ctx context.Context, notification *domain.ChatNotification) error {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.SendNotification")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(notification)

	// Create notification
	err := u.chatRepo.CreateNotification(ctx, notification)
	if err != nil {
		logger.Output("error creating notification 1", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) AddMemberToRoom(ctx context.Context, roomID string, userID string) error {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.AddMemberToRoom")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"roomID": roomID,
		"userID": userID,
	})

	// Find room
	room, err := u.FindRoom(ctx, roomID)
	if err != nil {
		logger.Output("error finding room 1", err)
		return err
	}
	if room == nil {
		err := fmt.Errorf("room not found")
		logger.Output(err, nil)
		return err
	}

	// Check if user is already a member
	for _, memberID := range room.Members {
		if memberID == userID {
			err := fmt.Errorf("user is already a member")
			logger.Output(err, nil)
			return err
		}
	}

	// Add member
	room.Members = append(room.Members, userID)
	if err := u.chatRepo.UpdateRoom(ctx, room); err != nil {
		logger.Output("error updating room 2", err)
		return err
	}

	// Create notification for the new member
	notification, err := u.CreateNotification(ctx, userID, "group_invite", roomID, "")
	if err != nil {
		logger.Output("error creating notification 3", err)
		return err
	}

	notification.Message = fmt.Sprintf("You have been added to group: %s", room.Name)

	if err := u.chatRepo.CreateNotification(ctx, notification); err != nil {
		logger.Output("error creating notification 4", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) RemoveMemberFromRoom(ctx context.Context, roomID string, userID string) error {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.RemoveMemberFromRoom")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"roomID": roomID,
		"userID": userID,
	})

	// Find room
	room, err := u.FindRoom(ctx, roomID)
	if err != nil {
		logger.Output("error finding room 1", err)
		return err
	}
	if room == nil {
		err := fmt.Errorf("room not found")
		logger.Output(err, nil)
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
		logger.Output(err, nil)
		return err
	}

	// Remove member
	room.Members = newMembers
	if err := u.chatRepo.UpdateRoom(ctx, room); err != nil {
		logger.Output("error updating room 2", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *chatUsecase) FindUserRooms(ctx context.Context, userID string) ([]*domain.ChatRoom, error) {
	ctx, span := u.tracer.Start(ctx, "ChatUsecase.FindUserRooms")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(userID)

	rooms, err := u.chatRepo.FindRoomsByUser(ctx, userID)
	if err != nil {
		logger.Output("error finding rooms 1", err)
		return nil, err
	}

	logger.Output(rooms, nil)
	return rooms, nil
}

func (u *chatUsecase) FindRoomsByUserID(ctx context.Context, userID string) ([]*domain.ChatRoom, error) {
	return u.chatRepo.FindRoomsByUser(ctx, userID)
}
