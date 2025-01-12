package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"vongga_api/internal/domain"
	"vongga_api/utils"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/trace"
)

type storyRepository struct {
	collection *mongo.Collection
	rdb        *redis.Client
	tracer     trace.Tracer
}

func NewStoryRepository(db *mongo.Database, rdb *redis.Client, trace trace.Tracer) domain.StoryRepository {
	return &storyRepository{
		collection: db.Collection("stories"),
		rdb:        rdb,
		tracer:     trace,
	}
}

func (r *storyRepository) Create(ctx context.Context, story *domain.Story) error {
	ctx, span := r.tracer.Start(ctx, "StoryRepository.Create")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(story)

	// Set default values
	story.ID = primitive.NewObjectID()
	story.CreatedAt = time.Now()
	story.UpdatedAt = time.Now()
	story.IsActive = true
	story.Version = 1
	story.ExpiresAt = time.Now().Add(24 * time.Hour) // Stories expire after 24 hours
	story.Viewers = []domain.StoryViewer{}           // Initialize empty viewers array
	story.ViewersCount = 0

	_, err := r.collection.InsertOne(ctx, story)
	if err != nil {
		logger.Output("1", err)
		return err
	}

	// Invalidate active stories cache and user stories cache
	pipe := r.rdb.Pipeline()

	// Delete active stories cache
	pipe.Del(ctx, "active_stories")

	// Delete user stories cache
	userStoriesKey := fmt.Sprintf("user_stories:%s", story.UserID)
	pipe.Del(ctx, userStoriesKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		logger.Output("2", err)
		return err
	}

	logger.Output(story, nil)
	return nil
}

func (r *storyRepository) FindByID(ctx context.Context, id string) (*domain.Story, error) {
	ctx, span := r.tracer.Start(ctx, "StoryRepository.FindByID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(id)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.Output("1", err)
		return nil, err
	}

	// Try to get from Redis first
	key := fmt.Sprintf("story:%s", id)
	storyJSON, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		// Found in Redis
		var story domain.Story
		err = json.Unmarshal([]byte(storyJSON), &story)
		if err != nil {
			logger.Output("2", err)
			return nil, err
		}

		// Check if story is expired
		if time.Now().After(story.ExpiresAt) {
			// Delete from Redis and return nil
			r.rdb.Del(ctx, key)
			return nil, nil
		}

		logger.Output(&story, nil)
		return &story, nil
	} else if err != redis.Nil {
		// Redis error
		logger.Output("3", err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
	var story domain.Story
	err = r.collection.FindOne(ctx, bson.M{
		"_id":      objectID,
		"isActive": true,
	}).Decode(&story)

	if err == mongo.ErrNoDocuments {
		logger.Output(nil, nil)
		return nil, nil
	}
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	// Check if story is expired
	if time.Now().After(story.ExpiresAt) {
		return nil, nil
	}

	// Cache in Redis until story expires
	storyBytes, err := json.Marshal(story)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	ttl := time.Until(story.ExpiresAt)
	err = r.rdb.Set(ctx, key, string(storyBytes), ttl).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.Output(nil, err)
	}

	logger.Output(&story, nil)
	return &story, nil
}

