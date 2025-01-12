package repository

import (
	"context"
	"time"

	"vongga_api/internal/domain"
	"vongga_api/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/trace"
)

type chatRepository struct {
	db                *mongo.Database
	roomsColl         *mongo.Collection
	messagesColl      *mongo.Collection
	notificationsColl *mongo.Collection
	userStatusColl    *mongo.Collection
	tracer            trace.Tracer
}

func NewChatRepository(db *mongo.Database, tracer trace.Tracer) domain.ChatRepository {
	return &chatRepository{
		db:                db,
		roomsColl:         db.Collection("chatRooms"),
		messagesColl:      db.Collection("chatMessages"),
		notificationsColl: db.Collection("chatNotifications"),
		userStatusColl:    db.Collection("chatUserStatus"),
		tracer:            tracer,
	}
}

// Room operations
func (r *chatRepository) CreateRoom(ctx context.Context, room *domain.ChatRoom) error {
	ctx, span := r.tracer.Start(ctx, "ChatRepository.CreateRoom")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(room)

	room.CreatedAt = time.Now()
	room.UpdatedAt = time.Now()
	_, err := r.roomsColl.InsertOne(ctx, room)
	if err != nil {
		logger.Output("failed to create chat room 1", err)
		return err
	}

	logger.Output(room, nil)
	return nil
}

func (r *chatRepository) FindRoom(ctx context.Context, roomID string) (*domain.ChatRoom, error) {
	ctx, span := r.tracer.Start(ctx, "ChatRepository.FindRoom")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(roomID)

	// Convert string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		logger.Output("failed to convert room id 1", err)
		return nil, err
	}

	filter := bson.M{"_id": objectID}
	var room domain.ChatRoom
	err = r.roomsColl.FindOne(ctx, filter).Decode(&room)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Output("room not found 2", nil)
			return nil, nil
		}
		logger.Output("failed to find room 3", err)
		return nil, err
	}

	logger.Output(&room, nil)
	return &room, nil
}

