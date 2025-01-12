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
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/trace"
)

type userRepository struct {
	collection *mongo.Collection
	rdb        *redis.Client
	tracer     trace.Tracer
}

func NewUserRepository(db *mongo.Database, rdb *redis.Client, trace trace.Tracer) domain.UserRepository {
	return &userRepository{
		collection: db.Collection("users"),
		rdb:        rdb,
		tracer:     trace,
	}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	_, span := r.tracer.Start(ctx, "UserRepository.Create")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	// Generate a unique username
	baseUsername := utils.GenerateUsername(user.Username, user.Email)

	// Keep trying until we find a unique username
	username := baseUsername
	attempt := 1
	for {
		existingUser, err := r.FindByUsername(ctx, username)
		if err != nil {
			logger.Output("FindByUsername failed 1", err)
			return err
		}
		if existingUser == nil {
			break
		}
		// If username exists, try with a different number
		username = fmt.Sprintf("%s%d", baseUsername, attempt)
		attempt++
	}

	// Set default values
	user.ID = primitive.NewObjectID()
	user.Username = username
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.IsActive = true
	user.Version = 1

	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		logger.Output('1', err)
		return err
	}

	// Cache the new user
	userBytes, err := json.Marshal(user)
	if err != nil {
		logger.Output('2', err)
		return err
	}

	pipe := r.rdb.Pipeline()

	// Cache by ID
	idKey := fmt.Sprintf("user:id:%s", user.ID.Hex())
	pipe.Set(ctx, idKey, string(userBytes), 24*time.Hour)

	// Cache by username
	usernameKey := fmt.Sprintf("user:username:%s", user.Username)
	pipe.Set(ctx, usernameKey, string(userBytes), 24*time.Hour)

	// Cache by email
	emailKey := fmt.Sprintf("user:email:%s", user.Email)
	pipe.Set(ctx, emailKey, string(userBytes), 24*time.Hour)

	// Cache by firebase UID
	if user.FirebaseUID != "" {
		firebaseKey := fmt.Sprintf("user:firebase:%s", user.FirebaseUID)
		pipe.Set(ctx, firebaseKey, string(userBytes), 24*time.Hour)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		logger.Output('3', err)
		return err
	}

	logger.Output(user, nil)
	return nil
}

