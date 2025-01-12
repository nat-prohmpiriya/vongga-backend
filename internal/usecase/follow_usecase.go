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
	logger := utils.NewTraceLogger("FollowUseCase.Follow")
	input := map[string]interface{}{
		"followerID":  followerID.Hex(),
		"followingID": followingID.Hex(),
	}
	logger.Input(input)

	if followerID == followingID {
		err := errors.New("cannot follow yourself")
		logger.Output(nil, err)
		return err
	}

	// Check if already following
	existing, err := f.followRepo.FindByFollowerAndFollowing(followerID, followingID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		logger.Output(nil, err)
		return err
	}
	if existing != nil {
		if existing.Status == "blocked" {
			err := errors.New("cannot follow blocked user")
			logger.Output(nil, err)
			return err
		}
		err := errors.New("already following this user")
		logger.Output(nil, err)
		return err
	}

	follow := &domain.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
		Status:      "active",
	}

	err = f.followRepo.Create(follow)
	if err != nil {
		logger.Output(nil, err)
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
		logger.Output(nil, err)
		// Don't return error here as the follow action was successful
		// Just log the notification error
	}

	logger.Output(follow, nil)
	return nil
}

// Unfollow removes a follow relationship
func (f *followUseCase) Unfollow(followerID, followingID primitive.ObjectID) error {
	logger := utils.NewTraceLogger("FollowUseCase.Unfollow")
	input := map[string]interface{}{
		"followerID":  followerID.Hex(),
		"followingID": followingID.Hex(),
	}
	logger.Input(input)

	_, err := f.followRepo.FindByFollowerAndFollowing(followerID, followingID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			err = errors.New("not following this user")
			logger.Output(nil, err)
			return err
		}
		logger.Output(nil, err)
		return err
	}

	if err := f.followRepo.Delete(followerID, followingID); err != nil {
		logger.Output(nil, err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

// Block updates the follow status to blocked
func (f *followUseCase) Block(userID, blockedID primitive.ObjectID) error {
	logger := utils.NewTraceLogger("FollowUseCase.Block")
	input := map[string]interface{}{
		"userID":    userID.Hex(),
		"blockedID": blockedID.Hex(),
	}
	logger.Input(input)

	if userID == blockedID {
		err := errors.New("cannot block yourself")
		logger.Output(nil, err)
		return err
	}

	// Check existing relationship
	existing, err := f.followRepo.FindByFollowerAndFollowing(blockedID, userID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		logger.Output(nil, err)
		return err
	}

	if existing != nil {
		if err := f.followRepo.UpdateStatus(blockedID, userID, "blocked"); err != nil {
			logger.Output(nil, err)
			return err
		}
		logger.Output(nil, nil)
		return nil
	}

	// Create new blocked relationship
	follow := &domain.Follow{
		FollowerID:  blockedID,
		FollowingID: userID,
		Status:      "blocked",
	}

	if err := f.followRepo.Create(follow); err != nil {
		logger.Output(nil, err)
		return err
	}

	logger.Output(follow, nil)
	return nil
}

// Unblock removes the blocked status
func (f *followUseCase) Unblock(userID, blockedID primitive.ObjectID) error {
	logger := utils.NewTraceLogger("FollowUseCase.Unblock")
	input := map[string]interface{}{
		"userID":    userID.Hex(),
		"blockedID": blockedID.Hex(),
	}
	logger.Input(input)

	existing, err := f.followRepo.FindByFollowerAndFollowing(blockedID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			err = errors.New("user is not blocked")
			logger.Output(nil, err)
			return err
		}
		logger.Output(nil, err)
		return err
	}

	if existing.Status != "blocked" {
		err = errors.New("user is not blocked")
		logger.Output(nil, err)
		return err
	}

	if err := f.followRepo.Delete(blockedID, userID); err != nil {
		logger.Output(nil, err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

// FindFollowers returns a list of followers for a user
func (f *followUseCase) FindFollowers(userID primitive.ObjectID, limit, offset int) ([]domain.Follow, error) {
	logger := utils.NewTraceLogger("FollowUseCase.FindFollowers")
	input := map[string]interface{}{
		"userID": userID.Hex(),
		"limit":  limit,
		"offset": offset,
	}
	logger.Input(input)

	followers, err := f.followRepo.FindFollowers(userID, limit, offset)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(followers, nil)
	return followers, nil
}

// FindFollowing returns a list of users that a user is following
func (f *followUseCase) FindFollowing(userID primitive.ObjectID, limit, offset int) ([]domain.Follow, error) {
	logger := utils.NewTraceLogger("FollowUseCase.FindFollowing")
	input := map[string]interface{}{
		"userID": userID.Hex(),
		"limit":  limit,
		"offset": offset,
	}
	logger.Input(input)

	following, err := f.followRepo.FindFollowing(userID, limit, offset)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(following, nil)
	return following, nil
}

// IsFollowing checks if a user is following another user
func (f *followUseCase) IsFollowing(followerID, followingID primitive.ObjectID) (bool, error) {
	logger := utils.NewTraceLogger("FollowUseCase.IsFollowing")
	input := map[string]interface{}{
		"followerID":  followerID.Hex(),
		"followingID": followingID.Hex(),
	}
	logger.Input(input)

	follow, err := f.followRepo.FindByFollowerAndFollowing(followerID, followingID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.Output(false, nil)
			return false, nil
		}
		logger.Output(nil, err)
		return false, err
	}

	isFollowing := follow.Status == "active"
	logger.Output(isFollowing, nil)
	return isFollowing, nil
}

// IsBlocked checks if a user is blocked by another user
func (f *followUseCase) IsBlocked(userID, blockedID primitive.ObjectID) (bool, error) {
	logger := utils.NewTraceLogger("FollowUseCase.IsBlocked")
	input := map[string]interface{}{
		"userID":    userID.Hex(),
		"blockedID": blockedID.Hex(),
	}
	logger.Input(input)

	follow, err := f.followRepo.FindByFollowerAndFollowing(blockedID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.Output(false, nil)
			return false, nil
		}
		logger.Output(nil, err)
		return false, err
	}

	isBlocked := follow.Status == "blocked"
	logger.Output(isBlocked, nil)
	return isBlocked, nil
}
