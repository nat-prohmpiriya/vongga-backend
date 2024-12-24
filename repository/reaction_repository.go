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

type reactionRepository struct {
	db *mongo.Database
}

func NewReactionRepository(db *mongo.Database) domain.ReactionRepository {
	return &reactionRepository{
		db: db,
	}
}

func (r *reactionRepository) Create(reaction *domain.Reaction) error {
	logger := utils.NewLogger("ReactionRepository.Create")
	logger.LogInput(reaction)

	reaction.CreatedAt = time.Now()
	reaction.UpdatedAt = time.Now()

	result, err := r.db.Collection("reactions").InsertOne(context.Background(), reaction)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	reaction.ID = result.InsertedID.(primitive.ObjectID)
	logger.LogOutput(reaction, nil)
	return nil
}

func (r *reactionRepository) Update(reaction *domain.Reaction) error {
	logger := utils.NewLogger("ReactionRepository.Update")
	logger.LogInput(reaction)

	reaction.UpdatedAt = time.Now()

	filter := bson.M{"_id": reaction.ID}
	update := bson.M{"$set": reaction}

	_, err := r.db.Collection("reactions").UpdateOne(context.Background(), filter, update)
	logger.LogOutput(nil, err)
	return err
}

func (r *reactionRepository) Delete(id primitive.ObjectID) error {
	logger := utils.NewLogger("ReactionRepository.Delete")
	logger.LogInput(id)

	filter := bson.M{"_id": id}
	_, err := r.db.Collection("reactions").DeleteOne(context.Background(), filter)
	logger.LogOutput(nil, err)
	return err
}

func (r *reactionRepository) FindByID(id primitive.ObjectID) (*domain.Reaction, error) {
	logger := utils.NewLogger("ReactionRepository.FindByID")
	logger.LogInput(id)

	var reaction domain.Reaction
	err := r.db.Collection("reactions").FindOne(context.Background(), bson.M{"_id": id}).Decode(&reaction)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.LogOutput(nil, nil)
			return nil, nil
		}
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(&reaction, nil)
	return &reaction, nil
}

func (r *reactionRepository) FindByPostID(postID primitive.ObjectID, limit, offset int) ([]domain.Reaction, error) {
	logger := utils.NewLogger("ReactionRepository.FindByPostID")
	logger.LogInput(postID, limit, offset)

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.db.Collection("reactions").Find(context.Background(), bson.M{"postId": postID}, opts)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var reactions []domain.Reaction
	if err = cursor.All(context.Background(), &reactions); err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(reactions, nil)
	return reactions, nil
}

func (r *reactionRepository) FindByCommentID(commentID primitive.ObjectID, limit, offset int) ([]domain.Reaction, error) {
	logger := utils.NewLogger("ReactionRepository.FindByCommentID")
	logger.LogInput(commentID, limit, offset)

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.db.Collection("reactions").Find(context.Background(), bson.M{"commentId": commentID}, opts)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var reactions []domain.Reaction
	if err = cursor.All(context.Background(), &reactions); err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(reactions, nil)
	return reactions, nil
}

func (r *reactionRepository) FindByUserAndTarget(userID, postID primitive.ObjectID, commentID *primitive.ObjectID) (*domain.Reaction, error) {
	logger := utils.NewLogger("ReactionRepository.FindByUserAndTarget")
	logger.LogInput(userID, postID, commentID)

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
			logger.LogOutput(nil, nil)
			return nil, nil
		}
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(&reaction, nil)
	return &reaction, nil
}
