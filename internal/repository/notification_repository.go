package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"vongga_api/internal/domain"
	"vongga_api/utils"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/trace"
)

type notificationRepository struct {
	collection *mongo.Collection
	rdb        *redis.Client
	tracer     trace.Tracer
}

func NewNotificationRepository(db *mongo.Database, rdb *redis.Client, tracer trace.Tracer) domain.NotificationRepository {
	return &notificationRepository{
		collection: db.Collection("notifications"),
		rdb:        rdb,
		tracer:     tracer,
	}
}

func (r *notificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	ctx, span := r.tracer.Start(ctx, "NotificationRepository.Create")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(notification)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, notification)
	if err != nil {
		logger.Output("failed to insert notification 1", err)
		return err
	}

	notification.ID = result.InsertedID.(primitive.ObjectID)

	// Invalidate recipient's notifications cache and unread count
	pattern := fmt.Sprintf("user_notifications:%s:*", notification.RecipientID)
	unreadKey := fmt.Sprintf("unread_count:%s", notification.RecipientID)

	keys, err := r.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.Output("failed to get cache keys 2", err)
		return err
	}
	if len(keys) > 0 {
		err = r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			logger.Output("failed to invalidate cache 3", err)
			return err
		}
	}

	// Delete unread count cache
	err = r.rdb.Del(ctx, unreadKey).Err()
	if err != nil {
		logger.Output("failed to delete unread count cache 4", err)
		return err
	}

	logger.Output(notification, nil)
	return nil
}

func (r *notificationRepository) Update(ctx context.Context, notification *domain.Notification) error {
	ctx, span := r.tracer.Start(ctx, "NotificationRepository.Update")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(notification)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	notification.UpdatedAt = time.Now()

	filter := bson.M{"_id": notification.ID}
	update := bson.M{"$set": notification}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Output("failed to update notification 1", err)
		return err
	}

	if result.MatchedCount == 0 {
		err := domain.ErrNotFound
		logger.Output("notification not found 2", err)
		return err
	}

	// Invalidate recipient's notifications cache and unread count
	pattern := fmt.Sprintf("user_notifications:%s:*", notification.RecipientID)
	unreadKey := fmt.Sprintf("unread_count:%s", notification.RecipientID)

	keys, err := r.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.Output("failed to get cache keys 3", err)
		return err
	}
	if len(keys) > 0 {
		err = r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			logger.Output("failed to invalidate cache 4", err)
			return err
		}
	}

	// Delete unread count cache
	err = r.rdb.Del(ctx, unreadKey).Err()
	if err != nil {
		logger.Output("failed to delete unread count cache 5", err)
		return err
	}

	logger.Output(notification, nil)
	return nil
}

func (r *notificationRepository) Delete(ctx context.Context, id string) error {
	ctx, span := r.tracer.Start(ctx, "NotificationRepository.Delete")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{"id": id})

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.Output("invalid notification ID format", err)
		return err
	}

	filter := bson.M{"_id": objectID}
	
	// Use FindOneAndDelete to get the document and delete it in one operation
	notification := &domain.Notification{}
	err = r.collection.FindOneAndDelete(ctx, filter).Decode(notification)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = domain.ErrNotFound
			logger.Output("notification not found 1", err)
			return err
		}
		logger.Output("failed to delete notification 2", err)
		return err
	}

	// Invalidate recipient's notifications cache and unread count
	pattern := fmt.Sprintf("user_notifications:%s:*", notification.RecipientID)
	unreadKey := fmt.Sprintf("unread_count:%s", notification.RecipientID)

	keys, err := r.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.Output("failed to get cache keys 3", err)
		return err
	}
	if len(keys) > 0 {
		err = r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			logger.Output("failed to invalidate cache 4", err)
			return err
		}
	}

	// Delete unread count cache
	err = r.rdb.Del(ctx, unreadKey).Err()
	if err != nil {
		logger.Output("failed to delete unread count cache 5", err)
		return err
	}

	logger.Output(map[string]interface{}{"deleted": true}, nil)
	return nil
}

func (r *notificationRepository) FindByID(ctx context.Context, id string) (*domain.Notification, error) {
	ctx, span := r.tracer.Start(ctx, "NotificationRepository.FindByID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{"id": id})

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.Output("invalid notification ID format", err)
		return nil, err
	}

	var notification domain.Notification
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&notification)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = domain.ErrNotFound
			logger.Output("notification not found 1", err)
			return nil, err
		}
		logger.Output("failed to find notification 2", err)
		return nil, err
	}

	logger.Output(&notification, nil)
	return &notification, nil
}