func (r *chatRepository) FindRoomsByUser(ctx context.Context, userID string) ([]*domain.ChatRoom, error) {
	ctx, span := r.tracer.Start(ctx, "ChatRepository.FindRoomsByUser")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(userID)

	cursor, err := r.roomsColl.Find(ctx, bson.M{"members": userID})
	if err != nil {
		logger.Output("failed to find rooms 1", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var rooms []*domain.ChatRoom
	if err = cursor.All(ctx, &rooms); err != nil {
		logger.Output("failed to decode rooms 2", err)
		return nil, err
	}

	logger.Output(rooms, nil)
	return rooms, nil
}

func (r *chatRepository) AddMemberToRoom(ctx context.Context, roomID string, userID string) error {
	ctx, span := r.tracer.Start(ctx, "ChatRepository.AddMemberToRoom")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]string{"roomID": roomID, "userID": userID})

	_, err := r.roomsColl.UpdateOne(
		ctx,
		bson.M{"_id": roomID},
		bson.M{"$addToSet": bson.M{"members": userID}},
	)
	if err != nil {
		logger.Output("failed to add member to room 1", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (r *chatRepository) RemoveMemberFromRoom(ctx context.Context, roomID string, userID string) error {
	ctx, span := r.tracer.Start(ctx, "ChatRepository.RemoveMemberFromRoom")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]string{"roomID": roomID, "userID": userID})

	_, err := r.roomsColl.UpdateOne(
		ctx,
		bson.M{"_id": roomID},
		bson.M{"$pull": bson.M{"members": userID}},
	)
	if err != nil {
		logger.Output("failed to remove member from room 1", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (r *chatRepository) DeleteRoom(ctx context.Context, roomID string) error {
	ctx, span := r.tracer.Start(ctx, "ChatRepository.DeleteRoom")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(roomID)

	filter := bson.M{"_id": roomID}
	_, err := r.roomsColl.DeleteOne(ctx, filter)
	if err != nil {
		logger.Output("failed to delete room 1", err)
		return err
	}

	// Delete all messages in the room
	messageFilter := bson.M{"roomId": roomID}
	_, err = r.messagesColl.DeleteMany(ctx, messageFilter)
	if err != nil {
		logger.Output("failed to delete room messages 2", err)
		return err
	}

	// Delete all notifications related to the room
	notificationFilter := bson.M{"roomId": roomID}
	_, err = r.notificationsColl.DeleteMany(ctx, notificationFilter)
	if err != nil {
		logger.Output("failed to delete room notifications 3", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (r *chatRepository) UpdateRoom(ctx context.Context, room *domain.ChatRoom) error {
	ctx, span := r.tracer.Start(ctx, "ChatRepository.UpdateRoom")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(room)

	filter := bson.M{"_id": room.ID}
	update := bson.M{
		"$set": bson.M{
			"name":      room.Name,
			"type":      room.Type,
			"members":   room.Members,
			"updatedAt": time.Now(),
		},
	}

	_, err := r.roomsColl.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Output("failed to update room 1", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

// Message operations
func (r *chatRepository) CreateMessage(ctx context.Context, message *domain.ChatMessage) error {
	ctx, span := r.tracer.Start(ctx, "ChatRepository.CreateMessage")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(message)

	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()
	_, err := r.messagesColl.InsertOne(ctx, message)
	if err != nil {
		logger.Output("failed to create message 1", err)
		return err
	}

	logger.Output(message, nil)
	return nil
}

func (r *chatRepository) FindRoomMessages(ctx context.Context, roomID string, limit, offset int64) ([]*domain.ChatMessage, error) {
	ctx, span := r.tracer.Start(ctx, "ChatRepository.FindRoomMessages")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"roomID": roomID,
		"limit":  limit,
		"offset": offset,
	})

	opts := options.Find().SetLimit(limit).SetSkip(offset).SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := r.messagesColl.Find(ctx, bson.M{"roomId": roomID}, opts)
	if err != nil {
		logger.Output("failed to find room messages 1", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*domain.ChatMessage
	if err = cursor.All(ctx, &messages); err != nil {
		logger.Output("failed to decode messages 2", err)
		return nil, err
	}

	logger.Output(messages, nil)
	return messages, nil
}

func (r *chatRepository) MarkMessageAsRead(ctx context.Context, messageID string, userID string) error {
	ctx, span := r.tracer.Start(ctx, "ChatRepository.MarkMessageAsRead")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]string{
		"messageID": messageID,
		"userID":    userID,
	})

	update := bson.M{
		"$addToSet": bson.M{"readBy": userID},
		"$set":      bson.M{"updatedAt": time.Now()},
	}

	_, err := r.messagesColl.UpdateOne(ctx, bson.M{"_id": messageID}, update)
	if err != nil {
		logger.Output("failed to mark message as read 1", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (r *chatRepository) FindUnreadMessages(ctx context.Context, userID string, roomID string) ([]*domain.ChatMessage, error) {
	ctx, span := r.tracer.Start(ctx, "ChatRepository.FindUnreadMessages")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]string{
		"userID": userID,
		"roomID": roomID,
	})

	objectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	filter := bson.M{
		"roomId": objectID,
		"readBy": bson.M{
			"$nin": []string{userID},
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.messagesColl.Find(context.Background(), filter, opts)
	if err != nil {
		logger.Output("failed to find unread messages 1", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*domain.ChatMessage
	if err = cursor.All(ctx, &messages); err != nil {
		logger.Output("failed to decode messages 2", err)
		return nil, err
	}

	logger.Output(messages, nil)
	return messages, nil
}

func (r *chatRepository) DeleteMessage(ctx context.Context, messageID string) error {
	ctx, span := r.tracer.Start(ctx, "ChatRepository.DeleteMessage")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(messageID)

	filter := bson.M{"_id": messageID}
	_, err := r.messagesColl.DeleteOne(ctx, filter)
	if err != nil {
		logger.Output("failed to delete message 1", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (r *chatRepository) FindMessage(ctx context.Context, messageID string) (*domain.ChatMessage, error) {
	_, span := r.tracer.Start(ctx, "ChatRepository.FindMessage")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(messageID)

	// Convert string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		logger.Output("string to ObjectID 0", err)
		return nil, err
	}

	filter := bson.M{"_id": objectID}
	var message domain.ChatMessage
	err = r.messagesColl.FindOne(context.Background(), filter).Decode(&message)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Output("message not found 1", nil)
			return nil, nil
		}
		logger.Output("failed to find message 2", err)
		return nil, err
	}

	logger.Output(&message, nil)
	return &message, nil
}

// User status operations
func (r *chatRepository) UpdateUserStatus(ctx context.Context, status *domain.ChatUserStatus) error {
	ctx, span := r.tracer.Start(ctx, "ChatRepository.UpdateUserStatus")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(status)

	filter := bson.M{"userId": status.UserID}
	update := bson.M{"$set": status}
	opts := options.Update().SetUpsert(true)

	_, err := r.userStatusColl.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		logger.Output("failed to update user status 1", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (r *chatRepository) FindUserStatus(ctx context.Context, userID string) (*domain.ChatUserStatus, error) {
	ctx, span := r.tracer.Start(ctx, "ChatRepository.FindUserStatus")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(userID)

	filter := bson.M{"userId": userID}
	var status domain.ChatUserStatus
	err := r.userStatusColl.FindOne(ctx, filter).Decode(&status)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Output("user status not found 1", nil)
			return nil, nil
		}
		logger.Output("failed to find user status 2", err)
		return nil, err
	}

	logger.Output(&status, nil)
	return &status, nil
}

func (r *chatRepository) FindOnlineUsers(ctx context.Context, userIDs []string) ([]*domain.ChatUserStatus, error) {
	ctx, span := r.tracer.Start(ctx, "ChatRepository.FindOnlineUsers")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(userIDs)

	cursor, err := r.userStatusColl.Find(
		ctx,
		bson.M{
			"_id":       bson.M{"$in": userIDs},
			"is_online": true,
		},
	)
	if err != nil {
		logger.Output("failed to find online users 1", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var statuses []*domain.ChatUserStatus
	if err = cursor.All(ctx, &statuses); err != nil {
		logger.Output("failed to decode user statuses 2", err)
		return nil, err
	}

	logger.Output(statuses, nil)
	return statuses, nil
}

// Notification operations
func (r *chatRepository) CreateNotification(ctx context.Context, notification *domain.ChatNotification) error {
	ctx, span := r.tracer.Start(ctx, "ChatRepository.CreateNotification")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(notification)

	notification.CreatedAt = time.Now()
	_, err := r.notificationsColl.InsertOne(ctx, notification)
	if err != nil {
		logger.Output("failed to create notification 1", err)
		return err
	}

	logger.Output(notification, nil)
	return nil
}

func (r *chatRepository) FindUserNotifications(ctx context.Context, userID string) ([]*domain.ChatNotification, error) {
	ctx, span := r.tracer.Start(ctx, "ChatRepository.FindUserNotifications")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(userID)

	cursor, err := r.notificationsColl.Find(
		ctx,
		bson.M{"userId": userID},
	)
	if err != nil {
		logger.Output("failed to find user notifications 1", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []*domain.ChatNotification
	if err = cursor.All(ctx, &notifications); err != nil {
		logger.Output("failed to decode notifications 2", err)
		return nil, err
	}

	logger.Output(notifications, nil)
	return notifications, nil
}

func (r *chatRepository) MarkNotificationAsRead(ctx context.Context, notificationID string) error {
	ctx, span := r.tracer.Start(ctx, "ChatRepository.MarkNotificationAsRead")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(notificationID)

	_, err := r.notificationsColl.UpdateOne(
		ctx,
		bson.M{"_id": notificationID},
		bson.M{"$set": bson.M{"is_read": true}},
	)
	if err != nil {
		logger.Output("failed to mark notification as read 1", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (r *chatRepository) DeleteNotification(ctx context.Context, notificationID string) error {
	ctx, span := r.tracer.Start(ctx, "ChatRepository.DeleteNotification")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(notificationID)

	filter := bson.M{"_id": notificationID}
	_, err := r.notificationsColl.DeleteOne(ctx, filter)
	if err != nil {
		logger.Output("failed to delete notification 1", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (r *chatRepository) FindNotification(ctx context.Context, notificationID string) (*domain.ChatNotification, error) {
	_, span := r.tracer.Start(ctx, "ChatRepository.FindNotification")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(notificationID)

	// Convert string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	filter := bson.M{"_id": objectID}
	var notification domain.ChatNotification
	err = r.notificationsColl.FindOne(context.Background(), filter).Decode(&notification)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Output("notification not found 1", nil)
			return nil, nil
		}
		logger.Output("failed to find notification 2", err)
		return nil, err
	}

	logger.Output(&notification, nil)
	return &notification, nil
}

func (r *chatRepository) DeleteRoomNotifications(ctx context.Context, roomID string) error {
	_, span := r.tracer.Start(ctx, "ChatRepository.DeleteRoomNotifications")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(roomID)

	// Convert string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		logger.Output(nil, err)
		return err
	}

	filter := bson.M{"roomId": objectID}
	_, err = r.notificationsColl.DeleteMany(context.Background(), filter)
	if err != nil {
		logger.Output("failed to delete room notifications 1", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}
