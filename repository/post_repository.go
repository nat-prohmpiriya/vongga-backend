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

	now := time.Now()
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"deletedAt": now}}

	_, err := r.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput("Post soft deleted successfully", nil)
	return nil
}

func (r *postRepository) FindByID(id primitive.ObjectID) (*domain.Post, error) {
	logger := utils.NewLogger("PostRepository.FindByID")
	logger.LogInput(id)

	filter := bson.M{
		"_id": id,
		"deletedAt": bson.M{"$exists": false},
	}

	var post domain.Post
	err := r.collection.FindOne(context.Background(), filter).Decode(&post)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			notFoundErr := domain.NewNotFoundError("post", id.Hex())
			logger.LogOutput(nil, notFoundErr)
			return nil, notFoundErr
		}
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(post, nil)
	return &post, nil
}

func (r *postRepository) FindByUserID(userID primitive.ObjectID, limit, offset int) ([]domain.Post, error) {
	logger := utils.NewLogger("PostRepository.FindByUserID")
	logger.LogInput(map[string]interface{}{
		"userID": userID,
		"limit":  limit,
		"offset": offset,
	})

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	filter := bson.M{
		"userId": userID,
		"deletedAt": bson.M{"$exists": false},
	}

	cursor, err := r.collection.Find(context.Background(), filter, opts)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var posts []domain.Post
	if err = cursor.All(context.Background(), &posts); err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(posts, nil)
	return posts, nil
}
