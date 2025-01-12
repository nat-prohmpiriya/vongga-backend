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

	"go.opentelemetry.io/otel/trace"
)

type subPostRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
	rdb        *redis.Client
	tracer     trace.Tracer
}

func NewSubPostRepository(db *mongo.Database, rdb *redis.Client, trace trace.Tracer) domain.SubPostRepository {
	return &subPostRepository{
		db:         db,
		collection: db.Collection("subposts"),
		rdb:        rdb,
		tracer:     trace,
	}
}

func (r *subPostRepository) Create(ctx context.Context, subPost *domain.SubPost) error {
	ctx, span := r.tracer.Start(ctx, "SubPostRepository.Create")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(subPost)

	_, err := r.collection.InsertOne(ctx, subPost)
	if err != nil {
		logger.Output("SubPost created failed 1", err)
		return err
	}

	// Invalidate parent's subposts cache
	pattern := fmt.Sprintf("parent_subposts:%s:*", subPost.ParentID.Hex())
	keys, err := r.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.Output("SubPost created failed 2", err)
		return err
	}
	if len(keys) > 0 {
		err = r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			logger.Output("SubPost created failed 3", err)
			return err
		}
	}

	logger.Output("SubPost created successfully", nil)
	return nil
}

func (r *subPostRepository) Update(ctx context.Context, subPost *domain.SubPost) error {
	ctx, span := r.tracer.Start(ctx, "SubPostRepository.Update")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(subPost)

	filter := bson.M{"_id": subPost.ID}
	update := bson.M{"$set": subPost}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Output("SubPost updated failed 1", err)
		return err
	}

	// Invalidate subpost cache
	key := fmt.Sprintf("subpost:%s", subPost.ID.Hex())
	err = r.rdb.Del(ctx, key).Err()
	if err != nil {
		logger.Output("SubPost updated failed 2", err)
		return err
	}

	// Invalidate parent's subposts cache
	pattern := fmt.Sprintf("parent_subposts:%s:*", subPost.ParentID.Hex())
	keys, err := r.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.Output("SubPost updated failed 3", err)
		return err
	}
	if len(keys) > 0 {
		err = r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			logger.Output("SubPost updated failed 4", err)
			return err
		}
	}

	logger.Output("SubPost updated successfully", nil)
	return nil
}

func (r *subPostRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	ctx, span := r.tracer.Start(ctx, "SubPostRepository.Delete")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(id)

	// Find subpost first to get parentID for cache invalidation
	var subPost domain.SubPost
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&subPost)
	if err != nil {
		logger.Output("SubPost deleted failed 1", err)
		return err
	}

	filter := bson.M{"_id": id}
	_, err = r.collection.DeleteOne(ctx, filter)
	if err != nil {
		logger.Output("SubPost deleted failed 2", err)
		return err
	}

	// Invalidate subpost cache
	key := fmt.Sprintf("subpost:%s", id.Hex())
	err = r.rdb.Del(ctx, key).Err()
	if err != nil {
		logger.Output("SubPost deleted failed 3", err)
		return err
	}

	// Invalidate parent's subposts cache
	pattern := fmt.Sprintf("parent_subposts:%s:*", subPost.ParentID.Hex())
	keys, err := r.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.Output("SubPost deleted failed 4", err)
		return err
	}
	if len(keys) > 0 {
		err = r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			logger.Output("SubPost deleted failed 5", err)
			return err
		}
	}

	logger.Output("SubPost deleted successfully", nil)
	return nil
}

