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
	"go.mongodb.org/mongo-driver/mongo/options"
)

type userRepository struct {
	collection *mongo.Collection
	rdb        *redis.Client
}

func NewUserRepository(db *mongo.Database, rdb *redis.Client) domain.UserRepository {
	return &userRepository{
		collection: db.Collection("users"),
		rdb:        rdb,
	}
}

func (r *userRepository) Create(user *domain.User) error {
	logger := utils.NewLogger("UserRepository.Create")
	logger.LogInput(user)

	// Generate a unique username
	baseUsername := utils.GenerateUsername(user.Username, user.Email)

	// Keep trying until we find a unique username
	username := baseUsername
	attempt := 1
	for {
		existingUser, err := r.FindByUsername(username)
		if err != nil {
			logger.LogOutput(nil, err)
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

	_, err := r.collection.InsertOne(context.Background(), user)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Cache the new user
	userBytes, err := json.Marshal(user)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	pipe := r.rdb.Pipeline()

	// Cache by ID
	idKey := fmt.Sprintf("user:id:%s", user.ID.Hex())
	pipe.Set(context.Background(), idKey, string(userBytes), 24*time.Hour)

	// Cache by username
	usernameKey := fmt.Sprintf("user:username:%s", user.Username)
	pipe.Set(context.Background(), usernameKey, string(userBytes), 24*time.Hour)

	// Cache by email
	emailKey := fmt.Sprintf("user:email:%s", user.Email)
	pipe.Set(context.Background(), emailKey, string(userBytes), 24*time.Hour)

	// Cache by firebase UID
	if user.FirebaseUID != "" {
		firebaseKey := fmt.Sprintf("user:firebase:%s", user.FirebaseUID)
		pipe.Set(context.Background(), firebaseKey, string(userBytes), 24*time.Hour)
	}

	_, err = pipe.Exec(context.Background())
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(user, nil)
	return nil
}

func (r *userRepository) FindByFirebaseUID(firebaseUID string) (*domain.User, error) {
	logger := utils.NewLogger("UserRepository.FindByFirebaseUID")
	logger.LogInput(firebaseUID)

	// Try to get from Redis first
	key := fmt.Sprintf("user:firebase:%s", firebaseUID)
	userJSON, err := r.rdb.Get(context.Background(), key).Result()
	if err == nil {
		// Found in Redis
		var user domain.User
		err = json.Unmarshal([]byte(userJSON), &user)
		if err != nil {
			logger.LogOutput(nil, err)
			return nil, err
		}
		logger.LogOutput(&user, nil)
		return &user, nil
	} else if err != redis.Nil {
		// Redis error
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
	var user domain.User
	err = r.collection.FindOne(context.Background(), bson.M{"firebaseUid": firebaseUID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		logger.LogOutput(nil, nil)
		return nil, nil
	}
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Cache in Redis for 24 hours
	userBytes, err := json.Marshal(user)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	err = r.rdb.Set(context.Background(), key, string(userBytes), 24*time.Hour).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.LogOutput(nil, err)
	}

	logger.LogOutput(&user, nil)
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	logger := utils.NewLogger("UserRepository.FindByEmail")
	logger.LogInput(email)

	// Try to get from Redis first
	key := fmt.Sprintf("user:email:%s", email)
	userJSON, err := r.rdb.Get(context.Background(), key).Result()
	if err == nil {
		// Found in Redis
		var user domain.User
		err = json.Unmarshal([]byte(userJSON), &user)
		if err != nil {
			logger.LogOutput(nil, err)
			return nil, err
		}
		logger.LogOutput(&user, nil)
		return &user, nil
	} else if err != redis.Nil {
		// Redis error
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
	var user domain.User
	err = r.collection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		logger.LogOutput(nil, nil)
		return nil, nil
	}
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Cache in Redis for 24 hours
	userBytes, err := json.Marshal(user)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	err = r.rdb.Set(context.Background(), key, string(userBytes), 24*time.Hour).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.LogOutput(nil, err)
	}

	logger.LogOutput(&user, nil)
	return &user, nil
}

func (r *userRepository) FindByID(id string) (*domain.User, error) {
	logger := utils.NewLogger("UserRepository.FindByID")
	logger.LogInput(id)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Try to get from Redis first
	key := fmt.Sprintf("user:id:%s", id)
	userJSON, err := r.rdb.Get(context.Background(), key).Result()
	if err == nil {
		// Found in Redis
		var user domain.User
		err = json.Unmarshal([]byte(userJSON), &user)
		if err != nil {
			logger.LogOutput(nil, err)
			return nil, err
		}
		logger.LogOutput(&user, nil)
		return &user, nil
	} else if err != redis.Nil {
		// Redis error
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
	var user domain.User
	err = r.collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		logger.LogOutput(nil, nil)
		return nil, nil
	}
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Cache in Redis for 24 hours
	userBytes, err := json.Marshal(user)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	err = r.rdb.Set(context.Background(), key, string(userBytes), 24*time.Hour).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.LogOutput(nil, err)
	}

	logger.LogOutput(&user, nil)
	return &user, nil
}

func (r *userRepository) FindByUsername(username string) (*domain.User, error) {
	logger := utils.NewLogger("UserRepository.FindByUsername")
	logger.LogInput(username)

	// Try to get from Redis first
	key := fmt.Sprintf("user:username:%s", username)
	userJSON, err := r.rdb.Get(context.Background(), key).Result()
	if err == nil {
		// Found in Redis
		var user domain.User
		err = json.Unmarshal([]byte(userJSON), &user)
		if err != nil {
			logger.LogOutput(nil, err)
			return nil, err
		}
		logger.LogOutput(&user, nil)
		return &user, nil
	} else if err != redis.Nil {
		// Redis error
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Not found in Redis, get from MongoDB
	var user domain.User
	err = r.collection.FindOne(context.Background(), bson.M{"username": username, "deletedAt": bson.M{"$exists": false}}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		logger.LogOutput(nil, nil)
		return nil, nil
	}
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Cache in Redis for 24 hours
	userBytes, err := json.Marshal(user)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	err = r.rdb.Set(context.Background(), key, string(userBytes), 24*time.Hour).Err()
	if err != nil {
		// Log Redis error but don't return it since we have the data
		logger.LogOutput(nil, err)
	}

	logger.LogOutput(&user, nil)
	return &user, nil
}

func (r *userRepository) Update(user *domain.User) error {
	logger := utils.NewLogger("UserRepository.Update")
	logger.LogInput(user)

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
		context.Background(),
		bson.M{
			"_id":     user.ID,
			"version": user.Version - 1, // Optimistic locking check
		},
		update,
	)

	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	if result.MatchedCount == 0 {
		err = fmt.Errorf("document not found or was modified by another request")
		logger.LogOutput(nil, err)
		return err
	}

	// Invalidate all user caches
	pipe := r.rdb.Pipeline()

	// Delete by ID
	idKey := fmt.Sprintf("user:id:%s", user.ID.Hex())
	pipe.Del(context.Background(), idKey)

	// Delete by username
	usernameKey := fmt.Sprintf("user:username:%s", user.Username)
	pipe.Del(context.Background(), usernameKey)

	// Delete by email
	emailKey := fmt.Sprintf("user:email:%s", user.Email)
	pipe.Del(context.Background(), emailKey)

	// Delete by firebase UID
	if user.FirebaseUID != "" {
		firebaseKey := fmt.Sprintf("user:firebase:%s", user.FirebaseUID)
		pipe.Del(context.Background(), firebaseKey)
	}

	_, err = pipe.Exec(context.Background())
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(user, nil)
	return nil
}

func (r *userRepository) GetUserByID(userID string) (*domain.User, error) {
	logger := utils.NewLogger("UserRepository.GetUserByID")
	logger.LogInput(userID)

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	var user domain.User
	err = r.collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(&user, nil)
	return &user, nil
}

func (r *userRepository) SoftDelete(id string) error {
	logger := utils.NewLogger("UserRepository.SoftDelete")
	logger.LogInput(id)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Get user first to invalidate all caches
	var user domain.User
	err = r.collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"deletedAt": time.Now(),
			"isActive":  false,
		},
	}

	result, err := r.collection.UpdateOne(context.Background(), bson.M{"_id": objectID}, update)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	if result.MatchedCount == 0 {
		err = fmt.Errorf("user not found")
		logger.LogOutput(nil, err)
		return err
	}

	// Invalidate all user caches
	pipe := r.rdb.Pipeline()

	// Delete by ID
	idKey := fmt.Sprintf("user:id:%s", user.ID.Hex())
	pipe.Del(context.Background(), idKey)

	// Delete by username
	usernameKey := fmt.Sprintf("user:username:%s", user.Username)
	pipe.Del(context.Background(), usernameKey)

	// Delete by email
	emailKey := fmt.Sprintf("user:email:%s", user.Email)
	pipe.Del(context.Background(), emailKey)

	// Delete by firebase UID
	if user.FirebaseUID != "" {
		firebaseKey := fmt.Sprintf("user:firebase:%s", user.FirebaseUID)
		pipe.Del(context.Background(), firebaseKey)
	}

	_, err = pipe.Exec(context.Background())
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(map[string]interface{}{"deleted": true}, nil)
	return nil
}

