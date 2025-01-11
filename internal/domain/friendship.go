package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	Create(friendship *Friendship) error
	Update(friendship *Friendship) error
	Delete(userID1, userID2 primitive.ObjectID) error
	FindByUsers(userID1, userID2 primitive.ObjectID) (*Friendship, error)
	FindFriends(userID primitive.ObjectID, limit, offset int) ([]Friendship, error)
	FindPendingRequests(userID primitive.ObjectID, limit, offset int) ([]Friendship, error)
	CountFriends(userID primitive.ObjectID) (int64, error)
	CountPendingRequests(userID primitive.ObjectID) (int64, error)
	FindByID(id primitive.ObjectID) (*Friendship, error)
	RemoveFriend(userID, targetID primitive.ObjectID) error
}

// FriendshipUseCase interface defines business logic for friendships
type FriendshipUseCase interface {
	SendFriendRequest(fromID, toID primitive.ObjectID) error
	AcceptFriendRequest(userID, friendID primitive.ObjectID) error
	RejectFriendRequest(userID, friendID primitive.ObjectID) error
	CancelFriendRequest(userID, friendID primitive.ObjectID) error
	Unfriend(userID1, userID2 primitive.ObjectID) error
	BlockFriend(userID, blockedID primitive.ObjectID) error
	UnblockFriend(userID, blockedID primitive.ObjectID) error
	GetFriends(userID primitive.ObjectID, limit, offset int) ([]Friendship, error)
	GetPendingRequests(userID primitive.ObjectID, limit, offset int) ([]Friendship, error)
	IsFriend(userID1, userID2 primitive.ObjectID) (bool, error)
	GetFriendshipStatus(userID1, userID2 primitive.ObjectID) (string, error)
	ListFriends(userID primitive.ObjectID, limit, offset int) ([]Friendship, error)
	ListFriendRequests(userID primitive.ObjectID, limit, offset int) ([]Friendship, error)
	RemoveFriend(userID, targetID primitive.ObjectID) error
}