func (r *subPostRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.SubPost, error) {
	ctx, span := r.tracer.Start(ctx, "SubPostRepository.FindByID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(id)

	// Try to get from Redis first
	key := fmt.Sprintf("subpost:%s", id.Hex())
	subPostJSON, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		// Found in Redis
		var subPost domain.SubPost
		err = json.Unmarshal([]byte(subPostJSON), &subPost)
		if err != nil {
			logger.Output("Found subpost in Redis but failed to unmarshal 1", err)
			return nil, err
		}
		logger.Output(&subPost, nil)
		return &subPost, nil
	} else if err != redis.Nil {
		// Redis error
		logger.Output("Failed to get subpost from Redis 3", err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
	var subPost domain.SubPost
	filter := bson.M{"_id": id}
	err = r.collection.FindOne(ctx, filter).Decode(&subPost)
	if err != nil {
		logger.Output("Failed to get subpost from MongoDB 4", err)
		return nil, err
	}

	// Cache in Redis for 1 hour
	subPostBytes, err := json.Marshal(subPost)
	if err != nil {
		logger.Output("Failed to marshal subpost 5", err)
		return nil, err
	}

	err = r.rdb.Set(ctx, key, string(subPostBytes), time.Hour).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.Output("Failed to cache subpost in Redis 6", err)
	}

	logger.Output(subPost, nil)
	return &subPost, nil
}

func (r *subPostRepository) FindByParentID(ctx context.Context, parentID primitive.ObjectID, limit, offset int) ([]domain.SubPost, error) {
	ctx, span := r.tracer.Start(ctx, "SubPostRepository.FindByParentID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	input := map[string]interface{}{
		"parentID": parentID,
		"limit":    limit,
		"offset":   offset,
	}
	logger.Input(input)

	// Try to get from Redis first
	key := fmt.Sprintf("parent_subposts:%s:%d:%d", parentID.Hex(), limit, offset)
	subPostsJSON, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		// Found in Redis
		var subPosts []domain.SubPost
		err = json.Unmarshal([]byte(subPostsJSON), &subPosts)
		if err != nil {
			logger.Output("Failed to unmarshal subposts from Redis 1", err)
			return nil, err
		}
		logger.Output(subPosts, nil)
		return subPosts, nil
	} else if err != redis.Nil {
		// Redis error
		logger.Output("Failed to get subposts from Redis 2", err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
	var subPosts []domain.SubPost
	filter := bson.M{"parentId": parentID}

	findOptions := options.Find()
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}
	if offset > 0 {
		findOptions.SetSkip(int64(offset))
	}
	findOptions.SetSort(bson.D{{Key: "order", Value: 1}, {Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		logger.Output("Failed to get subposts from MongoDB 3", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &subPosts)
	if err != nil {
		logger.Output("Failed to get subposts from MongoDB 4", err)
		return nil, err
	}

	// Cache in Redis for 15 minutes
	subPostsBytes, err := json.Marshal(subPosts)
	if err != nil {
		logger.Output("Failed to marshal subposts 5", err)
		return nil, err
	}

	err = r.rdb.Set(ctx, key, string(subPostsBytes), 15*time.Minute).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.Output("Failed to cache subposts in Redis 6", err)
	}

	logger.Output(subPosts, nil)
	return subPosts, nil
}

func (r *subPostRepository) UpdateOrder(ctx context.Context, parentID primitive.ObjectID, orders map[primitive.ObjectID]int) error {
	ctx, span := r.tracer.Start(ctx, "SubPostRepository.UpdateOrder")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	input := map[string]interface{}{
		"parentID": parentID,
		"orders":   orders,
	}
	logger.Input(input)

	// Use bulk write to update multiple documents efficiently
	var operations []mongo.WriteModel
	for subPostID, order := range orders {
		filter := bson.M{"_id": subPostID, "parentId": parentID}
		update := bson.M{"$set": bson.M{"order": order}}
		operation := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update)
		operations = append(operations, operation)
	}

	if len(operations) > 0 {
		result, err := r.collection.BulkWrite(ctx, operations)
		if err != nil {
			logger.Output("Failed to update subposts order 1", err)
			return err
		}

		// Invalidate parent's subposts cache
		pattern := fmt.Sprintf("parent_subposts:%s:*", parentID.Hex())
		keys, err := r.rdb.Keys(ctx, pattern).Result()
		if err != nil {
			logger.Output("Failed to update subposts order 2", err)
			return err
		}
		if len(keys) > 0 {
			err = r.rdb.Del(ctx, keys...).Err()
			if err != nil {
				logger.Output("Failed to update subposts order 3", err)
				return err
			}
		}

		// Invalidate individual subpost caches
		for subPostID := range orders {
			key := fmt.Sprintf("subpost:%s", subPostID.Hex())
			err = r.rdb.Del(ctx, key).Err()
			if err != nil {
				logger.Output("Failed to update subposts order 4", err)
				return err
			}
		}

		logger.Output(map[string]interface{}{
			"message":       "SubPosts order updated successfully",
			"matchedCount":  result.MatchedCount,
			"modifiedCount": result.ModifiedCount,
		}, nil)
		return nil
	}

	logger.Output("No orders to update", nil)
	return nil
}

func (r *subPostRepository) DeleteByParentID(ctx context.Context, parentID primitive.ObjectID) error {
	ctx, span := r.tracer.Start(ctx, "SubPostRepository.DeleteByParentID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"parentID": parentID,
	})

	// Find all subposts first to invalidate their individual caches
	var subPosts []domain.SubPost
	cursor, err := r.collection.Find(ctx, bson.M{"parentId": parentID})
	if err != nil {
		logger.Output("SubPosts deleted failed 1", err)
		return err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &subPosts)
	if err != nil {
		logger.Output("SubPosts deleted failed 2", err)
		return err
	}

	filter := bson.M{"parentId": parentID}
	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		logger.Output(nil, err)
		return err
	}

	// Invalidate parent's subposts cache
	pattern := fmt.Sprintf("parent_subposts:%s:*", parentID.Hex())
	keys, err := r.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.Output("SubPosts deleted failed 3", err)
		return err
	}
	if len(keys) > 0 {
		err = r.rdb.Del(ctx, keys...).Err()
		if err != nil {
			logger.Output("SubPosts deleted failed 4", err)
			return err
		}
	}

	// Invalidate individual subpost caches
	for _, subPost := range subPosts {
		key := fmt.Sprintf("subpost:%s", subPost.ID.Hex())
		err = r.rdb.Del(ctx, key).Err()
		if err != nil {
			logger.Output("SubPosts deleted failed 5", err)
			return err
		}
	}

	logger.Output(map[string]interface{}{
		"message":      "SubPosts deleted successfully",
		"deletedCount": result.DeletedCount,
	}, nil)
	return nil
}
