package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) domain.UserRepository {
	return &userRepository{
		collection: db.Collection("users"),
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

	logger.LogOutput(user, nil)
	return nil
}

func (r *userRepository) FindByFirebaseUID(firebaseUID string) (*domain.User, error) {
	logger := utils.NewLogger("UserRepository.FindByFirebaseUID")
	logger.LogInput(firebaseUID)

	var user domain.User
	err := r.collection.FindOne(context.Background(), bson.M{"firebase_uid": firebaseUID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		logger.LogOutput(nil, nil)
		return nil, nil
	}
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(&user, nil)
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	logger := utils.NewLogger("UserRepository.FindByEmail")
	logger.LogInput(email)

	var user domain.User
	err := r.collection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		logger.LogOutput(nil, nil)
		return nil, nil
	}
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
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

	logger.LogOutput(&user, nil)
	return &user, nil
}

func (r *userRepository) FindByUsername(username string) (*domain.User, error) {
	logger := utils.NewLogger("UserRepository.FindByUsername")
	logger.LogInput(username)

	var user domain.User
	err := r.collection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		logger.LogOutput(nil, nil)
		return nil, nil
	}
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(&user, nil)
	return &user, nil
}

func (r *userRepository) Update(user *domain.User) error {
	logger := utils.NewLogger("UserRepository.Update")
	logger.LogInput(user)

	update := bson.M{
		"$set": bson.M{
			"username":         user.Username,
			"email":           user.Email,
			"first_name":      user.FirstName,
			"last_name":       user.LastName,
			"display_name":    user.DisplayName,
			"bio":             user.Bio,
			"avatar":          user.Avatar,
			"photo_profile":   user.PhotoProfile,
			"photo_cover":     user.PhotoCover,
			"date_of_birth":   user.DateOfBirth,
			"gender":          user.Gender,
			"interested_in":   user.InterestedIn,
			"location":        user.Location,
			"relation_status": user.RelationStatus,
			"height":          user.Height,
			"interests":       user.Interests,
			"occupation":      user.Occupation,
			"education":       user.Education,
			"phone_number":    user.PhoneNumber,
			"dating_photos":   user.DatingPhotos,
			"is_verified":     user.IsVerified,
			"is_active":       user.IsActive,
			"live":            user.Live,
			"updated_at":      user.UpdatedAt,
			"version":         user.Version,
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

	logger.LogOutput(user, nil)
	return nil
}

func (r *userRepository) SoftDelete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"deletedAt": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}
