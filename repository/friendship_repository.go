package repository

import (
	"context"
	"time"

	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type friendshipRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

// NewFriendshipRepository creates a new instance of FriendshipRepository
func NewFriendshipRepository(db *mongo.Database) domain.FriendshipRepository {
	return &friendshipRepository{
		db:         db,
		collection: db.Collection("friendships"),
	}
}

func (r *friendshipRepository) Create(friendship *domain.Friendship) error {
	logger := utils.NewLogger("FriendshipRepository.Create")
	logger.LogInput(friendship)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, friendship)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(friendship, nil)
	return nil
}

func (r *friendshipRepository) Update(friendship *domain.Friendship) error {
	logger := utils.NewLogger("FriendshipRepository.Update")
	logger.LogInput(friendship)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
		logger.LogOutput(nil, err)
		return err
	}

	if result.MatchedCount == 0 {
		err = domain.ErrNotFound
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(result, nil)
	return nil
}

func (r *friendshipRepository) Delete(userID1, userID2 primitive.ObjectID) error {
	logger := utils.NewLogger("FriendshipRepository.Delete")
	input := map[string]interface{}{
		"userID1": userID1.Hex(),
		"userID2": userID2.Hex(),
	}
	logger.LogInput(input)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
		logger.LogOutput(nil, err)
		return err
	}

	if result.DeletedCount == 0 {
		err = domain.ErrNotFound
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(result, nil)
	return nil
}

func (r *friendshipRepository) FindByUsers(userID1, userID2 primitive.ObjectID) (*domain.Friendship, error) {
	logger := utils.NewLogger("FriendshipRepository.FindByUsers")
	input := map[string]interface{}{
		"userID1": userID1.Hex(),
		"userID2": userID2.Hex(),
	}
	logger.LogInput(input)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
			logger.LogOutput(nil, domain.ErrNotFound)
			return nil, domain.ErrNotFound
		}
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(&friendship, nil)
	return &friendship, nil
}

func (r *friendshipRepository) FindFriends(userID primitive.ObjectID, limit, offset int) ([]domain.Friendship, error) {
	logger := utils.NewLogger("FriendshipRepository.FindFriends")
	input := map[string]interface{}{
		"userID": userID.Hex(),
		"limit":  limit,
		"offset": offset,
	}
	logger.LogInput(input)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
		logger.LogOutput(nil, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var friendships []domain.Friendship
	if err = cursor.All(ctx, &friendships); err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(friendships, nil)
	return friendships, nil
}

func (r *friendshipRepository) FindPendingRequests(userID primitive.ObjectID, limit, offset int) ([]domain.Friendship, error) {
	logger := utils.NewLogger("FriendshipRepository.FindPendingRequests")
	input := map[string]interface{}{
		"userID": userID.Hex(),
		"limit":  limit,
		"offset": offset,
	}
	logger.LogInput(input)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
		logger.LogOutput(nil, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var friendships []domain.Friendship
	if err = cursor.All(ctx, &friendships); err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(friendships, nil)
	return friendships, nil
}

func (r *friendshipRepository) CountFriends(userID primitive.ObjectID) (int64, error) {
	logger := utils.NewLogger("FriendshipRepository.CountFriends")
	input := map[string]interface{}{
		"userID": userID.Hex(),
	}
	logger.LogInput(input)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
		logger.LogOutput(nil, err)
		return 0, err
	}

	logger.LogOutput(count, nil)
	return count, nil
}

func (r *friendshipRepository) CountPendingRequests(userID primitive.ObjectID) (int64, error) {
	logger := utils.NewLogger("FriendshipRepository.CountPendingRequests")
	input := map[string]interface{}{
		"userID": userID.Hex(),
	}
	logger.LogInput(input)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"userId2": userID,
		"status":  "pending",
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		logger.LogOutput(nil, err)
		return 0, err
	}

	logger.LogOutput(count, nil)
	return count, nil
}
