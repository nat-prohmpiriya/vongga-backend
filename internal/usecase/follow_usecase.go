package usecase

import (
	"context"
	"errors"

	"vongga_api/internal/domain"
	"vongga_api/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
)

type followUseCase struct {
	followRepo          domain.FollowRepository
	notificationUseCase domain.NotificationUseCase
	tracer              trace.Tracer
}

// NewFollowUseCase creates a new instance of FollowUseCase
func NewFollowUseCase(
	followRepo domain.FollowRepository,
	notificationUseCase domain.NotificationUseCase,
	tracer trace.Tracer,
) domain.FollowUseCase {
	return &followUseCase{
		followRepo:          followRepo,
		notificationUseCase: notificationUseCase,
		tracer:              tracer,
	}
}

// Follow creates a new follow relationship
func (f *followUseCase) Follow(ctx context.Context, followerID, followingID primitive.ObjectID) error {
	ctx, span := f.tracer.Start(ctx, "FollowUseCase.Follow")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"followerID":  followerID.Hex(),
		"followingID": followingID.Hex(),
	}
	logger.Input(input)

	if followerID == followingID {
		err := errors.New("cannot follow yourself")
		logger.Output("error following yourself 1", err)
		return err
	}

	// Check if already following
	existing, err := f.followRepo.FindByFollowerAndFollowing(ctx, followerID, followingID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		logger.Output("error checking follow status 2", err)
		return err
	}
	if existing != nil {
		if existing.Status == "blocked" {
			err := errors.New("cannot follow blocked user")
			logger.Output("error following blocked user 3", err)
			return err
		}
		err := errors.New("already following this user")
		logger.Output("already following 4", err)
		return err
	}

	follow := &domain.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
		Status:      "active",
	}

	err = f.followRepo.Create(ctx, follow)
	if err != nil {
		logger.Output("error creating follow 5", err)
		return err
	}

	// Create notification for the user being followed
	_, err = f.notificationUseCase.CreateNotification(
		ctx,
		followingID.Hex(), // recipientID (user being followed)
		followerID.Hex(),  // senderID (user who followed)
		followerID.Hex(),  // refID (reference to the follower)
		domain.NotificationTypeFollow,
		"user",                  // refType
		"started following you", // message
	)
	if err != nil {
		logger.Output("error creating notification 6", err)
		// Don't return error here as the follow action was successful
		// Just log the notification error
	}

	logger.Output(follow, nil)
	return nil
}

// Unfollow removes a follow relationship
func (f *followUseCase) Unfollow(ctx context.Context, followerID, followingID primitive.ObjectID) error {
	ctx, span := f.tracer.Start(ctx, "FollowUseCase.Unfollow")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"followerID":  followerID.Hex(),
		"followingID": followingID.Hex(),
	}
	logger.Input(input)

	_, err := f.followRepo.FindByFollowerAndFollowing(ctx, followerID, followingID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			err = errors.New("not following this user")
			logger.Output("not following 1", err)
			return err
		}
		logger.Output("error checking follow status 2", err)
		return err
	}

	if err := f.followRepo.Delete(ctx, followerID, followingID); err != nil {
		logger.Output("error deleting follow 3", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

// Block updates the follow status to blocked
func (f *followUseCase) Block(ctx context.Context, userID, blockedID primitive.ObjectID) error {
	ctx, span := f.tracer.Start(ctx, "FollowUseCase.Block")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"userID":    userID.Hex(),
		"blockedID": blockedID.Hex(),
	}
	logger.Input(input)

	if userID == blockedID {
		err := errors.New("cannot block yourself")
		logger.Output("error blocking yourself 1", err)
		return err
	}

	// Check existing relationship
	existing, err := f.followRepo.FindByFollowerAndFollowing(ctx, blockedID, userID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		logger.Output("error checking follow status 2", err)
		return err
	}

	if existing != nil {
		if err := f.followRepo.UpdateStatus(ctx, blockedID, userID, "blocked"); err != nil {
			logger.Output("error updating follow status 3", err)
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

	if err := f.followRepo.Create(ctx, follow); err != nil {
		logger.Output("error creating blocked follow 4", err)
		return err
	}

	logger.Output(follow, nil)
	return nil
}

// Unblock removes the blocked status
func (f *followUseCase) Unblock(ctx context.Context, userID, blockedID primitive.ObjectID) error {
	ctx, span := f.tracer.Start(ctx, "FollowUseCase.Unblock")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"userID":    userID.Hex(),
		"blockedID": blockedID.Hex(),
	}
	logger.Input(input)

	existing, err := f.followRepo.FindByFollowerAndFollowing(ctx, blockedID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			err = errors.New("user is not blocked")
			logger.Output("user is not blocked 1", err)
			return err
		}
		logger.Output("error checking follow status 2", err)
		return err
	}

	if existing.Status != "blocked" {
		err = errors.New("user is not blocked")
		logger.Output("user is not blocked 3", err)
		return err
	}

	if err := f.followRepo.Delete(ctx, blockedID, userID); err != nil {
		logger.Output("error deleting blocked follow 4", err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

// FindFollowers returns a list of followers for a user
func (f *followUseCase) FindFollowers(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]domain.Follow, error) {
	ctx, span := f.tracer.Start(ctx, "FollowUseCase.FindFollowers")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"userID": userID.Hex(),
		"limit":  limit,
		"offset": offset,
	}
	logger.Input(input)

	followers, err := f.followRepo.FindFollowers(ctx, userID, limit, offset)
	if err != nil {
		logger.Output("error finding followers 1", err)
		return nil, err
	}

	logger.Output(followers, nil)
	return followers, nil
}

// FindFollowing returns a list of users that a user is following
func (f *followUseCase) FindFollowing(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]domain.Follow, error) {
	ctx, span := f.tracer.Start(ctx, "FollowUseCase.FindFollowing")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"userID": userID.Hex(),
		"limit":  limit,
		"offset": offset,
	}
	logger.Input(input)

	following, err := f.followRepo.FindFollowing(ctx, userID, limit, offset)
	if err != nil {
		logger.Output("error finding following users 1", err)
		return nil, err
	}

	logger.Output(following, nil)
	return following, nil
}

// IsFollowing checks if a user is following another user
func (f *followUseCase) IsFollowing(ctx context.Context, followerID, followingID primitive.ObjectID) (bool, error) {
	ctx, span := f.tracer.Start(ctx, "FollowUseCase.IsFollowing")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"followerID":  followerID.Hex(),
		"followingID": followingID.Hex(),
	}
	logger.Input(input)

	follow, err := f.followRepo.FindByFollowerAndFollowing(ctx, followerID, followingID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.Output(false, nil)
			return false, nil
		}
		logger.Output("error checking follow status 1", err)
		return false, err
	}

	isFollowing := follow.Status == "active"
	logger.Output(isFollowing, nil)
	return isFollowing, nil
}

// IsBlocked checks if a user is blocked by another user
func (f *followUseCase) IsBlocked(ctx context.Context, userID, blockedID primitive.ObjectID) (bool, error) {
	ctx, span := f.tracer.Start(ctx, "FollowUseCase.IsBlocked")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"userID":    userID.Hex(),
		"blockedID": blockedID.Hex(),
	}
	logger.Input(input)

	follow, err := f.followRepo.FindByFollowerAndFollowing(ctx, blockedID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.Output(false, nil)
			return false, nil
		}
		logger.Output("error checking follow status 1", err)
		return false, err
	}

	isBlocked := follow.Status == "blocked"
	logger.Output(isBlocked, nil)
	return isBlocked, nil
}
