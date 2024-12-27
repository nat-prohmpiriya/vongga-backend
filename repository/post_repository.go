package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
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

func (r *postRepository) FindByUserID(userID primitive.ObjectID, limit, offset int) ([]domain.Post, error) {
	logger := utils.NewLogger("PostRepository.FindByUserID")
	logger.LogInput(map[string]interface{}{
		"userID": userID,
		"limit":  limit,
		"offset": offset,
	})

	ctx := context.Background()
	key := fmt.Sprintf("user_posts:%s:%d:%d", userID.Hex(), limit, offset)

	// Try to get from Redis first
	postsJSON, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		// Found in Redis
		var posts []domain.Post
		err = json.Unmarshal([]byte(postsJSON), &posts)
		if err != nil {
			logger.LogOutput(nil, err)
			return nil, err
		}
		logger.LogOutput(posts, nil)
		return posts, nil
	} else if err != redis.Nil {
		// Redis error
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	filter := bson.M{
		"userId":    userID,
		"deletedAt": bson.M{"$exists": false},
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []domain.Post
	if err = cursor.All(ctx, &posts); err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Cache in Redis for 15 minutes
	postsBytes, err := json.Marshal(posts)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	err = r.rdb.Set(ctx, key, string(postsBytes), 15*time.Minute).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.LogOutput(nil, err)
	}

	logger.LogOutput(posts, nil)
	return posts, nil
}
