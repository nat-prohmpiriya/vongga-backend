package repository

import (
	"context"
	"errors"
	"time"

	"vongga_api/internal/domain"
	"vongga_api/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/trace"
)

type friendshipRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
	tracer     trace.Tracer
}

// NewFriendshipRepository creates a new instance of FriendshipRepository
func NewFriendshipRepository(db *mongo.Database, tracer trace.Tracer) domain.FriendshipRepository {
	return &friendshipRepository{
		db:         db,
		collection: db.Collection("friendships"),
		tracer:     tracer,
	}
}

func (r *friendshipRepository) Create(ctx context.Context, friendship *domain.Friendship) error {
	ctx, span := r.tracer.Start(ctx, "FriendshipRepository.Create")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(friendship)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, friendship)
	if err != nil {
		logger.Output("failed to insert friendship 1", err)
		return err
	}

	logger.Output(friendship, nil)
	return nil
}

func (r *friendshipRepository) Update(ctx context.Context, friendship *domain.Friendship) error {
	ctx, span := r.tracer.Start(ctx, "FriendshipRepository.Update")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(friendship)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": friendship.ID}
	update := bson.M{
		"$set": bson.M{
			"status":    friendship.Status,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Output("failed to update friendship 1", err)
		return err
	}

	if result.MatchedCount == 0 {
		err := domain.ErrNotFound
		logger.Output("friendship not found 2", err)
		return err
	}

	logger.Output(result, nil)
	return nil
}

func (r *friendshipRepository) Delete(ctx context.Context, userID1, userID2 primitive.ObjectID) error {
	ctx, span := r.tracer.Start(ctx, "FriendshipRepository.Delete")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	input := map[string]interface{}{
		"userID1": userID1.Hex(),
		"userID2": userID2.Hex(),
	}
	logger.Input(input)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{
				"userId1": userID1,
				"userId2": userID2,
			},
			{
				"userId1": userID2,
				"userId2": userID1,
			},
		},
	}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		logger.Output("failed to delete friendship 3", err)
		return err
	}

	if result.DeletedCount == 0 {
		err := domain.ErrNotFound
		logger.Output("friendship not found 4", err)
		return err
	}

	logger.Output(result, nil)
	return nil
}

func (r *friendshipRepository) FindByUsers(ctx context.Context, userID1, userID2 primitive.ObjectID) (*domain.Friendship, error) {
	ctx, span := r.tracer.Start(ctx, "FriendshipRepository.FindByUsers")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	input := map[string]interface{}{
		"userID1": userID1.Hex(),
		"userID2": userID2.Hex(),
	}
	logger.Input(input)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{
				"userId1": userID1,
				"userId2": userID2,
			},
			{
				"userId1": userID2,
				"userId2": userID1,
			},
		},
	}

	var friendship domain.Friendship
	err := r.collection.FindOne(ctx, filter).Decode(&friendship)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Output("friendship not found 1", nil)
			return nil, domain.ErrNotFound
		}
		logger.Output("failed to find friendship 2", err)
		return nil, err
	}

	logger.Output(&friendship, nil)
	return &friendship, nil
}

