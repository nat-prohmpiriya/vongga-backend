package repository

import (
	"context"
	"time"

	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type storyRepository struct {
	collection *mongo.Collection
}

func NewStoryRepository(db *mongo.Database) domain.StoryRepository {
	return &storyRepository{
		collection: db.Collection("stories"),
	}
}

func (r *storyRepository) Create(story *domain.Story) error {
	logger := utils.NewLogger("StoryRepository.Create")
	logger.LogInput(story)

	// Set default values
	story.ID = primitive.NewObjectID()
	story.CreatedAt = time.Now()
	story.UpdatedAt = time.Now()
	story.IsActive = true
	story.Version = 1
	story.ExpiresAt = time.Now().Add(24 * time.Hour) // Stories expire after 24 hours

	_, err := r.collection.InsertOne(context.Background(), story)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(story, nil)
	return nil
}

func (r *storyRepository) FindByID(id string) (*domain.Story, error) {
	logger := utils.NewLogger("StoryRepository.FindByID")
	logger.LogInput(id)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	var story domain.Story
	err = r.collection.FindOne(context.Background(), bson.M{
		"_id":      objectID,
		"isActive": true,
	}).Decode(&story)

	if err == mongo.ErrNoDocuments {
		logger.LogOutput(nil, nil)
		return nil, nil
	}
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(&story, nil)
	return &story, nil
}

func (r *storyRepository) FindByUserID(userID string) ([]*domain.Story, error) {
	logger := utils.NewLogger("StoryRepository.FindByUserID")
	logger.LogInput(userID)

	filter := bson.M{
		"userId":   userID,
		"isActive": true,
	}

	cursor, err := r.collection.Find(context.Background(), filter)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var stories []*domain.Story
	if err = cursor.All(context.Background(), &stories); err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(stories, nil)
	return stories, nil
}

func (r *storyRepository) FindActiveStories() ([]*domain.Story, error) {
	logger := utils.NewLogger("StoryRepository.FindActiveStories")

	now := time.Now()
	filter := bson.M{
		"isActive":  true,
		"isArchive": false,
		"expiresAt": bson.M{"$gt": now},
	}

	cursor, err := r.collection.Find(context.Background(), filter)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var stories []*domain.Story
	if err = cursor.All(context.Background(), &stories); err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(stories, nil)
	return stories, nil
}

func (r *storyRepository) Update(story *domain.Story) error {
	logger := utils.NewLogger("StoryRepository.Update")
	logger.LogInput(story)

	story.UpdatedAt = time.Now()
	story.Version++

	filter := bson.M{
		"_id":      story.ID,
		"version":  story.Version - 1,
		"isActive": true,
	}

	update := bson.M{
		"$set": story,
	}

	result, err := r.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	if result.MatchedCount == 0 {
		err = mongo.ErrNoDocuments
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(story, nil)
	return nil
}

func (r *storyRepository) AddViewer(storyID string, viewer domain.StoryViewer) error {
	logger := utils.NewLogger("StoryRepository.AddViewer")
	logger.LogInput(map[string]interface{}{"storyID": storyID, "viewer": viewer})

	objectID, err := primitive.ObjectIDFromHex(storyID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	update := bson.M{
		"$push": bson.M{"viewers": viewer},
		"$inc":  bson.M{"viewersCount": 1},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	result, err := r.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID, "isActive": true},
		update,
	)

	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	if result.MatchedCount == 0 {
		err = mongo.ErrNoDocuments
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}

func (r *storyRepository) DeleteStory(id string) error {
	logger := utils.NewLogger("StoryRepository.DeleteStory")
	logger.LogInput(id)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"isActive":   false,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID, "isActive": true},
		update,
	)

	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	if result.MatchedCount == 0 {
		err = mongo.ErrNoDocuments
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}

func (r *storyRepository) ArchiveExpiredStories() error {
	logger := utils.NewLogger("StoryRepository.ArchiveExpiredStories")

	now := time.Now()
	filter := bson.M{
		"isActive":  true,
		"isArchive": false,
		"expiresAt": bson.M{"$lte": now},
	}

	update := bson.M{
		"$set": bson.M{
			"isArchive":  true,
			"updatedAt": time.Now(),
		},
	}

	_, err := r.collection.UpdateMany(context.Background(), filter, update)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}
