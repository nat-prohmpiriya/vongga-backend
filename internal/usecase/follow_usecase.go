package usecase

import (
	"errors"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type followUseCase struct {
	followRepo          domain.FollowRepository
	notificationUseCase domain.NotificationUseCase
}

// NewFollowUseCase creates a new instance of FollowUseCase
func NewFollowUseCase(fr domain.FollowRepository, nu domain.NotificationUseCase) domain.FollowUseCase {
	return &followUseCase{
		followRepo:          fr,
		notificationUseCase: nu,
	}
}

// Follow creates a new follow relationship
func (f *followUseCase) Follow(followerID, followingID primitive.ObjectID) error {
	logger := utils.NewLogger("FollowUseCase.Follow")
	input := map[string]interface{}{
		"followerID":  followerID.Hex(),
		"followingID": followingID.Hex(),
	}
	logger.LogInput(input)

	if followerID == followingID {
		err := errors.New("cannot follow yourself")
		logger.LogOutput(nil, err)
		return err
	}

	// Check if already following
	existing, err := f.followRepo.FindByFollowerAndFollowing(followerID, followingID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		logger.LogOutput(nil, err)
		return err
	}
	if existing != nil {
		if existing.Status == "blocked" {
			err := errors.New("cannot follow blocked user")
			logger.LogOutput(nil, err)
			return err
		}
		err := errors.New("already following this user")
		logger.LogOutput(nil, err)
		return err
	}

	follow := &domain.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
		Status:      "active",
	}

	err = f.followRepo.Create(follow)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Create notification for the user being followed
	_, err = f.notificationUseCase.CreateNotification(
		followingID, // recipientID (user being followed)
		followerID,  // senderID (user who followed)
		followerID,  // refID (reference to the follower)
		domain.NotificationTypeFollow,
		"user",                  // refType
		"started following you", // message
	)
	if err != nil {
		logger.LogOutput(nil, err)
		// Don't return error here as the follow action was successful
		// Just log the notification error
	}

	logger.LogOutput(follow, nil)
	return nil
}

// Unfollow removes a follow relationship
func (f *followUseCase) Unfollow(followerID, followingID primitive.ObjectID) error {
	logger := utils.NewLogger("FollowUseCase.Unfollow")
	input := map[string]interface{}{
		"followerID":  followerID.Hex(),
		"followingID": followingID.Hex(),
	}
	logger.LogInput(input)

	_, err := f.followRepo.FindByFollowerAndFollowing(followerID, followingID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			err = errors.New("not following this user")
			logger.LogOutput(nil, err)
			return err
		}
		logger.LogOutput(nil, err)
		return err
	}

	if err := f.followRepo.Delete(followerID, followingID); err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}

// Block updates the follow status to blocked
func (f *followUseCase) Block(userID, blockedID primitive.ObjectID) error {
	logger := utils.NewLogger("FollowUseCase.Block")
	input := map[string]interface{}{
		"userID":    userID.Hex(),
		"blockedID": blockedID.Hex(),
	}
	logger.LogInput(input)

	if userID == blockedID {
		err := errors.New("cannot block yourself")
		logger.LogOutput(nil, err)
		return err
	}

	// Check existing relationship
	existing, err := f.followRepo.FindByFollowerAndFollowing(blockedID, userID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		logger.LogOutput(nil, err)
		return err
	}

	if existing != nil {
		if err := f.followRepo.UpdateStatus(blockedID, userID, "blocked"); err != nil {
			logger.LogOutput(nil, err)
			return err
		}
		logger.LogOutput(nil, nil)
		return nil
	}

	// Create new blocked relationship
	follow := &domain.Follow{
		FollowerID:  blockedID,
		FollowingID: userID,
		Status:      "blocked",
	}

	if err := f.followRepo.Create(follow); err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(follow, nil)
	return nil
}

// Unblock removes the blocked status
func (f *followUseCase) Unblock(userID, blockedID primitive.ObjectID) error {
	logger := utils.NewLogger("FollowUseCase.Unblock")
	input := map[string]interface{}{
		"userID":    userID.Hex(),
		"blockedID": blockedID.Hex(),
	}
	logger.LogInput(input)

	existing, err := f.followRepo.FindByFollowerAndFollowing(blockedID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			err = errors.New("user is not blocked")
			logger.LogOutput(nil, err)
			return err
		}
		logger.LogOutput(nil, err)
		return err
	}

	if existing.Status != "blocked" {
		err = errors.New("user is not blocked")
		logger.LogOutput(nil, err)
		return err
	}

	if err := f.followRepo.Delete(blockedID, userID); err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}

// FindFollowers returns a list of followers for a user
func (f *followUseCase) FindFollowers(userID primitive.ObjectID, limit, offset int) ([]domain.Follow, error) {
	logger := utils.NewLogger("FollowUseCase.FindFollowers")
	input := map[string]interface{}{
		"userID": userID.Hex(),
		"limit":  limit,
		"offset": offset,
	}
	logger.LogInput(input)

	followers, err := f.followRepo.FindFollowers(userID, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(followers, nil)
	return followers, nil
}

// FindFollowing returns a list of users that a user is following
func (f *followUseCase) FindFollowing(userID primitive.ObjectID, limit, offset int) ([]domain.Follow, error) {
	logger := utils.NewLogger("FollowUseCase.FindFollowing")
	input := map[string]interface{}{
		"userID": userID.Hex(),
		"limit":  limit,
		"offset": offset,
	}
	logger.LogInput(input)

	following, err := f.followRepo.FindFollowing(userID, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(following, nil)
	return following, nil
}

// IsFollowing checks if a user is following another user
func (f *followUseCase) IsFollowing(followerID, followingID primitive.ObjectID) (bool, error) {
	logger := utils.NewLogger("FollowUseCase.IsFollowing")
	input := map[string]interface{}{
		"followerID":  followerID.Hex(),
		"followingID": followingID.Hex(),
	}
	logger.LogInput(input)

	follow, err := f.followRepo.FindByFollowerAndFollowing(followerID, followingID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.LogOutput(false, nil)
			return false, nil
		}
		logger.LogOutput(nil, err)
		return false, err
	}

	isFollowing := follow.Status == "active"
	logger.LogOutput(isFollowing, nil)
	return isFollowing, nil
}

// IsBlocked checks if a user is blocked by another user
func (f *followUseCase) IsBlocked(userID, blockedID primitive.ObjectID) (bool, error) {
	logger := utils.NewLogger("FollowUseCase.IsBlocked")
	input := map[string]interface{}{
		"userID":    userID.Hex(),
		"blockedID": blockedID.Hex(),
	}
	logger.LogInput(input)

	follow, err := f.followRepo.FindByFollowerAndFollowing(blockedID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.LogOutput(false, nil)
			return false, nil
		}
		logger.LogOutput(nil, err)
		return false, err
	}

	isBlocked := follow.Status == "blocked"
	logger.LogOutput(isBlocked, nil)
	return isBlocked, nil
}
