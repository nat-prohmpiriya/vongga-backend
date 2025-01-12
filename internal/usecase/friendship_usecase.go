package usecase

import (
	"context"
	"errors"
	"time"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
)

type friendshipUseCase struct {
	friendshipRepo      domain.FriendshipRepository
	notificationUseCase domain.NotificationUseCase
	tracer              trace.Tracer
}

func NewFriendshipUseCase(
	fr domain.FriendshipRepository,
	nu domain.NotificationUseCase,
	tracer trace.Tracer,
) domain.FriendshipUseCase {
	return &friendshipUseCase{
		friendshipRepo:      fr,
		notificationUseCase: nu,
		tracer:              tracer,
	}
}

func (f *friendshipUseCase) SendFriendRequest(ctx context.Context, fromID, toID primitive.ObjectID) error {
	ctx, span := f.tracer.Start(ctx, "FriendshipUseCase.SendFriendRequest")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"fromID": fromID.Hex(),
		"toID":   toID.Hex(),
	}
	logger.Input(input)

	if fromID == toID {
		err := domain.ErrInvalidInput
		logger.Output("invalid input 1", err)
		return err
	}

	// Check if friendship already exists
	existing, err := f.friendshipRepo.FindByUsers(ctx, fromID, toID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		logger.Output("error checking existing friendship 1", err)
		return domain.ErrInternalError
	}
	if existing != nil {
		if existing.Status == "blocked" {
			err := domain.ErrInvalidInput
			logger.Output("friendship already blocked 2", err)
			return err
		}
		if existing.Status == "pending" {
			err := domain.ErrFriendRequestAlreadySent
			logger.Output("friend request already sent 3", err)
			return err
		}
		if existing.Status == "accepted" {
			err := domain.ErrAlreadyFriends
			logger.Output("already friends 4", err)
			return err
		}
	}

	// Create friendship request
	friendship := &domain.Friendship{
		UserID1:     fromID,
		UserID2:     toID,
		Status:      "pending",
		RequestedBy: fromID,
	}

	err = f.friendshipRepo.Create(ctx, friendship)
	if err != nil {
		logger.Output("error creating friendship 5", err)
		return domain.ErrInternalError
	}

	// Create notification for friend request
	_, err = f.notificationUseCase.CreateNotification(
		ctx,
		toID,   // recipientID (user receiving the request)
		fromID, // senderID (user sending the request)
		fromID, // refID (reference to the requester)
		domain.NotificationTypeFriendReq,
		"user",                      // refType
		"sent you a friend request", // message
	)
	if err != nil {
		logger.Output("error creating notification 6", err)
		// Don't return error here as the friend request was successful
		// Just log the notification error
	}

	logger.Output("friendship created successfully", nil)
	return nil
}

func (f *friendshipUseCase) AcceptFriendRequest(ctx context.Context, userID, friendID primitive.ObjectID) error {
	ctx, span := f.tracer.Start(ctx, "FriendshipUseCase.AcceptFriendRequest")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"userID":   userID.Hex(),
		"friendID": friendID.Hex(),
	}
	logger.Input(input)

	// Find the friendship
	friendship, err := f.friendshipRepo.FindByUsers(ctx, friendID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.Output("friend request not found 1", domain.ErrFriendRequestNotFound)
			return domain.ErrFriendRequestNotFound
		}
		logger.Output("error finding friendship 2", err)
		return domain.ErrInternalError
	}

	if friendship.Status != "pending" {
		err = domain.ErrInvalidInput
		logger.Output("invalid friendship status 3", err)
		return err
	}

	if friendship.RequestedBy == userID {
		err = domain.ErrInvalidInput
		logger.Output("cannot accept own request 4", err)
		return err
	}

	friendship.Status = "accepted"
	friendship.UpdatedAt = time.Now()

	if err := f.friendshipRepo.Update(ctx, friendship); err != nil {
		logger.Output("error updating friendship 5", err)
		return domain.ErrInternalError
	}

	// Create notification for the user who sent the request
	_, err = f.notificationUseCase.CreateNotification(
		ctx,
		friendship.RequestedBy, // recipientID (user who sent the request)
		userID,                 // senderID (user accepting the request)
		userID,                 // refID (reference to the accepter)
		domain.NotificationTypeFriendReq,
		"user",                         // refType
		"accepted your friend request", // message
	)
	if err != nil {
		logger.Output("error creating notification 6", err)
		// Don't return error here as the friendship was accepted successfully
		// Just log the notification error
	}

	logger.Output("friendship accepted successfully", nil)
	return nil
}

