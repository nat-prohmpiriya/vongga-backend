package usecase

import (
	"errors"
	"time"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type friendshipUseCase struct {
	friendshipRepo      domain.FriendshipRepository
	notificationUseCase domain.NotificationUseCase
}

// NewFriendshipUseCase creates a new instance of FriendshipUseCase
func NewFriendshipUseCase(fr domain.FriendshipRepository, nu domain.NotificationUseCase) domain.FriendshipUseCase {
	return &friendshipUseCase{
		friendshipRepo:      fr,
		notificationUseCase: nu,
	}
}

// SendFriendRequest creates a new friendship request
func (f *friendshipUseCase) SendFriendRequest(fromID, toID primitive.ObjectID) error {
	logger := utils.NewLogger("FriendshipUseCase.SendFriendRequest")
	input := map[string]interface{}{
		"fromID": fromID.Hex(),
		"toID":   toID.Hex(),
	}
	logger.LogInput(input)

	if fromID == toID {
		err := domain.ErrInvalidInput
		logger.LogOutput(nil, err)
		return err
	}

	// Check if friendship already exists
	existing, err := f.friendshipRepo.FindByUsers(fromID, toID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		logger.LogOutput(nil, err)
		return domain.ErrInternalError
	}
	if existing != nil {
		if existing.Status == "blocked" {
			err := domain.ErrInvalidInput
			logger.LogOutput(nil, err)
			return err
		}
		if existing.Status == "pending" {
			err := domain.ErrFriendRequestAlreadySent
			logger.LogOutput(nil, err)
			return err
		}
		if existing.Status == "accepted" {
			err := domain.ErrAlreadyFriends
			logger.LogOutput(nil, err)
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

	err = f.friendshipRepo.Create(friendship)
	if err != nil {
		logger.LogOutput(nil, err)
		return domain.ErrInternalError
	}

	// Create notification for friend request
	_, err = f.notificationUseCase.CreateNotification(
		toID,   // recipientID (user receiving the request)
		fromID, // senderID (user sending the request)
		fromID, // refID (reference to the requester)
		domain.NotificationTypeFriendReq,
		"user",                      // refType
		"sent you a friend request", // message
	)
	if err != nil {
		logger.LogOutput(nil, err)
		// Don't return error here as the friend request was successful
		// Just log the notification error
	}

	logger.LogOutput(friendship, nil)
	return nil
}

// AcceptFriendRequest accepts a pending friend request
func (f *friendshipUseCase) AcceptFriendRequest(userID, friendID primitive.ObjectID) error {
	logger := utils.NewLogger("FriendshipUseCase.AcceptFriendRequest")
	input := map[string]interface{}{
		"userID":   userID.Hex(),
		"friendID": friendID.Hex(),
	}
	logger.LogInput(input)

	// Find the friendship
	friendship, err := f.friendshipRepo.FindByUsers(friendID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			err = domain.ErrFriendRequestNotFound
		}
		logger.LogOutput(nil, err)
		return err
	}

	if friendship.Status != "pending" {
		err = domain.ErrInvalidInput
		logger.LogOutput(nil, err)
		return err
	}

	if friendship.RequestedBy == userID {
		err = domain.ErrInvalidInput
		logger.LogOutput(nil, err)
		return err
	}

	friendship.Status = "accepted"
	friendship.UpdatedAt = time.Now()

	if err := f.friendshipRepo.Update(friendship); err != nil {
		logger.LogOutput(nil, err)
		return domain.ErrInternalError
	}

	// Create notification for the user who sent the request
	_, err = f.notificationUseCase.CreateNotification(
		friendship.RequestedBy, // recipientID (user who sent the request)
		userID,                 // senderID (user accepting the request)
		userID,                 // refID (reference to the accepter)
		domain.NotificationTypeFriendReq,
		"user",                         // refType
		"accepted your friend request", // message
	)
	if err != nil {
		logger.LogOutput(nil, err)
		// Don't return error here as the accept action was successful
		// Just log the notification error
	}

	logger.LogOutput(friendship, nil)
	return nil
}

// RejectFriendRequest rejects a pending friend request
func (f *friendshipUseCase) RejectFriendRequest(userID, friendID primitive.ObjectID) error {
	logger := utils.NewLogger("FriendshipUseCase.RejectFriendRequest")
	input := map[string]interface{}{
		"userID":   userID.Hex(),
		"friendID": friendID.Hex(),
	}
	logger.LogInput(input)

	friendship, err := f.friendshipRepo.FindByUsers(friendID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			err = domain.ErrFriendRequestNotFound
		}
		logger.LogOutput(nil, err)
		return err
	}

	if friendship.Status != "pending" {
		err = domain.ErrInvalidInput
		logger.LogOutput(nil, err)
		return err
	}

	if friendship.RequestedBy == userID {
		err = domain.ErrInvalidInput
		logger.LogOutput(nil, err)
		return err
	}

	if err := f.friendshipRepo.Delete(friendID, userID); err != nil {
		logger.LogOutput(nil, err)
		return domain.ErrInternalError
	}

	logger.LogOutput(nil, nil)
	return nil
}

// CancelFriendRequest cancels a sent friend request
func (f *friendshipUseCase) CancelFriendRequest(userID, friendID primitive.ObjectID) error {
	logger := utils.NewLogger("FriendshipUseCase.CancelFriendRequest")
	input := map[string]interface{}{
		"userID":   userID.Hex(),
		"friendID": friendID.Hex(),
	}
	logger.LogInput(input)

	friendship, err := f.friendshipRepo.FindByUsers(userID, friendID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			err = domain.ErrFriendRequestNotFound
		}
		logger.LogOutput(nil, err)
		return err
	}

	if friendship.Status != "pending" {
		err = domain.ErrInvalidInput
		logger.LogOutput(nil, err)
		return err
	}

	if friendship.RequestedBy != userID {
		err = domain.ErrInvalidInput
		logger.LogOutput(nil, err)
		return err
	}

	if err := f.friendshipRepo.Delete(userID, friendID); err != nil {
		logger.LogOutput(nil, err)
		return domain.ErrInternalError
	}

	logger.LogOutput(nil, nil)
	return nil
}