func (r *userRepository) FindByFirebaseUID(ctx context.Context, firebaseUID string) (*domain.User, error) {
	_, span := r.tracer.Start(ctx, "UserRepository.FindByFirebaseUID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(firebaseUID)
	// Try to get from Redis first
	key := fmt.Sprintf("user:firebase:%s", firebaseUID)
	userJSON, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		// Found in Redis
		var user domain.User
		err = json.Unmarshal([]byte(userJSON), &user)
		if err != nil {
			logger.Output('1', err)

			return nil, err
		}
		logger.Output(&user, nil)
		return &user, nil
	} else if err != redis.Nil {
		// Redis error
		logger.Output('3', err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
	var user domain.User
	err = r.collection.FindOne(ctx, bson.M{"firebaseUid": firebaseUID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		logger.Output('4', err)
		return nil, nil
	}
	if err != nil {
		logger.Output('5', err)
		return nil, err
	}

	// Cache in Redis for 24 hours
	userBytes, err := json.Marshal(user)
	if err != nil {
		logger.Output('6', err)
		return nil, err
	}

	err = r.rdb.Set(ctx, key, string(userBytes), 24*time.Hour).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.Output('7', err)
	}

	logger.Output(&user, nil)
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.FindByEmail")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(email)

	// Try to get from Redis first
	key := fmt.Sprintf("user:email:%s", email)
	userJSON, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		// Found in Redis
		var user domain.User
		err = json.Unmarshal([]byte(userJSON), &user)
		if err != nil {
			logger.Output('8', err)
			return nil, err
		}
		logger.Output(&user, nil)
		return &user, nil
	} else if err != redis.Nil {
		// Redis error
		logger.Output('9', err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
	var user domain.User
	err = r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		logger.Output(map[string]interface{}{"message": "mongo not found", "error": err}, nil)
		return nil, nil
	}
	if err != nil {
		logger.Output("11", err)
		return nil, err
	}

	// Cache in Redis for 24 hours
	userBytes, err := json.Marshal(user)
	if err != nil {
		logger.Output(map[string]interface{}{"message": "cannot marshal", "error": err}, nil)
	}

	err = r.rdb.Set(ctx, key, string(userBytes), 24*time.Hour).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.Error(err)
	}

	logger.Output(&user, nil)
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	_, span := r.tracer.Start(ctx, "UserRepository.FindByID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(id)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.Output("1", err)
		return nil, err
	}

	// Try to get from Redis first
	key := fmt.Sprintf("user:id:%s", id)
	userJSON, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		// Found in Redis
		var user domain.User
		err = json.Unmarshal([]byte(userJSON), &user)
		if err != nil {
			logger.Output("redis unmarshal 2", err)
			return nil, err
		}
		logger.Output(&user, nil)
		return &user, nil
	} else if err != redis.Nil {
		// Redis error
		logger.Output("redis error 3", err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
	var user domain.User
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		logger.Output(map[string]interface{}{"message": "mongo not found", "error": err}, nil)
		return nil, nil
	}
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// Cache in Redis for 24 hours
	userBytes, err := json.Marshal(user)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	err = r.rdb.Set(ctx, key, string(userBytes), 24*time.Hour).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.Error(err)
	}

	logger.Output(&user, nil)
	return &user, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	_, span := r.tracer.Start(ctx, "UserRepository.FindByUsername")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(username)

	// Try to get from Redis first
	key := fmt.Sprintf("user:username:%s", username)
	userJSON, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		// Found in Redis
		var user domain.User
		err = json.Unmarshal([]byte(userJSON), &user)
		if err != nil {
			logger.Output("redis unmarshal 2", err)
			return nil, err
		}
		logger.Output(&user, nil)
		return &user, nil
	} else if err != redis.Nil {
		// Redis error
		logger.Output("redis error 3", err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
	var user domain.User
	err = r.collection.FindOne(ctx, bson.M{"username": username, "deletedAt": bson.M{"$exists": false}}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		logger.Output(map[string]interface{}{"message": "mongo not found", "error": err}, nil)
		return nil, nil
	}
	if err != nil {
		logger.Output("mongo find 4", err)
		return nil, err
	}

	// Cache in Redis for 24 hours
	userBytes, err := json.Marshal(user)
	if err != nil {
		logger.Output("redis marshal 5", err)
		return nil, err
	}

	err = r.rdb.Set(ctx, key, string(userBytes), 24*time.Hour).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.Warn(map[string]interface{}{"message": "redis error 6", "error": err})
	}

	logger.Output(&user, nil)
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	_, span := r.tracer.Start(ctx, "UserRepository.Update")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(user)

	update := bson.M{
		"$set": bson.M{
			"username":       user.Username,
			"email":          user.Email,
			"firstName":      user.FirstName,
			"lastName":       user.LastName,
			"displayName":    user.DisplayName,
			"bio":            user.Bio,
			"avatar":         user.Avatar,
			"photoProfile":   user.PhotoProfile,
			"photoCover":     user.PhotoCover,
			"dateOfBirth":    user.DateOfBirth,
			"gender":         user.Gender,
			"interestedIn":   user.InterestedIn,
			"location":       user.Location,
			"relationStatus": user.RelationStatus,
			"height":         user.Height,
			"interests":      user.Interests,
			"occupation":     user.Occupation,
			"education":      user.Education,
			"phoneNumber":    user.PhoneNumber,
			"datingPhotos":   user.DatingPhotos,
			"isVerified":     user.IsVerified,
			"isActive":       user.IsActive,
			"live":           user.Live,
			"updatedAt":      user.UpdatedAt,
			"version":        user.Version,
		},
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{
			"_id":     user.ID,
			"version": user.Version - 1, // Optimistic locking check
		},
		update,
	)

	if err != nil {
		logger.Output("Failed to update user 1", err)
		return err
	}

	if result.MatchedCount == 0 {
		err = fmt.Errorf("document not found or was modified by another request")
		logger.Output(map[string]interface{}{"message": "document not found or was modified by another request", "error": err}, nil)
		return err
	}

	// Invalidate all user caches
	pipe := r.rdb.Pipeline()

	// Delete by ID
	idKey := fmt.Sprintf("user:id:%s", user.ID.Hex())
	pipe.Del(ctx, idKey)

	// Delete by username
	usernameKey := fmt.Sprintf("user:username:%s", user.Username)
	pipe.Del(ctx, usernameKey)

	// Delete by email
	emailKey := fmt.Sprintf("user:email:%s", user.Email)
	pipe.Del(ctx, emailKey)

	// Delete by firebase UID
	if user.FirebaseUID != "" {
		firebaseKey := fmt.Sprintf("user:firebase:%s", user.FirebaseUID)
		pipe.Del(ctx, firebaseKey)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		logger.Output("Failed to update user 2", err)
		return err
	}

	logger.Output(user, nil)
	return nil
}