func (f *friendshipUseCase) RejectFriendRequest(ctx context.Context, userID, friendID primitive.ObjectID) error {
	ctx, span := f.tracer.Start(ctx, "FriendshipUseCase.RejectFriendRequest")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"userID":   userID.Hex(),
		"friendID": friendID.Hex(),
	}
	logger.Input(input)

	friendship, err := f.friendshipRepo.FindByUsers(ctx, friendID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.Output("friend request not found 1", domain.ErrFriendRequestNotFound)
			return domain.ErrFriendRequestNotFound
		}
		logger.Output("error finding friendship 2", err)
		return domain.ErrInternalError
	}

	if friendship.Status != "pending" {
		err = domain.ErrInvalidInput
		logger.Output("invalid friendship status 3", err)
		return err
	}

	if friendship.RequestedBy == userID {
		err = domain.ErrInvalidInput
		logger.Output("cannot reject own request 4", err)
		return err
	}

	if err := f.friendshipRepo.Delete(ctx, friendID, userID); err != nil {
		logger.Output("error deleting friendship 5", err)
		return domain.ErrInternalError
	}

	logger.Output("friendship rejected successfully", nil)
	return nil
}

func (f *friendshipUseCase) CancelFriendRequest(ctx context.Context, userID, friendID primitive.ObjectID) error {
	ctx, span := f.tracer.Start(ctx, "FriendshipUseCase.CancelFriendRequest")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"userID":   userID.Hex(),
		"friendID": friendID.Hex(),
	}
	logger.Input(input)

	friendship, err := f.friendshipRepo.FindByUsers(ctx, userID, friendID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.Output("friend request not found 1", domain.ErrFriendRequestNotFound)
			return domain.ErrFriendRequestNotFound
		}
		logger.Output("error finding friendship 2", err)
		return domain.ErrInternalError
	}

	if friendship.Status != "pending" {
		err = domain.ErrInvalidInput
		logger.Output("invalid friendship status 3", err)
		return err
	}

	if friendship.RequestedBy != userID {
		err = domain.ErrInvalidInput
		logger.Output("cannot cancel others request 4", err)
		return err
	}

	if err := f.friendshipRepo.Delete(ctx, userID, friendID); err != nil {
		logger.Output("error deleting friendship 5", err)
		return domain.ErrInternalError
	}

	logger.Output("friendship cancelled successfully", nil)
	return nil
}

func (f *friendshipUseCase) Unfriend(ctx context.Context, userID1, userID2 primitive.ObjectID) error {
	ctx, span := f.tracer.Start(ctx, "FriendshipUseCase.Unfriend")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"userID1": userID1.Hex(),
		"userID2": userID2.Hex(),
	}
	logger.Input(input)

	friendship, err := f.friendshipRepo.FindByUsers(ctx, userID1, userID2)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.Output("friendship not found 1", domain.ErrFriendshipNotFound)
			return domain.ErrFriendshipNotFound
		}
		logger.Output("error finding friendship 2", err)
		return domain.ErrInternalError
	}

	if friendship.Status != "accepted" {
		err = domain.ErrInvalidInput
		logger.Output("invalid friendship status 3", err)
		return err
	}

	if err := f.friendshipRepo.Delete(ctx, userID1, userID2); err != nil {
		logger.Output("error deleting friendship 4", err)
		return domain.ErrInternalError
	}

	logger.Output("friendship unfriended successfully", nil)
	return nil
}

func (f *friendshipUseCase) BlockFriend(ctx context.Context, userID, blockedID primitive.ObjectID) error {
	ctx, span := f.tracer.Start(ctx, "FriendshipUseCase.BlockFriend")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"userID":    userID.Hex(),
		"blockedID": blockedID.Hex(),
	}
	logger.Input(input)

	if userID == blockedID {
		err := domain.ErrInvalidInput
		logger.Output("cannot block yourself 1", err)
		return err
	}

	friendship, err := f.friendshipRepo.FindByUsers(ctx, userID, blockedID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		logger.Output("error checking existing block 2", err)
		return domain.ErrInternalError
	}

	if friendship != nil {
		friendship.Status = "blocked"
		friendship.UpdatedAt = time.Now()
		if err := f.friendshipRepo.Update(ctx, friendship); err != nil {
			logger.Output("error updating block status 3", err)
			return domain.ErrInternalError
		}
	} else {
		friendship = &domain.Friendship{
			UserID1:     userID,
			UserID2:     blockedID,
			Status:      "blocked",
			RequestedBy: userID,
		}

		if err := f.friendshipRepo.Create(ctx, friendship); err != nil {
			logger.Output("error creating block 4", err)
			return domain.ErrInternalError
		}
	}

	logger.Output("user blocked successfully", nil)
	return nil
}

