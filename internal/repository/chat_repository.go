package repository

import (
	"context"
	"time"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type chatRepository struct {
	db                *mongo.Database
	roomsColl         *mongo.Collection
	messagesColl      *mongo.Collection
	notificationsColl *mongo.Collection
	userStatusColl    *mongo.Collection
}

func NewChatRepository(ctx context.Context, db *mongo.Database) domain.ChatRepository {
	return &chatRepository{
		db:                db,
		roomsColl:         db.Collection("chatRooms"),
		messagesColl:      db.Collection("chatMessages"),
		notificationsColl: db.Collection("chatNotifications"),
		userStatusColl:    db.Collection("chatUserStatus"),
	}
}

// Room operations
func (r *chatRepository) CreateRoom(ctx context.Context, room *domain.ChatRoom) error {
	logger := utils.NewLogger("ChatRepository.CreateRoom")
	logger.LogInput(room)

	room.CreatedAt = time.Now()
	room.UpdatedAt = time.Now()
	_, err := r.roomsColl.InsertOne(ctx, room)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(room, nil)
	return nil
}

func (r *chatRepository) FindRoom(ctx context.Context, roomID string) (*domain.ChatRoom, error) {
	logger := utils.NewLogger("ChatRepository.FindRoom")
	logger.LogInput(roomID)

	// Convert string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	filter := bson.M{"_id": objectID}
	var room domain.ChatRoom
	err = r.roomsColl.FindOne(context.Background(), filter).Decode(&room)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.LogOutput(nil, nil)
			return nil, nil
		}
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(&room, nil)
	return &room, nil
}

func (r *chatRepository) FindRoomsByUser(ctx context.Context, userID string) ([]*domain.ChatRoom, error) {
	logger := utils.NewLogger("ChatRepository.FindRoomsByUser")
	logger.LogInput(userID)

	cursor, err := r.roomsColl.Find(ctx, bson.M{"members": userID})
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var rooms []*domain.ChatRoom
	if err = cursor.All(context.Background(), &rooms); err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(rooms, nil)
	return rooms, nil
}

func (r *chatRepository) AddMemberToRoom(ctx context.Context, roomID string, userID string) error {
	logger := utils.NewLogger("ChatRepository.AddMemberToRoom")
	logger.LogInput(map[string]string{"roomID": roomID, "userID": userID})

	_, err := r.roomsColl.UpdateOne(
		ctx,
		bson.M{"_id": roomID},
		bson.M{"$addToSet": bson.M{"members": userID}},
	)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}

func (r *chatRepository) RemoveMemberFromRoom(ctx context.Context, roomID string, userID string) error {
	logger := utils.NewLogger("ChatRepository.RemoveMemberFromRoom")
	logger.LogInput(map[string]string{"roomID": roomID, "userID": userID})

	_, err := r.roomsColl.UpdateOne(
		ctx,
		bson.M{"_id": roomID},
		bson.M{"$pull": bson.M{"members": userID}},
	)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}

func (r *chatRepository) DeleteRoom(ctx context.Context, roomID string) error {
	logger := utils.NewLogger("ChatRepository.DeleteRoom")
	logger.LogInput(roomID)

	filter := bson.M{"_id": roomID}
	_, err := r.roomsColl.DeleteOne(ctx, filter)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Delete all messages in the room
	messageFilter := bson.M{"roomId": roomID}
	_, err = r.messagesColl.DeleteMany(context.Background(), messageFilter)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Delete all notifications related to the room
	notificationFilter := bson.M{"roomId": roomID}
	_, err = r.notificationsColl.DeleteMany(context.Background(), notificationFilter)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}

func (r *chatRepository) UpdateRoom(ctx context.Context, room *domain.ChatRoom) error {
	logger := utils.NewLogger("ChatRepository.UpdateRoom")
	logger.LogInput(room)

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
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}

// Message operations
func (r *chatRepository) CreateMessage(ctx context.Context, message *domain.ChatMessage) error {
	logger := utils.NewLogger("ChatRepository.CreateMessage")
	logger.LogInput(message)

	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()
	_, err := r.messagesColl.InsertOne(ctx, message)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(message, nil)
	return nil
}

func (r *chatRepository) FindRoomMessages(ctx context.Context, roomID string, limit, offset int64) ([]*domain.ChatMessage, error) {
	logger := utils.NewLogger("ChatRepository.FindRoomMessages")
	logger.LogInput(map[string]interface{}{
		"roomID": roomID,
		"limit":  limit,
		"offset": offset,
	})

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(offset).
		SetLimit(limit)

	cursor, err := r.messagesColl.Find(context.Background(), bson.M{"roomId": roomID}, opts)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var messages []*domain.ChatMessage
	if err = cursor.All(context.Background(), &messages); err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(messages, nil)
	return messages, nil
}

func (r *chatRepository) MarkMessageAsRead(ctx context.Context, messageID string, userID string) error {
	logger := utils.NewLogger("ChatRepository.MarkMessageAsRead")
	logger.LogInput(map[string]string{
		"messageID": messageID,
		"userID":    userID,
	})

	_, err := r.messagesColl.UpdateOne(
		context.Background(),
		bson.M{"_id": messageID},
		bson.M{"$addToSet": bson.M{"read_by": userID}},
	)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}

