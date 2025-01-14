package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"

	"context"
)

// Friendship represents a friendship relationship between two users
type Friendship struct {
	BaseModel
	UserID1     primitive.ObjectID `bson:"userId1" json:"userId1"`
	UserID2     primitive.ObjectID `bson:"userId2" json:"userId2"`
	Status      string             `bson:"status" json:"status"` // pending, accepted, blocked
	RequestedBy primitive.ObjectID `bson:"requestedBy" json:"requestedBy"`
}

// FriendshipRepository interface defines methods for friendship persistence
type FriendshipRepository interface {
	Create(ctx context.Context, friendship *Friendship) error
	Update(ctx context.Context, friendship *Friendship) error
	Delete(ctx context.Context, userID1, userID2 primitive.ObjectID) error
	FindByUsers(ctx context.Context, userID1, userID2 primitive.ObjectID) (*Friendship, error)
	FindFriends(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]Friendship, error)
	FindPendingRequests(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]Friendship, error)
	CountFriends(ctx context.Context, userID primitive.ObjectID) (int64, error)
	CountPendingRequests(ctx context.Context, userID primitive.ObjectID) (int64, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*Friendship, error)
	RemoveFriend(ctx context.Context, userID, targetID primitive.ObjectID) error
}

// FriendshipUseCase interface defines business logic for friendships
type FriendshipUseCase interface {
	SendFriendRequest(ctx context.Context, fromID, toID primitive.ObjectID) error
	AcceptFriendRequest(ctx context.Context, userID, friendID primitive.ObjectID) error
	RejectFriendRequest(ctx context.Context, userID, friendID primitive.ObjectID) error
	CancelFriendRequest(ctx context.Context, userID, friendID primitive.ObjectID) error
	Unfriend(ctx context.Context, userID1, userID2 primitive.ObjectID) error
	BlockFriend(ctx context.Context, userID, blockedID primitive.ObjectID) error
	UnblockFriend(ctx context.Context, userID, blockedID primitive.ObjectID) error
	FindFriends(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]Friendship, error)
	FindPendingRequests(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]Friendship, error)
	IsFriend(ctx context.Context, userID1, userID2 primitive.ObjectID) (bool, error)
	FindFriendshipStatus(ctx context.Context, userID1, userID2 primitive.ObjectID) (string, error)
	FindManyFriends(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]Friendship, error)
	FindManyFriendRequests(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]Friendship, error)
	RemoveFriend(ctx context.Context, userID, targetID primitive.ObjectID) error
}
