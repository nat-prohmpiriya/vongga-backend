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

type postRepository struct {
	db         *mongo.Database
	rdb        *redis.Client
	collection *mongo.Collection
	tracer     trace.Tracer
}

func NewPostRepository(db *mongo.Database, rdb *redis.Client, tracer trace.Tracer) domain.PostRepository {
	return &postRepository{
		db:         db,
		rdb:        rdb,
		collection: db.Collection("posts"),
		tracer:     tracer,
	}
}

func (r *postRepository) Create(ctx context.Context, post *domain.Post) error {
	ctx, span := r.tracer.Start(ctx, "PostRepository.Create")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(post)

	_, err := r.collection.InsertOne(ctx, post)
	if err != nil {
		logger.Output("failed to insert post 1", err)
		return err
	}

	// Invalidate user's posts cache
	pattern := fmt.Sprintf("user_posts:%s:*", post.UserID.Hex())
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

	logger.Output("Post created successfully", nil)
	return nil
}

func (r *postRepository) Update(ctx context.Context, post *domain.Post) error {
	ctx, span := r.tracer.Start(ctx, "PostRepository.Update")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(post)

	filter := bson.M{"_id": post.ID}
	update := bson.M{"$set": post}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Output("failed to update post 1", err)
		return err
	}

	// Invalidate post cache and user's posts cache
	key := fmt.Sprintf("post:%s", post.ID.Hex())
	pattern := fmt.Sprintf("user_posts:%s:*", post.UserID.Hex())

	err = r.rdb.Del(ctx, key).Err()
	if err != nil {
		logger.Output("failed to delete post cache 2", err)
		return err
	}

	keys, err := r.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.Output("failed to get user posts cache keys 3", err)
		return err
	}
	if len(keys) > 0 {
		err = r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			logger.Output("failed to delete user posts cache 4", err)
			return err
		}
	}

	logger.Output("Post updated successfully", nil)
	return nil
}

func (r *postRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	ctx, span := r.tracer.Start(ctx, "PostRepository.Delete")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(id)

	now := time.Now()
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"deletedAt": now}}

	// Find post first to get userID for cache invalidation
	var post domain.Post
	err := r.collection.FindOne(ctx, filter).Decode(&post)
	if err != nil {
		logger.Output("failed to find post 1", err)
		return err
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Output("failed to soft delete post 2", err)
		return err
	}

	// Invalidate post cache and user's posts cache
	key := fmt.Sprintf("post:%s", id.Hex())
	pattern := fmt.Sprintf("user_posts:%s:*", post.UserID.Hex())

	err = r.rdb.Del(ctx, key).Err()
	if err != nil {
		logger.Output("failed to delete post cache 3", err)
		return err
	}

	keys, err := r.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.Output("failed to get user posts cache keys 4", err)
		return err
	}
	if len(keys) > 0 {
		err = r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			logger.Output("failed to delete user posts cache 5", err)
			return err
		}
	}

	logger.Output("Post soft deleted successfully", nil)
	return nil
}

func (r *postRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Post, error) {
	ctx, span := r.tracer.Start(ctx, "PostRepository.FindByID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(id)

	key := fmt.Sprintf("post:%s", id.Hex())

	// Try to get from Redis first
	postJSON, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		// Found in Redis
		var post domain.Post
		err = json.Unmarshal([]byte(postJSON), &post)
		if err != nil {
			logger.Output("failed to unmarshal cached post 1", err)
			return nil, err
		}
		logger.Output(&post, nil)
		return &post, nil
	} else if err != redis.Nil {
		// Redis error
		logger.Output("redis error 2", err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
	filter := bson.M{
		"_id":       id,
		"deletedAt": bson.M{"$exists": false},
	}

	var post domain.Post
	err = r.collection.FindOne(ctx, filter).Decode(&post)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			notFoundErr := domain.NewNotFoundError("post", id.Hex())
			logger.Output("post not found 3", notFoundErr)
			return nil, notFoundErr
		}
		logger.Output("failed to find post 4", err)
		return nil, err
	}

	// Cache in Redis for 1 hour
	postBytes, err := json.Marshal(&post)
	if err != nil {
		logger.Output("failed to marshal post 5", err)
		return nil, err
	}

	err = r.rdb.Set(ctx, key, string(postBytes), time.Hour).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.Output("failed to cache post 6", err)
	}

	logger.Output(&post, nil)
	return &post, nil
}

func (r *postRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID, limit, offset int, hasMedia bool, mediaType string) ([]domain.Post, error) {
	ctx, span := r.tracer.Start(ctx, "PostRepository.FindByUserID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"userID":    userID,
		"limit":     limit,
		"offset":    offset,
		"hasMedia":  hasMedia,
		"mediaType": mediaType,
	})

	filter := bson.M{
		"userId":   userID,
		"isActive": true,
		"deletedAt": bson.M{
			"$exists": false,
		},
	}

	// Handle media filtering
	if hasMedia {
		if mediaType != "" {
			// Filter for specific media type
			filter = bson.M{
				"$and": []bson.M{
					{"userId": userID, "isActive": true},
					{"$or": []bson.M{
						{"media": bson.M{"$elemMatch": bson.M{"type": mediaType}}},
						{"subPosts.media": bson.M{"$elemMatch": bson.M{"type": mediaType}}},
					}},
				},
			}
		} else {
			// Filter for any media
			filter = bson.M{
				"$and": []bson.M{
					{"userId": userID, "isActive": true},
					{"$or": []bson.M{
						{"media": bson.M{"$exists": true, "$ne": []interface{}{}}},
						{"subPosts.media": bson.M{"$exists": true, "$ne": []interface{}{}}},
					}},
				},
			}
		}
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	if offset > 0 {
		opts.SetSkip(int64(offset))
	}
	opts.SetSort(bson.M{"createdAt": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		logger.Output("failed to find posts 1", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []domain.Post
	if err := cursor.All(ctx, &posts); err != nil {
		logger.Output("failed to decode posts 2", err)
		return nil, err
	}

	logger.Output(posts, nil)
	return posts, nil
}
