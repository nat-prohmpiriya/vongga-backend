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

type commentRepository struct {
	db         *mongo.Database
	rdb        *redis.Client
	collection *mongo.Collection
	tracer     trace.Tracer
}

func NewCommentRepository(db *mongo.Database, rdb *redis.Client, tracer trace.Tracer) domain.CommentRepository {
	return &commentRepository{
		db:         db,
		rdb:        rdb,
		collection: db.Collection("comments"),
		tracer:     tracer,
	}
}

func (r *commentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	ctx, span := r.tracer.Start(ctx, "CommentRepository.Create")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(comment)

	_, err := r.collection.InsertOne(ctx, comment)
	if err != nil {
		logger.Output("failed to insert comment 1", err)
		return err
	}

	// Invalidate post comments cache
	pattern := fmt.Sprintf("post_comments:%s:*", comment.PostID.Hex())
	keys, err := r.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.Output("failed to get cache keys 2", err)
		return err
	}
	if len(keys) > 0 {
		err = r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			logger.Output("failed to delete cache keys 3", err)
			return err
		}
	}

	logger.Output("Comment created successfully", nil)
	return nil
}

func (r *commentRepository) Update(ctx context.Context, comment *domain.Comment) error {
	ctx, span := r.tracer.Start(ctx, "CommentRepository.Update")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(comment)

	filter := bson.M{"_id": comment.ID}
	update := bson.M{"$set": comment}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Output("failed to update comment 1", err)
		return err
	}

	// Invalidate comment cache and post comments cache
	commentKey := fmt.Sprintf("comment:%s", comment.ID.Hex())
	pattern := fmt.Sprintf("post_comments:%s:*", comment.PostID.Hex())

	// Delete comment cache
	err = r.rdb.Del(ctx, commentKey).Err()
	if err != nil {
		logger.Output("failed to delete comment cache 2", err)
		return err
	}

	// Delete post comments cache
	keys, err := r.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.Output("failed to get post comments cache keys 3", err)
		return err
	}
	if len(keys) > 0 {
		err = r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			logger.Output("failed to delete post comments cache keys 4", err)
			return err
		}
	}

	logger.Output("Comment updated successfully", nil)
	return nil
}

func (r *commentRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	ctx, span := r.tracer.Start(ctx, "CommentRepository.Delete")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(id)

	// Find comment first to get postID for cache invalidation
	var comment domain.Comment
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&comment)
	if err != nil {
		logger.Output("failed to find comment 1", err)
		return err
	}

	filter := bson.M{"_id": id}
	_, err = r.collection.DeleteOne(ctx, filter)
	if err != nil {
		logger.Output("failed to delete comment 2", err)
		return err
	}

	// Invalidate comment cache and post comments cache
	commentKey := fmt.Sprintf("comment:%s", id.Hex())
	pattern := fmt.Sprintf("post_comments:%s:*", comment.PostID.Hex())

	// Delete comment cache
	err = r.rdb.Del(ctx, commentKey).Err()
	if err != nil {
		logger.Output("failed to delete comment cache 3", err)
		return err
	}

	// Delete post comments cache
	keys, err := r.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.Output("failed to get post comments cache keys 4", err)
		return err
	}
	if len(keys) > 0 {
		err = r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			logger.Output("failed to delete post comments cache keys 5", err)
			return err
		}
	}

	logger.Output("Comment deleted successfully", nil)
	return nil
}

func (r *commentRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Comment, error) {
	ctx, span := r.tracer.Start(ctx, "CommentRepository.FindByID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(id)

	key := fmt.Sprintf("comment:%s", id.Hex())

	// Try to get from Redis first
	commentJSON, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		// Found in Redis
		var comment domain.Comment
		err = json.Unmarshal([]byte(commentJSON), &comment)
		if err != nil {
			logger.Output("failed to unmarshal comment from cache 1", err)
			return nil, err
		}
		logger.Output(&comment, nil)
		return &comment, nil
	} else if err != redis.Nil {
		// Redis error
		logger.Output("redis error 2", err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
	var comment domain.Comment
	filter := bson.M{"_id": id}
	err = r.collection.FindOne(ctx, filter).Decode(&comment)
	if err != nil {
		logger.Output("failed to find comment 3", err)
		return nil, err
	}

	// Cache in Redis for 30 minutes
	commentBytes, err := json.Marshal(&comment)
	if err != nil {
		logger.Output("failed to marshal comment 4", err)
		return nil, err
	}

	err = r.rdb.Set(ctx, key, string(commentBytes), 30*time.Minute).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.Output("failed to set comment cache 5", err)
	}

	logger.Output(&comment, nil)
	return &comment, nil
}

func (r *commentRepository) FindByPostID(ctx context.Context, postID primitive.ObjectID, limit, offset int) ([]domain.Comment, error) {
	ctx, span := r.tracer.Start(ctx, "CommentRepository.FindByPostID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"postID": postID.Hex(),
		"limit":  limit,
		"offset": offset,
	})

	key := fmt.Sprintf("post_comments:%s:%d:%d", postID.Hex(), limit, offset)

	// Try to get from Redis first
	commentsJSON, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		// Found in Redis
		var comments []domain.Comment
		err = json.Unmarshal([]byte(commentsJSON), &comments)
		if err != nil {
			logger.Output("failed to unmarshal comments from cache 1", err)
			return nil, err
		}
		logger.Output(comments, nil)
		return comments, nil
	} else if err != redis.Nil {
		// Redis error
		logger.Output("redis error 2", err)
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
		logger.Output("failed to find comments 3", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &comments); err != nil {
		logger.Output("failed to decode comments 4", err)
		return nil, err
	}

	// Cache in Redis for 30 minutes
	commentsBytes, err := json.Marshal(comments)
	if err != nil {
		logger.Output("failed to marshal comments 5", err)
		return nil, err
	}

	err = r.rdb.Set(ctx, key, string(commentsBytes), 30*time.Minute).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.Output("failed to set comments cache 6", err)
	}

	logger.Output(comments, nil)
	return comments, nil
}

func (r *commentRepository) DeleteByPostID(ctx context.Context, postID primitive.ObjectID) error {
	ctx, span := r.tracer.Start(ctx, "CommentRepository.DeleteByPostID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(postID)

	filter := bson.M{"postId": postID}
	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		logger.Output("failed to delete comments 1", err)
		return err
	}

	// Invalidate post comments cache
	pattern := fmt.Sprintf("post_comments:%s:*", postID.Hex())
	keys, err := r.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.Output("failed to get cache keys 2", err)
		return err
	}
	if len(keys) > 0 {
		err = r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			logger.Output("failed to delete cache keys 3", err)
			return err
		}
	}

	logger.Output(map[string]interface{}{
		"message":      "Comments deleted successfully",
		"deletedCount": result.DeletedCount,
	}, nil)
	return nil
}