// Unfriend removes an accepted friendship
func (f *friendshipUseCase) Unfriend(userID1, userID2 primitive.ObjectID) error {
	logger := utils.NewLogger("FriendshipUseCase.Unfriend")
	input := map[string]interface{}{
		"userID1": userID1.Hex(),
		"userID2": userID2.Hex(),
	}
	logger.LogInput(input)

	friendship, err := f.friendshipRepo.FindByUsers(userID1, userID2)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			err = domain.ErrFriendshipNotFound
		}
		logger.LogOutput(nil, err)
		return err
	}

	if friendship.Status != "accepted" {
		err = domain.ErrInvalidInput
		logger.LogOutput(nil, err)
		return err
	}

	if err := f.friendshipRepo.Delete(userID1, userID2); err != nil {
		logger.LogOutput(nil, err)
		return domain.ErrInternalError
	}

	logger.LogOutput(nil, nil)
	return nil
}

// BlockFriend blocks a user
func (f *friendshipUseCase) BlockFriend(userID, blockedID primitive.ObjectID) error {
	logger := utils.NewLogger("FriendshipUseCase.BlockFriend")
	input := map[string]interface{}{
		"userID":    userID.Hex(),
		"blockedID": blockedID.Hex(),
	}
	logger.LogInput(input)

	if userID == blockedID {
		err := domain.ErrInvalidInput
		logger.LogOutput(nil, err)
		return err
	}

	friendship, err := f.friendshipRepo.FindByUsers(userID, blockedID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		logger.LogOutput(nil, err)
		return domain.ErrInternalError
	}

	if friendship != nil {
		friendship.Status = "blocked"
		friendship.UpdatedAt = time.Now()
		if err := f.friendshipRepo.Update(friendship); err != nil {
			logger.LogOutput(nil, err)
			return domain.ErrInternalError
		}
	} else {
		// Create new blocked relationship
		friendship = &domain.Friendship{
			UserID1:     userID,
			UserID2:     blockedID,
			Status:      "blocked",
			RequestedBy: userID,
		}

		if err := f.friendshipRepo.Create(friendship); err != nil {
			logger.LogOutput(nil, err)
			return domain.ErrInternalError
		}
	}

	logger.LogOutput(friendship, nil)
	return nil
}

