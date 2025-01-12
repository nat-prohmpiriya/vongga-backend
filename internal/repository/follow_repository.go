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
	"go.opentelemetry.io/otel/trace"
)

type followRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
	tracer     trace.Tracer
}

func NewFollowRepository(db *mongo.Database, tracer trace.Tracer) domain.FollowRepository {
	return &followRepository{
		db:         db,
		collection: db.Collection("follows"),
		tracer:     tracer,
	}
}

func (r *followRepository) Create(ctx context.Context, follow *domain.Follow) error {
	ctx, span := r.tracer.Start(ctx, "FollowRepository.Create")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(follow)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, follow)
	if err != nil {
		logger.Output("failed to insert follow 1", err)
		return err
	}

	logger.Output(follow, nil)
	return nil
}

func (r *followRepository) Delete(ctx context.Context, followerID, followingID primitive.ObjectID) error {
	ctx, span := r.tracer.Start(ctx, "FollowRepository.Delete")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"followerID":  followerID.Hex(),
		"followingID": followingID.Hex(),
	})

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	filter := bson.M{
		"followerId":  followerID,
		"followingId": followingID,
	}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		logger.Output("failed to delete follow 1", err)
		return err
	}

	if result.DeletedCount == 0 {
		err = domain.ErrNotFound
		logger.Output("follow not found 2", err)
		return err
	}

	logger.Output(result, nil)
	return nil
}

func (r *followRepository) FindByFollowerAndFollowing(ctx context.Context, followerID, followingID primitive.ObjectID) (*domain.Follow, error) {
	ctx, span := r.tracer.Start(ctx, "FollowRepository.FindByFollowerAndFollowing")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"followerID":  followerID.Hex(),
		"followingID": followingID.Hex(),
	})

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	filter := bson.M{
		"followerId":  followerID,
		"followingId": followingID,
	}

	var follow domain.Follow
	err := r.collection.FindOne(ctx, filter).Decode(&follow)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Output("follow not found 1", domain.ErrNotFound)
			return nil, domain.ErrNotFound
		}
		logger.Output("failed to find follow 2", err)
		return nil, err
	}

	logger.Output(&follow, nil)
	return &follow, nil
}

func (r *followRepository) FindFollowers(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]domain.Follow, error) {
	ctx, span := r.tracer.Start(ctx, "FollowRepository.FindFollowers")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"userID": userID.Hex(),
		"limit":  limit,
		"offset": offset,
	})

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	filter := bson.M{
		"followingId": userID,
		"status":      "active",
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		logger.Output("failed to find followers 1", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var follows []domain.Follow
	if err = cursor.All(ctx, &follows); err != nil {
		logger.Output("failed to decode followers 2", err)
		return nil, err
	}

	logger.Output(follows, nil)
	return follows, nil
}

func (r *followRepository) FindFollowing(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]domain.Follow, error) {
	ctx, span := r.tracer.Start(ctx, "FollowRepository.FindFollowing")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"userID": userID.Hex(),
		"limit":  limit,
		"offset": offset,
	})

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	filter := bson.M{
		"followerId": userID,
		"status":     "active",
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		logger.Output("failed to find following 1", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var follows []domain.Follow
	if err = cursor.All(ctx, &follows); err != nil {
		logger.Output("failed to decode following 2", err)
		return nil, err
	}

	logger.Output(follows, nil)
	return follows, nil
}

func (r *followRepository) CountFollowers(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	ctx, span := r.tracer.Start(ctx, "FollowRepository.CountFollowers")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"userID": userID.Hex(),
	})

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	filter := bson.M{
		"followingId": userID,
		"status":      "active",
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		logger.Output("failed to count followers 1", err)
		return 0, err
	}

	logger.Output(count, nil)
	return count, nil
}

func (r *followRepository) CountFollowing(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	ctx, span := r.tracer.Start(ctx, "FollowRepository.CountFollowing")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"userID": userID.Hex(),
	})

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	filter := bson.M{
		"followerId": userID,
		"status":     "active",
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		logger.Output("failed to count following 1", err)
		return 0, err
	}

	logger.Output(count, nil)
	return count, nil
}

func (r *followRepository) UpdateStatus(ctx context.Context, followerID, followingID primitive.ObjectID, status string) error {
	ctx, span := r.tracer.Start(ctx, "FollowRepository.UpdateStatus")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"followerID":  followerID.Hex(),
		"followingID": followingID.Hex(),
		"status":      status,
	})

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	filter := bson.M{
		"followerId":  followerID,
		"followingId": followingID,
	}

	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Output("failed to update follow status 1", err)
		return err
	}

	if result.MatchedCount == 0 {
		err = domain.ErrNotFound
		logger.Output("follow not found 2", err)
		return err
	}

	logger.Output(result, nil)
	return nil
}
