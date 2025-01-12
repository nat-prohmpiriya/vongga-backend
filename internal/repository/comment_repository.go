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

type commentRepository struct {
	db         *mongo.Database
	rdb        *redis.Client
	collection *mongo.Collection
}

func NewCommentRepository(db *mongo.Database, rdb *redis.Client) domain.CommentRepository {
	return &commentRepository{
		db:         db,
		rdb:        rdb,
		collection: db.Collection("comments"),
	}
}

func (r *commentRepository) Create(comment *domain.Comment) error {
	logger := utils.NewLogger("CommentRepository.Create")
	logger.LogInput(comment)

	_, err := r.collection.InsertOne(context.Background(), comment)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Invalidate post comments cache
	ctx := context.Background()
	pattern := fmt.Sprintf("post_comments:%s:*", comment.PostID.Hex())
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

	logger.LogOutput("Comment created successfully", nil)
	return nil
}

func (r *commentRepository) Update(comment *domain.Comment) error {
	logger := utils.NewLogger("CommentRepository.Update")
	logger.LogInput(comment)

	filter := bson.M{"_id": comment.ID}
	update := bson.M{"$set": comment}
	_, err := r.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Invalidate comment cache and post comments cache
	ctx := context.Background()
	commentKey := fmt.Sprintf("comment:%s", comment.ID.Hex())
	pattern := fmt.Sprintf("post_comments:%s:*", comment.PostID.Hex())

	// Delete comment cache
	err = r.rdb.Del(ctx, commentKey).Err()
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Delete post comments cache
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

	logger.LogOutput("Comment updated successfully", nil)
	return nil
}

func (r *commentRepository) Delete(id primitive.ObjectID) error {
	logger := utils.NewLogger("CommentRepository.Delete")
	logger.LogInput(id)

	// Find comment first to get postID for cache invalidation
	var comment domain.Comment
	err := r.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&comment)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	filter := bson.M{"_id": id}
	_, err = r.collection.DeleteOne(context.Background(), filter)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Invalidate comment cache and post comments cache
	ctx := context.Background()
	commentKey := fmt.Sprintf("comment:%s", id.Hex())
	pattern := fmt.Sprintf("post_comments:%s:*", comment.PostID.Hex())

	// Delete comment cache
	err = r.rdb.Del(ctx, commentKey).Err()
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Delete post comments cache
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

	logger.LogOutput("Comment deleted successfully", nil)
	return nil
}

func (r *commentRepository) FindByID(id primitive.ObjectID) (*domain.Comment, error) {
	logger := utils.NewLogger("CommentRepository.FindByID")
	logger.LogInput(id)

	ctx := context.Background()
	key := fmt.Sprintf("comment:%s", id.Hex())

	// Try to get from Redis first
	commentJSON, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		// Found in Redis
		var comment domain.Comment
		err = json.Unmarshal([]byte(commentJSON), &comment)
		if err != nil {
			logger.LogOutput(nil, err)
			return nil, err
		}
		logger.LogOutput(&comment, nil)
		return &comment, nil
	} else if err != redis.Nil {
		// Redis error
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
	var comment domain.Comment
	filter := bson.M{"_id": id}
	err = r.collection.FindOne(ctx, filter).Decode(&comment)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Cache in Redis for 30 minutes
	commentBytes, err := json.Marshal(&comment)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	err = r.rdb.Set(ctx, key, string(commentBytes), 30*time.Minute).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.LogOutput(nil, err)
	}

	logger.LogOutput(&comment, nil)
	return &comment, nil
}

func (r *commentRepository) FindByPostID(postID primitive.ObjectID, limit, offset int) ([]domain.Comment, error) {
	logger := utils.NewLogger("CommentRepository.FindByPostID")
	input := map[string]interface{}{
		"postID": postID,
		"limit":  limit,
		"offset": offset,
	}
	logger.LogInput(input)

	ctx := context.Background()
	key := fmt.Sprintf("post_comments:%s:%d:%d", postID.Hex(), limit, offset)

	// Try to get from Redis first
	commentsJSON, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		// Found in Redis
		var comments []domain.Comment
		err = json.Unmarshal([]byte(commentsJSON), &comments)
		if err != nil {
			logger.LogOutput(nil, err)
			return nil, err
		}
		logger.LogOutput(comments, nil)
		return comments, nil
	} else if err != redis.Nil {
		// Redis error
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
	var comments []domain.Comment
	filter := bson.M{"postId": postID}

	findOptions := options.Find()
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}
	if offset > 0 {
		findOptions.SetSkip(int64(offset))
	}
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &comments)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Cache in Redis for 10 minutes
	commentsBytes, err := json.Marshal(comments)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	err = r.rdb.Set(ctx, key, string(commentsBytes), 10*time.Minute).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.LogOutput(nil, err)
	}

	logger.LogOutput(comments, nil)
	return comments, nil
}

func (r *commentRepository) DeleteByPostID(postID primitive.ObjectID) error {
	logger := utils.NewLogger("CommentRepository.DeleteByPostID")
	logger.LogInput(postID)

	filter := bson.M{"postId": postID}
	result, err := r.collection.DeleteMany(context.Background(), filter)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Invalidate post comments cache
	ctx := context.Background()
	pattern := fmt.Sprintf("post_comments:%s:*", postID.Hex())
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

	logger.LogOutput(map[string]interface{}{
		"message":      "Comments deleted successfully",
		"deletedCount": result.DeletedCount,
	}, nil)
	return nil
}
