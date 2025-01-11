package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Follow represents a follow relationship between users
type Follow struct {
	BaseModel   `bson:",inline"`
	FollowerID  primitive.ObjectID `bson:"followerId" json:"followerId"`
	FollowingID primitive.ObjectID `bson:"followingId" json:"followingId"`
	Status      string             `bson:"status" json:"status"` // active, blocked
}

// FollowRepository interface defines methods for follow persistence
type FollowRepository interface {
	Create(follow *Follow) error
	Delete(followerID, followingID primitive.ObjectID) error
	FindByFollowerAndFollowing(followerID, followingID primitive.ObjectID) (*Follow, error)
	FindFollowers(userID primitive.ObjectID, limit, offset int) ([]Follow, error)
	FindFollowing(userID primitive.ObjectID, limit, offset int) ([]Follow, error)
	CountFollowers(userID primitive.ObjectID) (int64, error)
	CountFollowing(userID primitive.ObjectID) (int64, error)
	UpdateStatus(followerID, followingID primitive.ObjectID, status string) error
}

// FollowUseCase interface defines business logic for follows
type FollowUseCase interface {
	Follow(followerID, followingID primitive.ObjectID) error
	Unfollow(followerID, followingID primitive.ObjectID) error
	Block(userID, blockedID primitive.ObjectID) error
	Unblock(userID, blockedID primitive.ObjectID) error
	GetFollowers(userID primitive.ObjectID, limit, offset int) ([]Follow, error)
	GetFollowing(userID primitive.ObjectID, limit, offset int) ([]Follow, error)
	IsFollowing(followerID, followingID primitive.ObjectID) (bool, error)
	IsBlocked(userID, blockedID primitive.ObjectID) (bool, error)
}
