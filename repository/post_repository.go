package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type postRepository struct {
	db         *mongo.Database
	rdb        *redis.Client
	collection *mongo.Collection
}

func NewPostRepository(db *mongo.Database, rdb *redis.Client) domain.PostRepository {
	return &postRepository{
		db:         db,
		rdb:        rdb,
		collection: db.Collection("posts"),
	}
}

func (r *postRepository) Create(post *domain.Post) error {
	logger := utils.NewLogger("PostRepository.Create")
	logger.LogInput(post)

	_, err := r.collection.InsertOne(context.Background(), post)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Invalidate user's posts cache
	ctx := context.Background()
	pattern := fmt.Sprintf("user_posts:%s:*", post.UserID.Hex())
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

	logger.LogOutput("Post created successfully", nil)
	return nil
}

func (r *postRepository) Update(post *domain.Post) error {
	logger := utils.NewLogger("PostRepository.Update")
	logger.LogInput(post)

	filter := bson.M{"_id": post.ID}
	update := bson.M{"$set": post}
	_, err := r.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Invalidate post cache and user's posts cache
	ctx := context.Background()
	key := fmt.Sprintf("post:%s", post.ID.Hex())
	pattern := fmt.Sprintf("user_posts:%s:*", post.UserID.Hex())

	// Delete post cache
	err = r.rdb.Del(ctx, key).Err()
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Delete user's posts cache
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

	logger.LogOutput("Post updated successfully", nil)
	return nil
}

func (r *postRepository) Delete(id primitive.ObjectID) error {
	logger := utils.NewLogger("PostRepository.Delete")
	logger.LogInput(id)

	now := time.Now()
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"deletedAt": now}}

	// Get post first to get userID for cache invalidation
	var post domain.Post
	err := r.collection.FindOne(context.Background(), filter).Decode(&post)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	_, err = r.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Invalidate post cache and user's posts cache
	ctx := context.Background()
	key := fmt.Sprintf("post:%s", id.Hex())
	pattern := fmt.Sprintf("user_posts:%s:*", post.UserID.Hex())

	// Delete post cache
	err = r.rdb.Del(ctx, key).Err()
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Delete user's posts cache
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

	logger.LogOutput("Post soft deleted successfully", nil)
	return nil
}

func (r *postRepository) FindByID(id primitive.ObjectID) (*domain.Post, error) {
	logger := utils.NewLogger("PostRepository.FindByID")
	logger.LogInput(id)

	ctx := context.Background()
	key := fmt.Sprintf("post:%s", id.Hex())

	// Try to get from Redis first
	postJSON, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		// Found in Redis
		var post domain.Post
		err = json.Unmarshal([]byte(postJSON), &post)
		if err != nil {
			logger.LogOutput(nil, err)
			return nil, err
		}
		logger.LogOutput(&post, nil)
		return &post, nil
	} else if err != redis.Nil {
		// Redis error
		logger.LogOutput(nil, err)
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
			logger.LogOutput(nil, notFoundErr)
			return nil, notFoundErr
		}
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Cache in Redis for 1 hour
	postBytes, err := json.Marshal(&post)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	err = r.rdb.Set(ctx, key, string(postBytes), time.Hour).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.LogOutput(nil, err)
	}

	logger.LogOutput(&post, nil)
	return &post, nil
}

func (r *postRepository) FindByUserID(userID primitive.ObjectID, limit, offset int, hasMedia bool, mediaType string) ([]domain.Post, error) {
	logger := utils.NewLogger("PostRepository.FindByUserID")

	input := map[string]interface{}{
		"userID":    userID,
		"limit":     limit,
		"offset":    offset,
		"hasMedia":  hasMedia,
		"mediaType": mediaType,
	}
	logger.LogInput(input)

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

	cursor, err := r.collection.Find(context.Background(), filter, opts)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var posts []domain.Post
	if err := cursor.All(context.Background(), &posts); err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(posts, nil)
	return posts, nil
}