func (f *friendshipUseCase) UnblockFriend(ctx context.Context, userID, blockedID primitive.ObjectID) error {
	ctx, span := f.tracer.Start(ctx, "FriendshipUseCase.UnblockFriend")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"userID":    userID.Hex(),
		"blockedID": blockedID.Hex(),
	}
	logger.Input(input)

	friendship, err := f.friendshipRepo.FindByUsers(ctx, userID, blockedID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.Output("block not found 1", domain.ErrFriendshipNotFound)
			return domain.ErrFriendshipNotFound
		}
		logger.Output("error finding block 2", err)
		return domain.ErrInternalError
	}

	if friendship.Status != "blocked" {
		err = domain.ErrInvalidInput
		logger.Output("user not blocked 3", err)
		return err
	}

	if err := f.friendshipRepo.Delete(ctx, userID, blockedID); err != nil {
		logger.Output("error removing block 4", err)
		return domain.ErrInternalError
	}

	logger.Output("user unblocked successfully", nil)
	return nil
}

func (f *friendshipUseCase) FindFriends(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]domain.Friendship, error) {
	ctx, span := f.tracer.Start(ctx, "FriendshipUseCase.FindFriends")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"userID": userID.Hex(),
		"limit":  limit,
		"offset": offset,
	}
	logger.Input(input)

	friends, err := f.friendshipRepo.FindFriends(ctx, userID, limit, offset)
	if err != nil {
		logger.Output("error finding friends 1", err)
		return nil, domain.ErrInternalError
	}

	logger.Output("friends found successfully", nil)
	return friends, nil
}

func (f *friendshipUseCase) FindPendingRequests(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]domain.Friendship, error) {
	ctx, span := f.tracer.Start(ctx, "FriendshipUseCase.FindPendingRequests")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"userID": userID.Hex(),
		"limit":  limit,
		"offset": offset,
	}
	logger.Input(input)

	requests, err := f.friendshipRepo.FindPendingRequests(ctx, userID, limit, offset)
	if err != nil {
		logger.Output("error finding pending requests 1", err)
		return nil, domain.ErrInternalError
	}

	logger.Output("pending requests found successfully", nil)
	return requests, nil
}

func (f *friendshipUseCase) IsFriend(ctx context.Context, userID1, userID2 primitive.ObjectID) (bool, error) {
	ctx, span := f.tracer.Start(ctx, "FriendshipUseCase.IsFriend")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"userID1": userID1.Hex(),
		"userID2": userID2.Hex(),
	}
	logger.Input(input)

	friendship, err := f.friendshipRepo.FindByUsers(ctx, userID1, userID2)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.Output("users are not friends 1", nil)
			return false, nil
		}
		logger.Output("error checking friendship 2", err)
		return false, domain.ErrInternalError
	}

	isFriend := friendship.Status == "accepted"
	logger.Output("friendship status checked successfully", nil)
	return isFriend, nil
}

func (f *friendshipUseCase) FindFriendshipStatus(ctx context.Context, userID1, userID2 primitive.ObjectID) (string, error) {
	ctx, span := f.tracer.Start(ctx, "FriendshipUseCase.FindFriendshipStatus")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"userID1": userID1.Hex(),
		"userID2": userID2.Hex(),
	}
	logger.Input(input)

	friendship, err := f.friendshipRepo.FindByUsers(ctx, userID1, userID2)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.Output("no friendship found 1", nil)
			return "none", nil
		}
		logger.Output("error finding friendship status 2", err)
		return "", domain.ErrInternalError
	}

	logger.Output("friendship status found", nil)
	return friendship.Status, nil
}

func (f *friendshipUseCase) FindManyFriends(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]domain.Friendship, error) {
	ctx, span := f.tracer.Start(ctx, "FriendshipUseCase.FindManyFriends")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"userID": userID.Hex(),
		"limit":  limit,
		"offset": offset,
	}
	logger.Input(input)

	friends, err := f.friendshipRepo.FindFriends(ctx, userID, limit, offset)
	if err != nil {
		logger.Output("error finding many friends 1", err)
		return nil, domain.ErrInternalError
	}

	logger.Output("friends found successfully", nil)
	return friends, nil
}

func (f *friendshipUseCase) FindManyFriendRequests(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]domain.Friendship, error) {
	ctx, span := f.tracer.Start(ctx, "FriendshipUseCase.FindManyFriendRequests")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"userID": userID.Hex(),
		"limit":  limit,
		"offset": offset,
	}
	logger.Input(input)

	requests, err := f.friendshipRepo.FindPendingRequests(ctx, userID, limit, offset)
	if err != nil {
		logger.Output("error finding many friend requests 1", err)
		return nil, domain.ErrInternalError
	}

	logger.Output("friend requests found successfully", nil)
	return requests, nil
}

func (f *friendshipUseCase) RemoveFriend(ctx context.Context, userID, targetID primitive.ObjectID) error {
	ctx, span := f.tracer.Start(ctx, "FriendshipUseCase.RemoveFriend")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"userID":   userID.Hex(),
		"targetID": targetID.Hex(),
	}
	logger.Input(input)

	err := f.friendshipRepo.RemoveFriend(ctx, userID, targetID)
	if err != nil {
		logger.Output("error removing friend 1", err)
		return domain.ErrInternalError
	}

	logger.Output("friend removed successfully", nil)
	return nil
}
