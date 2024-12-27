package repository

import (
	"context"
	"time"

	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
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

func NewChatRepository(db *mongo.Database) domain.ChatRepository {
	return &chatRepository{
		db:                db,
		roomsColl:         db.Collection("chatRooms"),
		messagesColl:      db.Collection("chatMessages"),
		notificationsColl: db.Collection("chatNotifications"),
		userStatusColl:    db.Collection("chatUserStatus"),
	}
}

// Room operations
func (r *chatRepository) SaveRoom(room *domain.ChatRoom) error {
	logger := utils.NewLogger("ChatRepository.SaveRoom")
	logger.LogInput(room)

	room.CreatedAt = time.Now()
	room.UpdatedAt = time.Now()
	_, err := r.roomsColl.InsertOne(context.Background(), room)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(room, nil)
	return nil
}

func (r *chatRepository) GetRoom(roomID string) (*domain.ChatRoom, error) {
	logger := utils.NewLogger("ChatRepository.GetRoom")
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

func (r *chatRepository) GetRoomsByUser(userID string) ([]*domain.ChatRoom, error) {
	logger := utils.NewLogger("ChatRepository.GetRoomsByUser")
	logger.LogInput(userID)

	cursor, err := r.roomsColl.Find(context.Background(), bson.M{"members": userID})
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

func (r *chatRepository) AddMemberToRoom(roomID string, userID string) error {
	logger := utils.NewLogger("ChatRepository.AddMemberToRoom")
	logger.LogInput(map[string]string{"roomID": roomID, "userID": userID})

	_, err := r.roomsColl.UpdateOne(
		context.Background(),
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

func (r *chatRepository) RemoveMemberFromRoom(roomID string, userID string) error {
	logger := utils.NewLogger("ChatRepository.RemoveMemberFromRoom")
	logger.LogInput(map[string]string{"roomID": roomID, "userID": userID})

	_, err := r.roomsColl.UpdateOne(
		context.Background(),
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

func (r *chatRepository) DeleteRoom(roomID string) error {
	logger := utils.NewLogger("ChatRepository.DeleteRoom")
	logger.LogInput(roomID)

	filter := bson.M{"_id": roomID}
	_, err := r.roomsColl.DeleteOne(context.Background(), filter)
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

func (r *chatRepository) UpdateRoom(room *domain.ChatRoom) error {
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

	_, err := r.roomsColl.UpdateOne(context.Background(), filter, update)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}

// Message operations
func (r *chatRepository) SaveMessage(message *domain.ChatMessage) error {
	logger := utils.NewLogger("ChatRepository.SaveMessage")
	logger.LogInput(message)

	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()
	_, err := r.messagesColl.InsertOne(context.Background(), message)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(message, nil)
	return nil
}

func (r *chatRepository) GetRoomMessages(roomID string, limit, offset int64) ([]*domain.ChatMessage, error) {
	logger := utils.NewLogger("ChatRepository.GetRoomMessages")
	logger.LogInput(map[string]interface{}{
		"roomID": roomID,
		"limit":  limit,
		"offset": offset,
	})

	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetSkip(offset).
		SetLimit(limit)

	cursor, err := r.messagesColl.Find(
		context.Background(),
		bson.M{"room_id": roomID},
		opts,
	)
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

func (r *chatRepository) MarkMessageAsRead(messageID string, userID string) error {
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

func (r *chatRepository) GetUnreadMessages(userID string, roomID string) ([]*domain.ChatMessage, error) {
	logger := utils.NewLogger("ChatRepository.GetUnreadMessages")
	logger.LogInput(map[string]string{
		"userID": userID,
		"roomID": roomID,
	})

	cursor, err := r.messagesColl.Find(
		context.Background(),
		bson.M{
			"room_id": roomID,
			"read_by": bson.M{"$nin": []string{userID}},
		},
	)
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

func (r *chatRepository) DeleteMessage(messageID string) error {
	logger := utils.NewLogger("ChatRepository.DeleteMessage")
	logger.LogInput(messageID)

	filter := bson.M{"_id": messageID}
	_, err := r.messagesColl.DeleteOne(context.Background(), filter)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}

func (r *chatRepository) GetMessage(messageID string) (*domain.ChatMessage, error) {
	logger := utils.NewLogger("ChatRepository.GetMessage")
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
func (r *chatRepository) UpdateUserStatus(status *domain.ChatUserStatus) error {
	logger := utils.NewLogger("ChatRepository.UpdateUserStatus")
	logger.LogInput(status)

	status.UpdatedAt = time.Now()
	opts := options.Update().SetUpsert(true)
	_, err := r.userStatusColl.UpdateOne(
		context.Background(),
		bson.M{"_id": status.UserID},
		bson.M{"$set": status},
		opts,
	)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(status, nil)
	return nil
}

func (r *chatRepository) GetUserStatus(userID string) (*domain.ChatUserStatus, error) {
	logger := utils.NewLogger("ChatRepository.GetUserStatus")
	logger.LogInput(userID)

	var status domain.ChatUserStatus
	err := r.userStatusColl.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&status)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(&status, nil)
	return &status, nil
}

func (r *chatRepository) GetOnlineUsers(userIDs []string) ([]*domain.ChatUserStatus, error) {
	logger := utils.NewLogger("ChatRepository.GetOnlineUsers")
	logger.LogInput(userIDs)

	cursor, err := r.userStatusColl.Find(
		context.Background(),
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
func (r *chatRepository) CreateNotification(notification *domain.ChatNotification) error {
	logger := utils.NewLogger("ChatRepository.CreateNotification")
	logger.LogInput(notification)

	notification.CreatedAt = time.Now()
	_, err := r.notificationsColl.InsertOne(context.Background(), notification)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(notification, nil)
	return nil
}

func (r *chatRepository) GetUserNotifications(userID string) ([]*domain.ChatNotification, error) {
	logger := utils.NewLogger("ChatRepository.GetUserNotifications")
	logger.LogInput(userID)

	cursor, err := r.notificationsColl.Find(
		context.Background(),
		bson.M{"user_id": userID},
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

func (r *chatRepository) MarkNotificationAsRead(notificationID string) error {
	logger := utils.NewLogger("ChatRepository.MarkNotificationAsRead")
	logger.LogInput(notificationID)

	_, err := r.notificationsColl.UpdateOne(
		context.Background(),
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

func (r *chatRepository) DeleteNotification(notificationID string) error {
	logger := utils.NewLogger("ChatRepository.DeleteNotification")
	logger.LogInput(notificationID)

	filter := bson.M{"_id": notificationID}
	_, err := r.notificationsColl.DeleteOne(context.Background(), filter)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}

func (r *chatRepository) GetNotification(notificationID string) (*domain.ChatNotification, error) {
	logger := utils.NewLogger("ChatRepository.GetNotification")
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

func (r *chatRepository) SaveNotification(notification *domain.ChatNotification) error {
	logger := utils.NewLogger("ChatRepository.SaveNotification")
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

func (r *chatRepository) DeleteRoomNotifications(roomID string) error {
	logger := utils.NewLogger("ChatRepository.DeleteRoomNotifications")
	logger.LogInput(roomID)

	// Convert string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	filter := bson.M{"room_id": objectID}
	_, err = r.notificationsColl.DeleteMany(context.Background(), filter)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}