// UnblockFriend removes a block
func (f *friendshipUseCase) UnblockFriend(userID, blockedID primitive.ObjectID) error {
	logger := utils.NewLogger("FriendshipUseCase.UnblockFriend")
	input := map[string]interface{}{
		"userID":    userID.Hex(),
		"blockedID": blockedID.Hex(),
	}
	logger.LogInput(input)

	friendship, err := f.friendshipRepo.FindByUsers(userID, blockedID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			err = domain.ErrFriendshipNotFound
		}
		logger.LogOutput(nil, err)
		return err
	}

	if friendship.Status != "blocked" {
		err = domain.ErrInvalidInput
		logger.LogOutput(nil, err)
		return err
	}

	if err := f.friendshipRepo.Delete(userID, blockedID); err != nil {
		logger.LogOutput(nil, err)
		return domain.ErrInternalError
	}

	logger.LogOutput(nil, nil)
	return nil
}

// FindFriends returns a list of accepted friends
func (f *friendshipUseCase) FindFriends(userID primitive.ObjectID, limit, offset int) ([]domain.Friendship, error) {
	logger := utils.NewLogger("FriendshipUseCase.FindFriends")
	input := map[string]interface{}{
		"userID": userID.Hex(),
		"limit":  limit,
		"offset": offset,
	}
	logger.LogInput(input)

	friends, err := f.friendshipRepo.FindFriends(userID, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, domain.ErrInternalError
	}

	logger.LogOutput(friends, nil)
	return friends, nil
}

// FindPendingRequests returns a list of pending friend requests
func (f *friendshipUseCase) FindPendingRequests(userID primitive.ObjectID, limit, offset int) ([]domain.Friendship, error) {
	logger := utils.NewLogger("FriendshipUseCase.FindPendingRequests")
	input := map[string]interface{}{
		"userID": userID.Hex(),
		"limit":  limit,
		"offset": offset,
	}
	logger.LogInput(input)

	requests, err := f.friendshipRepo.FindPendingRequests(userID, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, domain.ErrInternalError
	}

	logger.LogOutput(requests, nil)
	return requests, nil
}

// IsFriend checks if two users are friends
func (f *friendshipUseCase) IsFriend(userID1, userID2 primitive.ObjectID) (bool, error) {
	logger := utils.NewLogger("FriendshipUseCase.IsFriend")
	input := map[string]interface{}{
		"userID1": userID1.Hex(),
		"userID2": userID2.Hex(),
	}
	logger.LogInput(input)

	friendship, err := f.friendshipRepo.FindByUsers(userID1, userID2)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.LogOutput(false, nil)
			return false, nil
		}
		logger.LogOutput(nil, err)
		return false, domain.ErrInternalError
	}

	isFriend := friendship.Status == "accepted"
	logger.LogOutput(isFriend, nil)
	return isFriend, nil
}

// FindFriendshipStatus returns the current friendship status between two users
func (f *friendshipUseCase) FindFriendshipStatus(userID1, userID2 primitive.ObjectID) (string, error) {
	logger := utils.NewLogger("FriendshipUseCase.FindFriendshipStatus")
	input := map[string]interface{}{
		"userID1": userID1.Hex(),
		"userID2": userID2.Hex(),
	}
	logger.LogInput(input)

	friendship, err := f.friendshipRepo.FindByUsers(userID1, userID2)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.LogOutput("none", nil)
			return "none", nil
		}
		logger.LogOutput(nil, err)
		return "", domain.ErrInternalError
	}

	logger.LogOutput(friendship.Status, nil)
	return friendship.Status, nil
}

// FindManyFriends returns a list of friends
func (f *friendshipUseCase) FindManyFriends(userID primitive.ObjectID, limit, offset int) ([]domain.Friendship, error) {
	return f.friendshipRepo.FindFriends(userID, limit, offset)
}

// FindManyFriendRequests returns a list of friend requests
func (f *friendshipUseCase) FindManyFriendRequests(userID primitive.ObjectID, limit, offset int) ([]domain.Friendship, error) {
	return f.friendshipRepo.FindPendingRequests(userID, limit, offset)
}

func (f *friendshipUseCase) RemoveFriend(userID, targetID primitive.ObjectID) error {
	return f.friendshipRepo.RemoveFriend(userID, targetID)
}
