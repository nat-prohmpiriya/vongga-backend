package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type notificationRepository struct {
	collection *mongo.Collection
	rdb        *redis.Client
}

func NewNotificationRepository(db *mongo.Database, rdb *redis.Client) domain.NotificationRepository {
	return &notificationRepository{
		collection: db.Collection("notifications"),
		rdb:        rdb,
	}
}

func (r *notificationRepository) Create(notification *domain.Notification) error {
	logger := utils.NewLogger("NotificationRepository.Create")
	logger.LogInput(notification)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, notification)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	notification.ID = result.InsertedID.(primitive.ObjectID)

	// Invalidate recipient's notifications cache and unread count
	pattern := fmt.Sprintf("user_notifications:%s:*", notification.RecipientID.Hex())
	unreadKey := fmt.Sprintf("unread_count:%s", notification.RecipientID.Hex())

	keys, err := r.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}
	if len(keys) > 0 {
		err = r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			logger.LogOutput(nil, err)
			return err
		}
	}

	// Delete unread count cache
	err = r.rdb.Del(ctx, unreadKey).Err()
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(notification, nil)
	return nil
}

func (r *notificationRepository) Update(notification *domain.Notification) error {
	logger := utils.NewLogger("NotificationRepository.Update")
	logger.LogInput(notification)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	notification.UpdatedAt = time.Now()

	filter := bson.M{"_id": notification.ID}
	update := bson.M{"$set": notification}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	if result.MatchedCount == 0 {
		err := domain.ErrNotFound
		logger.LogOutput(nil, err)
		return err
	}

	// Invalidate recipient's notifications cache and unread count
	pattern := fmt.Sprintf("user_notifications:%s:*", notification.RecipientID.Hex())
	unreadKey := fmt.Sprintf("unread_count:%s", notification.RecipientID.Hex())

	keys, err := r.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}
	if len(keys) > 0 {
		err = r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			logger.LogOutput(nil, err)
			return err
		}
	}

	// Delete unread count cache
	err = r.rdb.Del(ctx, unreadKey).Err()
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(notification, nil)
	return nil
}

func (r *notificationRepository) Delete(id primitive.ObjectID) error {
	logger := utils.NewLogger("NotificationRepository.Delete")
	logger.LogInput(map[string]interface{}{"id": id.Hex()})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	if result.DeletedCount == 0 {
		err := domain.ErrNotFound
		logger.LogOutput(nil, err)
		return err
	}

	// Invalidate recipient's notifications cache and unread count
	notification := &domain.Notification{}
	err = r.collection.FindOne(ctx, filter).Decode(notification)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}
	pattern := fmt.Sprintf("user_notifications:%s:*", notification.RecipientID.Hex())
	unreadKey := fmt.Sprintf("unread_count:%s", notification.RecipientID.Hex())

	keys, err := r.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}
	if len(keys) > 0 {
		err = r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			logger.LogOutput(nil, err)
			return err
		}
	}

	// Delete unread count cache
	err = r.rdb.Del(ctx, unreadKey).Err()
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(map[string]interface{}{"deleted": true}, nil)
	return nil
}

func (r *notificationRepository) FindByID(id primitive.ObjectID) (*domain.Notification, error) {
	logger := utils.NewLogger("NotificationRepository.FindByID")
	logger.LogInput(map[string]interface{}{"id": id.Hex()})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var notification domain.Notification
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&notification)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = domain.ErrNotFound
		}
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(&notification, nil)
	return &notification, nil
}

func (r *notificationRepository) FindByRecipient(recipientID primitive.ObjectID, limit, offset int) ([]domain.Notification, error) {
	logger := utils.NewLogger("NotificationRepository.FindByRecipient")
	logger.LogInput(map[string]interface{}{
		"recipientID": recipientID.Hex(),
		"limit":       limit,
		"offset":      offset,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().
		SetSort(bson.M{"createdAt": -1}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, bson.M{"recipientId": recipientID}, opts)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []domain.Notification
	if err = cursor.All(ctx, &notifications); err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Cache notifications
	notificationsKey := fmt.Sprintf("user_notifications:%s:%d:%d", recipientID.Hex(), limit, offset)
	notificationsJSON, err := json.Marshal(notifications)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}
	err = r.rdb.Set(ctx, notificationsKey, notificationsJSON, time.Hour*24).Err()
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(notifications, nil)
	return notifications, nil
}

func (r *notificationRepository) MarkAsRead(notificationID primitive.ObjectID) error {
	logger := utils.NewLogger("NotificationRepository.MarkAsRead")
	logger.LogInput(map[string]interface{}{"notificationID": notificationID.Hex()})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": notificationID}
	update := bson.M{
		"$set": bson.M{
			"isRead":    true,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	if result.MatchedCount == 0 {
		err := domain.ErrNotFound
		logger.LogOutput(nil, err)
		return err
	}

	// Invalidate recipient's notifications cache and unread count
	notification := &domain.Notification{}
	err = r.collection.FindOne(ctx, filter).Decode(notification)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}
	pattern := fmt.Sprintf("user_notifications:%s:*", notification.RecipientID.Hex())
	unreadKey := fmt.Sprintf("unread_count:%s", notification.RecipientID.Hex())

	keys, err := r.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}
	if len(keys) > 0 {
		err = r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			logger.LogOutput(nil, err)
			return err
		}
	}

	// Delete unread count cache
	err = r.rdb.Del(ctx, unreadKey).Err()
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(map[string]interface{}{"updated": true}, nil)
	return nil
}

func (r *notificationRepository) MarkAllAsRead(recipientID primitive.ObjectID) error {
	logger := utils.NewLogger("NotificationRepository.MarkAllAsRead")
	logger.LogInput(map[string]interface{}{"recipientID": recipientID.Hex()})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"recipientId": recipientID,
		"isRead":      false,
	}
	update := bson.M{
		"$set": bson.M{
			"isRead":    true,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Invalidate recipient's notifications cache and unread count
	pattern := fmt.Sprintf("user_notifications:%s:*", recipientID.Hex())
	unreadKey := fmt.Sprintf("unread_count:%s", recipientID.Hex())

	keys, err := r.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}
	if len(keys) > 0 {
		err = r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			logger.LogOutput(nil, err)
			return err
		}
	}

	// Delete unread count cache
	err = r.rdb.Del(ctx, unreadKey).Err()
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(map[string]interface{}{"modifiedCount": result.ModifiedCount}, nil)
	return nil
}

func (r *notificationRepository) CountUnread(recipientID primitive.ObjectID) (int64, error) {
	logger := utils.NewLogger("NotificationRepository.CountUnread")
	logger.LogInput(map[string]interface{}{"recipientID": recipientID.Hex()})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	unreadKey := fmt.Sprintf("unread_count:%s", recipientID.Hex())
	unreadCount, err := r.rdb.Get(ctx, unreadKey).Int64()
	if err != nil && err != redis.Nil {
		logger.LogOutput(nil, err)
		return 0, err
	}
	if err == redis.Nil {
		filter := bson.M{
			"recipientId": recipientID,
			"isRead":      false,
		}

		count, err := r.collection.CountDocuments(ctx, filter)
		if err != nil {
			logger.LogOutput(nil, err)
			return 0, err
		}

		// Cache unread count
		err = r.rdb.Set(ctx, unreadKey, count, time.Hour*24).Err()
		if err != nil {
			logger.LogOutput(nil, err)
			return 0, err
		}

		unreadCount = count
	}

	logger.LogOutput(map[string]interface{}{"count": unreadCount}, nil)
	return unreadCount, nil
}
