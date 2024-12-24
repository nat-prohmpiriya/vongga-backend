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

type postRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewPostRepository(db *mongo.Database) domain.PostRepository {
	return &postRepository{
		db:         db,
		collection: db.Collection("posts"),
	}
}

func (r *postRepository) Create(post *domain.Post) error {
	logger := utils.NewLogger("PostRepository.Create")
	logger.LogInput(post)

	_, err := r.collection.InsertOne(context.Background(), post)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput("Post created successfully", nil)
	return nil
}

func (r *postRepository) Update(post *domain.Post) error {
	logger := utils.NewLogger("PostRepository.Update")
	logger.LogInput(post)

	filter := bson.M{"_id": post.ID}
	update := bson.M{"$set": post}
	_, err := r.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput("Post updated successfully", nil)
	return nil
}

func (r *postRepository) Delete(id primitive.ObjectID) error {
	logger := utils.NewLogger("PostRepository.Delete")
	logger.LogInput(id)

	filter := bson.M{"_id": id}
	_, err := r.collection.DeleteOne(context.Background(), filter)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput("Post deleted successfully", nil)
	return nil
}

func (r *postRepository) FindByID(id primitive.ObjectID) (*domain.Post, error) {
	logger := utils.NewLogger("PostRepository.FindByID")
	logger.LogInput(id)

	var post domain.Post
	filter := bson.M{"_id": id}
	err := r.collection.FindOne(context.Background(), filter).Decode(&post)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(post, nil)
	return &post, nil
}

func (r *postRepository) FindByUserID(userID primitive.ObjectID, limit, offset int) ([]domain.Post, error) {
	logger := utils.NewLogger("PostRepository.FindByUserID")
	input := map[string]interface{}{
		"userID": userID,
		"limit":  limit,
		"offset": offset,
	}
	logger.LogInput(input)

	var posts []domain.Post
	filter := bson.M{"userId": userID}
	
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

	err = cursor.All(context.Background(), &posts)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(posts, nil)
	return posts, nil
}