func (r *userRepository) FindUserByID(ctx context.Context, userID string) (*domain.User, error) {
	_, span := r.tracer.Start(ctx, "UserRepository.FindUserByID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(userID)

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logger.Output("1", err)
		return nil, err
	}

	var user domain.User
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		logger.Output("2", err)
		return nil, err
	}

	logger.Output(user, nil)
	return &user, nil
}

func (r *userRepository) SoftDelete(ctx context.Context, id string) error {
	ctx, span := r.tracer.Start(ctx, "UserRepository.SoftDelete")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(id)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.Output("1", err)
		return err
	}

	// Find user first to invalidate all caches
	var user domain.User
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		logger.Output("2", err)
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"deletedAt": time.Now(),
			"isActive":  false,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		logger.Output("3", err)
		return err
	}

	if result.MatchedCount == 0 {
		err = fmt.Errorf("user not found")
		logger.Output("4", err)
		return err
	}

	// Invalidate all user caches
	pipe := r.rdb.Pipeline()

	// Delete by ID
	idKey := fmt.Sprintf("user:id:%s", user.ID.Hex())
	pipe.Del(ctx, idKey)

	// Delete by username
	usernameKey := fmt.Sprintf("user:username:%s", user.Username)
	pipe.Del(ctx, usernameKey)

	// Delete by email
	emailKey := fmt.Sprintf("user:email:%s", user.Email)
	pipe.Del(ctx, emailKey)

	// Delete by firebase UID
	if user.FirebaseUID != "" {
		firebaseKey := fmt.Sprintf("user:firebase:%s", user.FirebaseUID)
		pipe.Del(ctx, firebaseKey)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		logger.Output("5", err)
		return err
	}

	logger.Output(map[string]interface{}{"deleted": true}, nil)
	return nil
}

func (r *userRepository) FindUserFindMany(ctx context.Context, req *domain.UserFindManyRequest) ([]domain.User, int64, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.FindUserFindMany")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(req)

	// Try to get from Redis first
	cacheKey := fmt.Sprintf("user_list:%d:%d:%s:%s:%s:%s",
		req.Page, req.PageSize, req.Search, req.SortBy, req.SortDir, req.Status)

	var users []domain.User
	var totalCount int64

	// Try to get from cache
	cachedData, err := r.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedResponse struct {
			Users      []domain.User `json:"users"`
			TotalCount int64         `json:"totalCount"`
		}
		if err := json.Unmarshal([]byte(cachedData), &cachedResponse); err == nil {
			logger.Output("Retrieved user list from cache", nil)
			return cachedResponse.Users, cachedResponse.TotalCount, nil
		}
	}

	// If not in cache, query from MongoDB
	collection := r.collection

	// Build filter
	filter := bson.M{"deletedAt": nil}
	if req.Status != "" {
		filter["status"] = req.Status
	}
	if req.Search != "" {
		filter["$or"] = []bson.M{
			{"firstName": bson.M{"$regex": req.Search, "$options": "i"}},
			{"lastName": bson.M{"$regex": req.Search, "$options": "i"}},
			{"username": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}

	// Find total count
	totalCount, err = collection.CountDocuments(ctx, filter)
	if err != nil {
		logger.Output(map[string]interface{}{"totalCount": 0, "err": err}, nil)
		return nil, 0, err
	}

	// Build sort
	sort := bson.D{{Key: "createdAt", Value: -1}} // default sort
	if req.SortBy != "" {
		sortDir := 1
		if req.SortDir == "desc" {
			sortDir = -1
		}
		sort = bson.D{{Key: req.SortBy, Value: sortDir}}
	}

	// Calculate skip
	skip := (req.Page - 1) * req.PageSize

	// Execute query
	opts := options.Find().
		SetSort(sort).
		SetSkip(int64(skip)).
		SetLimit(int64(req.PageSize))

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		logger.Output("Error finding users", err)
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &users); err != nil {
		logger.Output("Error decoding users", err)
		return nil, 0, err
	}

	// Cache the results
	cacheData := struct {
		Users      []domain.User `json:"users"`
		TotalCount int64         `json:"totalCount"`
	}{
		Users:      users,
		TotalCount: totalCount,
	}

	if cacheBytes, err := json.Marshal(cacheData); err == nil {
		// Cache for 5 minutes
		r.rdb.Set(ctx, cacheKey, string(cacheBytes), 5*time.Minute)
	}
	logger.Output(map[string]interface{}{"totalCount": totalCount, "users": users}, nil)
	return users, totalCount, nil
}
