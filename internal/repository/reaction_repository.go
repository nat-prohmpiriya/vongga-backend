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

type reactionRepository struct {
	db *mongo.Database
}

func NewReactionRepository(db *mongo.Database) domain.ReactionRepository {
	return &reactionRepository{
		db: db,
	}
}

func (r *reactionRepository) Create(reaction *domain.Reaction) error {
	logger := utils.NewTraceLogger("ReactionRepository.Create")
	logger.Input(reaction)

	reaction.CreatedAt = time.Now()
	reaction.UpdatedAt = time.Now()

	result, err := r.db.Collection("reactions").InsertOne(context.Background(), reaction)
	if err != nil {
		logger.Output(nil, err)
		return err
	}

	reaction.ID = result.InsertedID.(primitive.ObjectID)
	logger.Output(reaction, nil)
	return nil
}

func (r *reactionRepository) Update(reaction *domain.Reaction) error {
	logger := utils.NewTraceLogger("ReactionRepository.Update")
	logger.Input(reaction)

	reaction.UpdatedAt = time.Now()

	filter := bson.M{"_id": reaction.ID}
	update := bson.M{"$set": reaction}

	_, err := r.db.Collection("reactions").UpdateOne(context.Background(), filter, update)
	logger.Output(nil, err)
	return err
}

func (r *reactionRepository) Delete(id primitive.ObjectID) error {
	logger := utils.NewTraceLogger("ReactionRepository.Delete")
	logger.Input(id)

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"deletedAt": time.Now(),
			"isActive":  false,
		},
	}

	_, err := r.db.Collection("reactions").UpdateOne(context.Background(), filter, update)
	logger.Output(nil, err)
	return err
}

func (r *reactionRepository) FindByID(id primitive.ObjectID) (*domain.Reaction, error) {
	logger := utils.NewTraceLogger("ReactionRepository.FindByID")
	logger.Input(id)

	var reaction domain.Reaction
	err := r.db.Collection("reactions").FindOne(context.Background(), bson.M{"_id": id, "deletedAt": bson.M{"$exists": false}}).Decode(&reaction)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Output(nil, domain.ErrNotFound)
			return nil, domain.ErrNotFound
		}
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(&reaction, nil)
	return &reaction, nil
}

func (r *reactionRepository) FindByPostID(postID primitive.ObjectID, limit, offset int) ([]domain.Reaction, error) {
	logger := utils.NewTraceLogger("ReactionRepository.FindByPostID")
	logger.Input(postID, limit, offset)

	opts := options.Find().
		// SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.db.Collection("reactions").Find(context.Background(), bson.M{"postId": postID, "deletedAt": bson.M{"$exists": false}}, opts)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var reactions []domain.Reaction
	if err = cursor.All(context.Background(), &reactions); err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(reactions, nil)
	return reactions, nil
}

func (r *reactionRepository) FindByCommentID(commentID primitive.ObjectID, limit, offset int) ([]domain.Reaction, error) {
	logger := utils.NewTraceLogger("ReactionRepository.FindByCommentID")
	logger.Input(commentID, limit, offset)

	opts := options.Find().
		// SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.db.Collection("reactions").Find(context.Background(), bson.M{"commentId": commentID, "deletedAt": bson.M{"$exists": false}}, opts)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var reactions []domain.Reaction
	if err = cursor.All(context.Background(), &reactions); err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(reactions, nil)
	return reactions, nil
}

func (r *reactionRepository) FindByUserAndTarget(userID, postID primitive.ObjectID, commentID *primitive.ObjectID) (*domain.Reaction, error) {
	logger := utils.NewTraceLogger("ReactionRepository.FindByUserAndTarget")
	logger.Input(userID, postID, commentID)

	filter := bson.M{
		"userId": userID,
		"postId": postID,
	}
	if commentID != nil {
		filter["commentId"] = commentID
	}

	var reaction domain.Reaction
	err := r.db.Collection("reactions").FindOne(context.Background(), filter).Decode(&reaction)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Output(nil, nil)
			return nil, nil
		}
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(&reaction, nil)
	return &reaction, nil
}