func (r *notificationRepository) FindByRecipient(ctx context.Context, recipientID string, limit, offset int) ([]domain.Notification, error) {
	ctx, span := r.tracer.Start(ctx, "NotificationRepository.FindByRecipient")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"recipientID": recipientID,
		"limit":       limit,
		"offset":      offset,
	})

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	opts := options.Find().
		SetSort(bson.M{"createdAt": -1}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, bson.M{"recipientId": recipientID}, opts)
	if err != nil {
		logger.Output("failed to find notifications 1", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []domain.Notification
	if err = cursor.All(ctx, &notifications); err != nil {
		logger.Output("failed to decode notifications 2", err)
		return nil, err
	}

	// Cache notifications
	notificationsKey := fmt.Sprintf("user_notifications:%s:%d:%d", recipientID, limit, offset)
	notificationsJSON, err := json.Marshal(notifications)
	if err != nil {
		logger.Output("failed to marshal notifications 3", err)
		return nil, err
	}
	err = r.rdb.Set(ctx, notificationsKey, notificationsJSON, time.Hour*24).Err()
	if err != nil {
		logger.Output("failed to cache notifications 4", err)
		return nil, err
	}

	logger.Output(notifications, nil)
	return notifications, nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, notificationID string) error {
	ctx, span := r.tracer.Start(ctx, "NotificationRepository.MarkAsRead")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{"notificationID": notificationID})

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		logger.Output("invalid notification ID format", err)
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$set": bson.M{
			"isRead":    true,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Output("failed to mark notification as read 1", err)
		return err
	}

	if result.MatchedCount == 0 {
		err := domain.ErrNotFound
		logger.Output("notification not found 2", err)
		return err
	}

	// Invalidate recipient's notifications cache and unread count
	notification := &domain.Notification{}
	err = r.collection.FindOne(ctx, filter).Decode(notification)
	if err != nil {
		logger.Output("failed to find notification 3", err)
		return err
	}
	pattern := fmt.Sprintf("user_notifications:%s:*", notification.RecipientID)
	unreadKey := fmt.Sprintf("unread_count:%s", notification.RecipientID)

	keys, err := r.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.Output("failed to get cache keys 4", err)
		return err
	}
	if len(keys) > 0 {
		err = r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			logger.Output("failed to invalidate cache 5", err)
			return err
		}
	}

	// Delete unread count cache
	err = r.rdb.Del(ctx, unreadKey).Err()
	if err != nil {
		logger.Output("failed to delete unread count cache 6", err)
		return err
	}

	logger.Output(map[string]interface{}{"updated": true}, nil)
	return nil
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, recipientID string) error {
	ctx, span := r.tracer.Start(ctx, "NotificationRepository.MarkAllAsRead")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{"recipientID": recipientID})

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
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
		logger.Output("failed to mark all notifications as read 1", err)
		return err
	}

	// Invalidate recipient's notifications cache and unread count
	pattern := fmt.Sprintf("user_notifications:%s:*", recipientID)
	unreadKey := fmt.Sprintf("unread_count:%s", recipientID)

	keys, err := r.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.Output("failed to get cache keys 2", err)
		return err
	}
	if len(keys) > 0 {
		err = r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			logger.Output("failed to invalidate cache 3", err)
			return err
		}
	}

	// Delete unread count cache
	err = r.rdb.Del(ctx, unreadKey).Err()
	if err != nil {
		logger.Output("failed to delete unread count cache 4", err)
		return err
	}

	logger.Output(map[string]interface{}{"modifiedCount": result.ModifiedCount}, nil)
	return nil
}

func (r *notificationRepository) CountUnread(ctx context.Context, recipientID string) (int64, error) {
	ctx, span := r.tracer.Start(ctx, "NotificationRepository.CountUnread")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{"recipientID": recipientID})

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	unreadKey := fmt.Sprintf("unread_count:%s", recipientID)
	unreadCount, err := r.rdb.Get(ctx, unreadKey).Int64()
	if err != nil && err != redis.Nil {
		logger.Output("failed to get unread count from cache 1", err)
		return 0, err
	}
	if err == redis.Nil {
		filter := bson.M{
			"recipientId": recipientID,
			"isRead":      false,
		}

		count, err := r.collection.CountDocuments(ctx, filter)
		if err != nil {
			logger.Output("failed to count unread notifications 2", err)
			return 0, err
		}

		// Cache unread count
		err = r.rdb.Set(ctx, unreadKey, count, time.Hour*24).Err()
		if err != nil {
			logger.Output("failed to cache unread count 3", err)
			return 0, err
		}

		unreadCount = count
	}

	logger.Output(map[string]interface{}{"count": unreadCount}, nil)
	return unreadCount, nil
}
