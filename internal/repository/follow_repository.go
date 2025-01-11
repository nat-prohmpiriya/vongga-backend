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
)

type followRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

// NewFollowRepository creates a new instance of FollowRepository
func NewFollowRepository(db *mongo.Database) domain.FollowRepository {
	return &followRepository{
		db:         db,
		collection: db.Collection("follows"),
	}
}

func (r *followRepository) Create(follow *domain.Follow) error {
	logger := utils.NewLogger("FollowRepository.Create")
	logger.LogInput(follow)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, follow)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(follow, nil)
	return nil
}

func (r *followRepository) Delete(followerID, followingID primitive.ObjectID) error {
	logger := utils.NewLogger("FollowRepository.Delete")
	input := map[string]interface{}{
		"followerID":  followerID.Hex(),
		"followingID": followingID.Hex(),
	}
	logger.LogInput(input)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"followerId":  followerID,
		"followingId": followingID,
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

func (r *followRepository) FindByFollowerAndFollowing(followerID, followingID primitive.ObjectID) (*domain.Follow, error) {
	logger := utils.NewLogger("FollowRepository.FindByFollowerAndFollowing")
	input := map[string]interface{}{
		"followerID":  followerID.Hex(),
		"followingID": followingID.Hex(),
	}
	logger.LogInput(input)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"followerId":  followerID,
		"followingId": followingID,
	}

	var follow domain.Follow
	err := r.collection.FindOne(ctx, filter).Decode(&follow)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.LogOutput(nil, domain.ErrNotFound)
			return nil, domain.ErrNotFound
		}
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(&follow, nil)
	return &follow, nil
}

func (r *followRepository) FindFollowers(userID primitive.ObjectID, limit, offset int) ([]domain.Follow, error) {
	logger := utils.NewLogger("FollowRepository.FindFollowers")
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
		"followingId": userID,
		"status":      "active",
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var follows []domain.Follow
	if err = cursor.All(ctx, &follows); err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(follows, nil)
	return follows, nil
}

func (r *followRepository) FindFollowing(userID primitive.ObjectID, limit, offset int) ([]domain.Follow, error) {
	logger := utils.NewLogger("FollowRepository.FindFollowing")
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
		"followerId": userID,
		"status":     "active",
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var follows []domain.Follow
	if err = cursor.All(ctx, &follows); err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(follows, nil)
	return follows, nil
}

func (r *followRepository) CountFollowers(userID primitive.ObjectID) (int64, error) {
	logger := utils.NewLogger("FollowRepository.CountFollowers")
	input := map[string]interface{}{
		"userID": userID.Hex(),
	}
	logger.LogInput(input)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"followingId": userID,
		"status":      "active",
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		logger.LogOutput(nil, err)
		return 0, err
	}

	logger.LogOutput(count, nil)
	return count, nil
}

func (r *followRepository) CountFollowing(userID primitive.ObjectID) (int64, error) {
	logger := utils.NewLogger("FollowRepository.CountFollowing")
	input := map[string]interface{}{
		"userID": userID.Hex(),
	}
	logger.LogInput(input)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"followerId": userID,
		"status":     "active",
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		logger.LogOutput(nil, err)
		return 0, err
	}

	logger.LogOutput(count, nil)
	return count, nil
}

func (r *followRepository) UpdateStatus(followerID, followingID primitive.ObjectID, status string) error {
	logger := utils.NewLogger("FollowRepository.UpdateStatus")
	input := map[string]interface{}{
		"followerID":  followerID.Hex(),
		"followingID": followingID.Hex(),
		"status":      status,
	}
	logger.LogInput(input)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"followerId":  followerID,
		"followingId": followingID,
	}

	update := bson.M{
		"$set": bson.M{
			"status": status,
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