func (r *chatRepository) FindUnreadMessages(ctx context.Context, userID string, roomID string) ([]*domain.ChatMessage, error) {
	logger := utils.NewLogger("ChatRepository.FindUnreadMessages")
	logger.LogInput(map[string]string{
		"userID": userID,
		"roomID": roomID,
	})

	objectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		logger.LogOutput(nil, err)
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
		logger.LogOutput(nil, err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var messages []*domain.ChatMessage
	if err = cursor.All(context.Background(), &messages); err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(messages, nil)
	return messages, nil
}

func (r *chatRepository) DeleteMessage(ctx context.Context, messageID string) error {
	logger := utils.NewLogger("ChatRepository.DeleteMessage")
	logger.LogInput(messageID)

	filter := bson.M{"_id": messageID}
	_, err := r.messagesColl.DeleteOne(ctx, filter)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}

func (r *chatRepository) FindMessage(ctx context.Context, messageID string) (*domain.ChatMessage, error) {
	logger := utils.NewLogger("ChatRepository.FindMessage")
	logger.LogInput(messageID)

	// Convert string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	filter := bson.M{"_id": objectID}
	var message domain.ChatMessage
	err = r.messagesColl.FindOne(context.Background(), filter).Decode(&message)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.LogOutput(nil, nil)
			return nil, nil
		}
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(&message, nil)
	return &message, nil
}

// User status operations
func (r *chatRepository) UpdateUserStatus(ctx context.Context, status *domain.ChatUserStatus) error {
	logger := utils.NewLogger("ChatRepository.UpdateUserStatus")
	logger.LogInput(status)

	filter := bson.M{"userId": status.UserID}
	update := bson.M{"$set": status}
	opts := options.Update().SetUpsert(true)

	_, err := r.userStatusColl.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}

func (r *chatRepository) FindUserStatus(ctx context.Context, userID string) (*domain.ChatUserStatus, error) {
	logger := utils.NewLogger("ChatRepository.FindUserStatus")
	logger.LogInput(userID)

	filter := bson.M{"userId": userID}
	var status domain.ChatUserStatus
	err := r.userStatusColl.FindOne(ctx, filter).Decode(&status)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.LogOutput(nil, nil)
			return nil, nil
		}
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(&status, nil)
	return &status, nil
}

func (r *chatRepository) FindOnlineUsers(ctx context.Context, userIDs []string) ([]*domain.ChatUserStatus, error) {
	logger := utils.NewLogger("ChatRepository.FindOnlineUsers")
	logger.LogInput(userIDs)

	cursor, err := r.userStatusColl.Find(
		ctx,
		bson.M{
			"_id":       bson.M{"$in": userIDs},
			"is_online": true,
		},
	)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var statuses []*domain.ChatUserStatus
	if err = cursor.All(context.Background(), &statuses); err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(statuses, nil)
	return statuses, nil
}

// Notification operations
func (r *chatRepository) CreateNotification(ctx context.Context, notification *domain.ChatNotification) error {
	logger := utils.NewLogger("ChatRepository.CreateNotification")
	logger.LogInput(notification)

	notification.CreatedAt = time.Now()
	_, err := r.notificationsColl.InsertOne(ctx, notification)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(notification, nil)
	return nil
}

func (r *chatRepository) FindUserNotifications(ctx context.Context, userID string) ([]*domain.ChatNotification, error) {
	logger := utils.NewLogger("ChatRepository.FindUserNotifications")
	logger.LogInput(userID)

	cursor, err := r.notificationsColl.Find(
		ctx,
		bson.M{"userId": userID},
	)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var notifications []*domain.ChatNotification
	if err = cursor.All(context.Background(), &notifications); err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(notifications, nil)
	return notifications, nil
}

func (r *chatRepository) MarkNotificationAsRead(ctx context.Context, notificationID string) error {
	logger := utils.NewLogger("ChatRepository.MarkNotificationAsRead")
	logger.LogInput(notificationID)

	_, err := r.notificationsColl.UpdateOne(
		ctx,
		bson.M{"_id": notificationID},
		bson.M{"$set": bson.M{"is_read": true}},
	)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}

func (r *chatRepository) DeleteNotification(ctx context.Context, notificationID string) error {
	logger := utils.NewLogger("ChatRepository.DeleteNotification")
	logger.LogInput(notificationID)

	filter := bson.M{"_id": notificationID}
	_, err := r.notificationsColl.DeleteOne(ctx, filter)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}

func (r *chatRepository) FindNotification(ctx context.Context, notificationID string) (*domain.ChatNotification, error) {
	logger := utils.NewLogger("ChatRepository.FindNotification")
	logger.LogInput(notificationID)

	// Convert string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	filter := bson.M{"_id": objectID}
	var notification domain.ChatNotification
	err = r.notificationsColl.FindOne(context.Background(), filter).Decode(&notification)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.LogOutput(nil, nil)
			return nil, nil
		}
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(&notification, nil)
	return &notification, nil
}

func (r *chatRepository) CreateNotification(ctx context.Context, notification *domain.ChatNotification) error {
	logger := utils.NewLogger("ChatRepository.CreateNotification")
	logger.LogInput(notification)

	notification.UpdatedAt = time.Now()
	_, err := r.notificationsColl.UpdateOne(
		context.Background(),
		bson.M{"_id": notification.ID},
		bson.M{"$set": notification},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}

func (r *chatRepository) DeleteRoomNotifications(ctx context.Context, roomID string) error {
	logger := utils.NewLogger("ChatRepository.DeleteRoomNotifications")
	logger.LogInput(roomID)

	// Convert string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	filter := bson.M{"roomId": objectID}
	_, err = r.notificationsColl.DeleteMany(context.Background(), filter)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}
