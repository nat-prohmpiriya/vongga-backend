package usecase

import (
	"errors"
	"time"

	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type friendshipUseCase struct {
	friendshipRepo domain.FriendshipRepository
}

// NewFriendshipUseCase creates a new instance of FriendshipUseCase
func NewFriendshipUseCase(fr domain.FriendshipRepository) domain.FriendshipUseCase {
	return &friendshipUseCase{
		friendshipRepo: fr,
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
		err := errors.New("cannot send friend request to yourself")
		logger.LogOutput(nil, err)
		return err
	}

	// Check if friendship already exists
	existing, err := f.friendshipRepo.FindByUsers(fromID, toID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		logger.LogOutput(nil, err)
		return err
	}
	if existing != nil {
		if existing.Status == "blocked" {
			err := errors.New("cannot send friend request to blocked user")
			logger.LogOutput(nil, err)
			return err
		}
		if existing.Status == "pending" {
			err := errors.New("friend request already sent")
			logger.LogOutput(nil, err)
			return err
		}
		if existing.Status == "accepted" {
			err := errors.New("already friends with this user")
			logger.LogOutput(nil, err)
			return err
		}
	}

	friendship := &domain.Friendship{
		UserID1:     fromID,
		UserID2:     toID,
		Status:      "pending",
		RequestedBy: fromID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := f.friendshipRepo.Create(friendship); err != nil {
		logger.LogOutput(nil, err)
		return err
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

	friendship, err := f.friendshipRepo.FindByUsers(friendID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			err = errors.New("friend request not found")
		}
		logger.LogOutput(nil, err)
		return err
	}

	if friendship.Status != "pending" {
		err = errors.New("no pending friend request")
		logger.LogOutput(nil, err)
		return err
	}

	if friendship.RequestedBy == userID {
		err = errors.New("cannot accept your own friend request")
		logger.LogOutput(nil, err)
		return err
	}

	friendship.Status = "accepted"
	friendship.UpdatedAt = time.Now()

	if err := f.friendshipRepo.Update(friendship); err != nil {
		logger.LogOutput(nil, err)
		return err
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
			err = errors.New("friend request not found")
		}
		logger.LogOutput(nil, err)
		return err
	}

	if friendship.Status != "pending" {
		err = errors.New("no pending friend request")
		logger.LogOutput(nil, err)
		return err
	}

	if friendship.RequestedBy == userID {
		err = errors.New("cannot reject your own friend request")
		logger.LogOutput(nil, err)
		return err
	}

	if err := f.friendshipRepo.Delete(friendID, userID); err != nil {
		logger.LogOutput(nil, err)
		return err
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
			err = errors.New("friend request not found")
		}
		logger.LogOutput(nil, err)
		return err
	}

	if friendship.Status != "pending" {
		err = errors.New("no pending friend request")
		logger.LogOutput(nil, err)
		return err
	}

	if friendship.RequestedBy != userID {
		err = errors.New("cannot cancel friend request sent by another user")
		logger.LogOutput(nil, err)
		return err
	}

	if err := f.friendshipRepo.Delete(userID, friendID); err != nil {
		logger.LogOutput(nil, err)
		return err
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
			err = errors.New("not friends with this user")
		}
		logger.LogOutput(nil, err)
		return err
	}

	if friendship.Status != "accepted" {
		err = errors.New("not friends with this user")
		logger.LogOutput(nil, err)
		return err
	}

	if err := f.friendshipRepo.Delete(userID1, userID2); err != nil {
		logger.LogOutput(nil, err)
		return err
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
		err := errors.New("cannot block yourself")
		logger.LogOutput(nil, err)
		return err
	}

	friendship, err := f.friendshipRepo.FindByUsers(userID, blockedID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		logger.LogOutput(nil, err)
		return err
	}

	if friendship != nil {
		friendship.Status = "blocked"
		friendship.UpdatedAt = time.Now()
		if err := f.friendshipRepo.Update(friendship); err != nil {
			logger.LogOutput(nil, err)
			return err
		}
	} else {
		// Create new blocked relationship
		friendship = &domain.Friendship{
			UserID1:     userID,
			UserID2:     blockedID,
			Status:      "blocked",
			RequestedBy: userID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := f.friendshipRepo.Create(friendship); err != nil {
			logger.LogOutput(nil, err)
			return err
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
			err = errors.New("user is not blocked")
		}
		logger.LogOutput(nil, err)
		return err
	}

	if friendship.Status != "blocked" {
		err = errors.New("user is not blocked")
		logger.LogOutput(nil, err)
		return err
	}

	if err := f.friendshipRepo.Delete(userID, blockedID); err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}

// GetFriends returns a list of accepted friends
func (f *friendshipUseCase) GetFriends(userID primitive.ObjectID, limit, offset int) ([]domain.Friendship, error) {
	logger := utils.NewLogger("FriendshipUseCase.GetFriends")
	input := map[string]interface{}{
		"userID": userID.Hex(),
		"limit":  limit,
		"offset": offset,
	}
	logger.LogInput(input)

	friends, err := f.friendshipRepo.FindFriends(userID, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(friends, nil)
	return friends, nil
}

// GetPendingRequests returns a list of pending friend requests
func (f *friendshipUseCase) GetPendingRequests(userID primitive.ObjectID, limit, offset int) ([]domain.Friendship, error) {
	logger := utils.NewLogger("FriendshipUseCase.GetPendingRequests")
	input := map[string]interface{}{
		"userID": userID.Hex(),
		"limit":  limit,
		"offset": offset,
	}
	logger.LogInput(input)

	requests, err := f.friendshipRepo.FindPendingRequests(userID, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
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
		return false, err
	}

	isFriend := friendship.Status == "accepted"
	logger.LogOutput(isFriend, nil)
	return isFriend, nil
}

// GetFriendshipStatus returns the current friendship status between two users
func (f *friendshipUseCase) GetFriendshipStatus(userID1, userID2 primitive.ObjectID) (string, error) {
	logger := utils.NewLogger("FriendshipUseCase.GetFriendshipStatus")
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
		return "", err
	}

	logger.LogOutput(friendship.Status, nil)
	return friendship.Status, nil
}
