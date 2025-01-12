package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type storyRepository struct {
	collection *mongo.Collection
	rdb        *redis.Client
}

func NewStoryRepository(db *mongo.Database, rdb *redis.Client) domain.StoryRepository {
	return &storyRepository{
		collection: db.Collection("stories"),
		rdb:        rdb,
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
	story.Viewers = []domain.StoryViewer{}           // Initialize empty viewers array
	story.ViewersCount = 0

	_, err := r.collection.InsertOne(context.Background(), story)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Invalidate active stories cache and user stories cache
	pipe := r.rdb.Pipeline()

	// Delete active stories cache
	pipe.Del(context.Background(), "active_stories")

	// Delete user stories cache
	userStoriesKey := fmt.Sprintf("user_stories:%s", story.UserID)
	pipe.Del(context.Background(), userStoriesKey)

	_, err = pipe.Exec(context.Background())
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

	// Try to get from Redis first
	key := fmt.Sprintf("story:%s", id)
	storyJSON, err := r.rdb.Get(context.Background(), key).Result()
	if err == nil {
		// Found in Redis
		var story domain.Story
		err = json.Unmarshal([]byte(storyJSON), &story)
		if err != nil {
			logger.LogOutput(nil, err)
			return nil, err
		}

		// Check if story is expired
		if time.Now().After(story.ExpiresAt) {
			// Delete from Redis and return nil
			r.rdb.Del(context.Background(), key)
			return nil, nil
		}

		logger.LogOutput(&story, nil)
		return &story, nil
	} else if err != redis.Nil {
		// Redis error
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
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

	// Check if story is expired
	if time.Now().After(story.ExpiresAt) {
		return nil, nil
	}

	// Cache in Redis until story expires
	storyBytes, err := json.Marshal(story)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	ttl := time.Until(story.ExpiresAt)
	err = r.rdb.Set(context.Background(), key, string(storyBytes), ttl).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.LogOutput(nil, err)
	}

	logger.LogOutput(&story, nil)
	return &story, nil
}

func (r *storyRepository) FindByUserID(userID string) ([]*domain.Story, error) {
	logger := utils.NewLogger("StoryRepository.FindByUserID")
	logger.LogInput(userID)

	// Try to get from Redis first
	key := fmt.Sprintf("user_stories:%s", userID)
	storiesJSON, err := r.rdb.Get(context.Background(), key).Result()
	if err == nil {
		// Found in Redis
		var stories []*domain.Story
		err = json.Unmarshal([]byte(storiesJSON), &stories)
		if err != nil {
			logger.LogOutput(nil, err)
			return nil, err
		}

		// Filter out expired stories
		now := time.Now()
		activeStories := make([]*domain.Story, 0)
		for _, story := range stories {
			if now.Before(story.ExpiresAt) {
				activeStories = append(activeStories, story)
			}
		}

		logger.LogOutput(activeStories, nil)
		return activeStories, nil
	} else if err != redis.Nil {
		// Redis error
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
	filter := bson.M{
		"userId": userID,
		// "isActive": true,
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

	// Filter out expired stories not filter becouse use in page site
	// now := time.Now()
	// activeStories := make([]*domain.Story, 0)
	// for _, story := range stories {
	// 	if now.Before(story.ExpiresAt) {
	// 		activeStories = append(activeStories, story)
	// 	}
	// }

	// Cache in Redis for 5 minutes
	storiesBytes, err := json.Marshal(stories)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	err = r.rdb.Set(context.Background(), key, string(storiesBytes), 5*time.Minute).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.LogOutput(nil, err)
	}

	logger.LogOutput(stories, nil)
	return stories, nil
}

func (r *storyRepository) FindActiveStories() ([]*domain.Story, error) {
	logger := utils.NewLogger("StoryRepository.FindActiveStories")

	// Try to get from Redis first
	key := "active_stories"
	storiesJSON, err := r.rdb.Get(context.Background(), key).Result()
	if err == nil {
		// Found in Redis
		var stories []*domain.Story
		err = json.Unmarshal([]byte(storiesJSON), &stories)
		if err != nil {
			logger.LogOutput(nil, err)
			return nil, err
		}

		// Filter out expired stories
		now := time.Now()
		activeStories := make([]*domain.Story, 0)
		for _, story := range stories {
			if now.Before(story.ExpiresAt) {
				activeStories = append(activeStories, story)
			}
		}

		logger.LogOutput(activeStories, nil)
		return activeStories, nil
	} else if err != redis.Nil {
		// Redis error
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
	now := time.Now()
	filter := bson.M{
		"isActive": true,
		// "isArchive": false,
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

	// Cache in Redis for 1 minute
	storiesBytes, err := json.Marshal(stories)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	err = r.rdb.Set(context.Background(), key, string(storiesBytes), time.Minute).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.LogOutput(nil, err)
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

	// Invalidate all related caches
	pipe := r.rdb.Pipeline()

	// Delete story cache
	storyKey := fmt.Sprintf("story:%s", story.ID.Hex())
	pipe.Del(context.Background(), storyKey)

	// Delete user stories cache
	userStoriesKey := fmt.Sprintf("user_stories:%s", story.UserID)
	pipe.Del(context.Background(), userStoriesKey)

	// Delete active stories cache
	pipe.Del(context.Background(), "active_stories")

	_, err = pipe.Exec(context.Background())
	if err != nil {
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
		bson.M{
			"_id":      objectID,
			"isActive": true,
		},
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

	// ลบ cache
	key := fmt.Sprintf("story:%s", storyID)
	err = r.rdb.Del(context.Background(), key).Err()
	if err != nil {
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

	// Find story first to get userID for cache invalidation
	var story domain.Story
	err = r.collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&story)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"isActive":  false,
			"deletedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
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

	// Invalidate all related caches
	pipe := r.rdb.Pipeline()

	// Delete story cache
	storyKey := fmt.Sprintf("story:%s", id)
	pipe.Del(context.Background(), storyKey)

	// Delete user stories cache
	userStoriesKey := fmt.Sprintf("user_stories:%s", story.UserID)
	pipe.Del(context.Background(), userStoriesKey)

	// Delete active stories cache
	pipe.Del(context.Background(), "active_stories")

	_, err = pipe.Exec(context.Background())
	if err != nil {
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
			"isArchive": true,
			"updatedAt": now,
		},
	}

	result, err := r.collection.UpdateMany(context.Background(), filter, update)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// If any stories were archived, invalidate active stories cache
	if result.ModifiedCount > 0 {
		err = r.rdb.Del(context.Background(), "active_stories").Err()
		if err != nil {
			logger.LogOutput(nil, err)
			return err
		}
	}

	logger.LogOutput(map[string]interface{}{
		"archivedCount": result.ModifiedCount,
	}, nil)
	return nil
}