func (r *storyRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Story, error) {
	ctx, span := r.tracer.Start(ctx, "StoryRepository.FindByUserID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(userID)

	// Try to get from Redis first
	key := fmt.Sprintf("user_stories:%s", userID)
	storiesJSON, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		// Found in Redis
		var stories []*domain.Story
		err = json.Unmarshal([]byte(storiesJSON), &stories)
		if err != nil {
			logger.Output("1", err)
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

		logger.Output(activeStories, nil)
		return activeStories, nil
	} else if err != redis.Nil {
		// Redis error
		logger.Output("2", err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
	filter := bson.M{
		"userId": userID,
		// "isActive": true,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		logger.Output("3", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var stories []*domain.Story
	if err = cursor.All(ctx, &stories); err != nil {
		logger.Output("4", err)
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
		logger.Output("5", err)
		return nil, err
	}

	err = r.rdb.Set(ctx, key, string(storiesBytes), 5*time.Minute).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.Error(err)
	}

	logger.Output(stories, nil)
	return stories, nil
}

func (r *storyRepository) FindActiveStories(ctx context.Context) ([]*domain.Story, error) {
	ctx, span := r.tracer.Start(ctx, "StoryRepository.FindActiveStories")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	// Try to get from Redis first
	key := "active_stories"
	storiesJSON, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		// Found in Redis
		var stories []*domain.Story
		err = json.Unmarshal([]byte(storiesJSON), &stories)
		if err != nil {
			logger.Output("1", err)
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

		logger.Output(activeStories, nil)
		return activeStories, nil
	} else if err != redis.Nil {
		// Redis error
		logger.Output("2", err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
	now := time.Now()
	filter := bson.M{
		"isActive": true,
		// "isArchive": false,
		"expiresAt": bson.M{"$gt": now},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		logger.Output("3", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var stories []*domain.Story
	if err = cursor.All(ctx, &stories); err != nil {
		logger.Output("4", err)
		return nil, err
	}

	// Cache in Redis for 1 minute
	storiesBytes, err := json.Marshal(stories)
	if err != nil {
		logger.Output("5", err)
		return nil, err
	}

	err = r.rdb.Set(ctx, key, string(storiesBytes), time.Minute).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.Output("6", err)
	}

	logger.Output(stories, nil)
	return stories, nil
}

func (r *storyRepository) Update(ctx context.Context, story *domain.Story) error {
	ctx, span := r.tracer.Start(ctx, "StoryRepository.Update")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(story)

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

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Output("1", err)
		return err
	}

	if result.MatchedCount == 0 {
		err = mongo.ErrNoDocuments
		logger.Output("2", err)
		return err
	}

	// Invalidate all related caches
	pipe := r.rdb.Pipeline()

	// Delete story cache
	storyKey := fmt.Sprintf("story:%s", story.ID.Hex())
	pipe.Del(ctx, storyKey)

	// Delete user stories cache
	userStoriesKey := fmt.Sprintf("user_stories:%s", story.UserID)
	pipe.Del(ctx, userStoriesKey)

	// Delete active stories cache
	pipe.Del(ctx, "active_stories")

	_, err = pipe.Exec(ctx)
	if err != nil {
		logger.Output("3", err)
		return err
	}

	logger.Output(story, nil)
	return nil
}

func (r *storyRepository) AddViewer(ctx context.Context, storyID string, viewer domain.StoryViewer) error {
	ctx, span := r.tracer.Start(ctx, "StoryRepository.AddViewer")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{"storyID": storyID, "viewer": viewer})

	objectID, err := primitive.ObjectIDFromHex(storyID)
	if err != nil {
		logger.Output("1", err)
		return err
	}

	update := bson.M{
		"$push": bson.M{"viewers": viewer},
		"$inc":  bson.M{"viewersCount": 1},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{
			"_id":      objectID,
			"isActive": true,
		},
		update,
	)

	if err != nil {
		logger.Output("2", err)
		return err
	}

	if result.MatchedCount == 0 {
		err = mongo.ErrNoDocuments
		logger.Output("3", err)
		return err
	}

	// ลบ cache
	key := fmt.Sprintf("story:%s", storyID)
	err = r.rdb.Del(ctx, key).Err()
	if err != nil {
		logger.Output("4", err)
		return err
	}

	logger.Output("AddViewer success", nil)
	return nil
}

func (r *storyRepository) DeleteStory(ctx context.Context, id string) error {
	ctx, span := r.tracer.Start(ctx, "StoryRepository.DeleteStory")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(id)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.Output("1", err)
		return err
	}

	// Find story first to get userID for cache invalidation
	var story domain.Story
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&story)
	if err != nil {
		logger.Output("2", err)
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"isActive":  false,
			"deletedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		update,
	)

	if err != nil {
		logger.Output("3", err)
		return err
	}

	if result.MatchedCount == 0 {
		err = mongo.ErrNoDocuments
		logger.Output("4", err)
		return err
	}

	// Invalidate all related caches
	pipe := r.rdb.Pipeline()

	// Delete story cache
	storyKey := fmt.Sprintf("story:%s", id)
	pipe.Del(ctx, storyKey)

	// Delete user stories cache
	userStoriesKey := fmt.Sprintf("user_stories:%s", story.UserID)
	pipe.Del(ctx, userStoriesKey)

	// Delete active stories cache
	pipe.Del(ctx, "active_stories")

	_, err = pipe.Exec(ctx)
	if err != nil {
		logger.Output("5", err)
		return err
	}

	logger.Output("DeleteStory success", nil)
	return nil
}

func (r *storyRepository) ArchiveExpiredStories(ctx context.Context) error {
	ctx, span := r.tracer.Start(ctx, "StoryRepository.ArchiveExpiredStories")
	defer span.End()
	logger := utils.NewTraceLogger(span)

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

	result, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		logger.Output("1", err)
		return err
	}

	// If any stories were archived, invalidate active stories cache
	if result.ModifiedCount > 0 {
		err = r.rdb.Del(ctx, "active_stories").Err()
		if err != nil {
			logger.Output("2", err)
			return err
		}
	}

	logger.Output(map[string]interface{}{
		"archivedCount": result.ModifiedCount,
	}, nil)
	return nil
}