func (r *friendshipRepository) FindFriends(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]domain.Friendship, error) {
	ctx, span := r.tracer.Start(ctx, "FriendshipRepository.FindFriends")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	input := map[string]interface{}{
		"userID": userID.Hex(),
		"limit":  limit,
		"offset": offset,
	}
	logger.Input(input)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "updatedAt", Value: -1}})

	filter := bson.M{
		"$and": []bson.M{
			{
				"$or": []bson.M{
					{"userId1": userID},
					{"userId2": userID},
				},
			},
			{"status": "accepted"},
		},
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		logger.Output("failed to find friends 1", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var friendships []domain.Friendship
	if err = cursor.All(ctx, &friendships); err != nil {
		logger.Output("failed to decode friendships 2", err)
		return nil, err
	}

	logger.Output(friendships, nil)
	return friendships, nil
}

func (r *friendshipRepository) FindPendingRequests(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]domain.Friendship, error) {
	ctx, span := r.tracer.Start(ctx, "FriendshipRepository.FindPendingRequests")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	input := map[string]interface{}{
		"userID": userID.Hex(),
		"limit":  limit,
		"offset": offset,
	}
	logger.Input(input)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	filter := bson.M{
		"userId2": userID,
		"status":  "pending",
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		logger.Output("failed to find pending requests 1", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var friendships []domain.Friendship
	if err = cursor.All(ctx, &friendships); err != nil {
		logger.Output("failed to decode pending requests 2", err)
		return nil, err
	}

	logger.Output(friendships, nil)
	return friendships, nil
}

func (r *friendshipRepository) CountFriends(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	ctx, span := r.tracer.Start(ctx, "FriendshipRepository.CountFriends")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	input := map[string]interface{}{
		"userID": userID.Hex(),
	}
	logger.Input(input)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	filter := bson.M{
		"$and": []bson.M{
			{
				"$or": []bson.M{
					{"userId1": userID},
					{"userId2": userID},
				},
			},
			{"status": "accepted"},
		},
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		logger.Output("failed to count friends 1", err)
		return 0, err
	}

	logger.Output(count, nil)
	return count, nil
}

func (r *friendshipRepository) CountPendingRequests(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	ctx, span := r.tracer.Start(ctx, "FriendshipRepository.CountPendingRequests")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	input := map[string]interface{}{
		"userID": userID.Hex(),
	}
	logger.Input(input)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	filter := bson.M{
		"userId2": userID,
		"status":  "pending",
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		logger.Output("failed to count pending requests 1", err)
		return 0, err
	}

	logger.Output(count, nil)
	return count, nil
}

func (r *friendshipRepository) FindByUserAndTarget(ctx context.Context, userID, targetID primitive.ObjectID) (*domain.Friendship, error) {
	ctx, span := r.tracer.Start(ctx, "FriendshipRepository.FindByUserAndTarget")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"userID":   userID.Hex(),
		"targetID": targetID.Hex(),
	})

	filter := bson.M{
		"$or": []bson.M{
			{
				"userId":   userID,
				"friendId": targetID,
			},
			{
				"userId":   targetID,
				"friendId": userID,
			},
		},
	}

	var friendship domain.Friendship
	err := r.collection.FindOne(ctx, filter).Decode(&friendship)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Output("friendship not found 1", nil)
			return nil, nil
		}
		logger.Output("failed to find friendship 2", err)
		return nil, err
	}

	logger.Output(friendship, nil)
	return &friendship, nil
}

func (r *friendshipRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Friendship, error) {
	ctx, span := r.tracer.Start(ctx, "FriendshipRepository.FindByID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(id)

	filter := bson.M{"_id": id}

	var friendship domain.Friendship
	err := r.collection.FindOne(ctx, filter).Decode(&friendship)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Output("friendship not found 1", nil)
			return nil, nil
		}
		logger.Output("failed to find friendship 2", err)
		return nil, err
	}

	logger.Output(friendship, nil)
	return &friendship, nil
}

func (r *friendshipRepository) RemoveFriend(ctx context.Context, userID, targetID primitive.ObjectID) error {
	ctx, span := r.tracer.Start(ctx, "FriendshipRepository.RemoveFriend")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"userID":   userID.Hex(),
		"targetID": targetID.Hex(),
	})

	filter := bson.M{
		"$or": []bson.M{
			{
				"userId":   userID,
				"friendId": targetID,
			},
			{
				"userId":   targetID,
				"friendId": userID,
			},
		},
	}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		logger.Output("failed to remove friend 3", err)
		return err
	}

	if result.DeletedCount == 0 {
		logger.Output("friendship not found 4", errors.New("friendship not found"))
		return errors.New("friendship not found")
	}

	logger.Output("Friendship removed successfully", nil)
	return nil
}