func (r *userRepository) GetUserList(req *domain.UserListRequest) ([]domain.User, int64, error) {
	logger := utils.NewLogger("UserRepository.GetUserList")

	// Try to get from Redis first
	cacheKey := fmt.Sprintf("user_list:%d:%d:%s:%s:%s:%s",
		req.Page, req.PageSize, req.Search, req.SortBy, req.SortDir, req.Status)

	var users []domain.User
	var totalCount int64

	// Try to get from cache
	cachedData, err := r.rdb.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var cachedResponse struct {
			Users      []domain.User `json:"users"`
			TotalCount int64         `json:"totalCount"`
		}
		if err := json.Unmarshal([]byte(cachedData), &cachedResponse); err == nil {
			logger.LogOutput("Retrieved user list from cache", nil)
			return cachedResponse.Users, cachedResponse.TotalCount, nil
		}
	}

	// If not in cache, query from MongoDB
	collection := r.collection
	ctx := context.Background()

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

	// Get total count
	totalCount, err = collection.CountDocuments(ctx, filter)
	if err != nil {
		logger.LogOutput("Error counting documents:", err)
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
		logger.LogOutput("Error finding users:", err)
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &users); err != nil {
		logger.LogOutput("Error decoding users:", err)
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
		r.rdb.Set(context.Background(), cacheKey, string(cacheBytes), 5*time.Minute)
	}

	return users, totalCount, nil
}
