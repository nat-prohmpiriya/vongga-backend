package repository

import (
	"context"

	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type commentRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewCommentRepository(db *mongo.Database) domain.CommentRepository {
	return &commentRepository{
		db:         db,
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

	logger.LogOutput("Comment updated successfully", nil)
	return nil
}

func (r *commentRepository) Delete(id primitive.ObjectID) error {
	logger := utils.NewLogger("CommentRepository.Delete")
	logger.LogInput(id)

	filter := bson.M{"_id": id}
	_, err := r.collection.DeleteOne(context.Background(), filter)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput("Comment deleted successfully", nil)
	return nil
}

func (r *commentRepository) FindByID(id primitive.ObjectID) (*domain.Comment, error) {
	logger := utils.NewLogger("CommentRepository.FindByID")
	logger.LogInput(id)

	var comment domain.Comment
	filter := bson.M{"_id": id}
	err := r.collection.FindOne(context.Background(), filter).Decode(&comment)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(comment, nil)
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

	cursor, err := r.collection.Find(context.Background(), filter, findOptions)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	err = cursor.All(context.Background(), &comments)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
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

	logger.LogOutput(map[string]interface{}{
		"message":       "Comments deleted successfully",
		"deletedCount": result.DeletedCount,
	}, nil)
	return nil
}
