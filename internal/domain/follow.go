package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"

	"context"
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
	Create(ctx context.Context, follow *Follow) error
	Delete(ctx context.Context, followerID, followingID primitive.ObjectID) error
	FindByFollowerAndFollowing(ctx context.Context, followerID, followingID primitive.ObjectID) (*Follow, error)
	FindFollowers(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]Follow, error)
	FindFollowing(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]Follow, error)
	CountFollowers(ctx context.Context, userID primitive.ObjectID) (int64, error)
	CountFollowing(ctx context.Context, userID primitive.ObjectID) (int64, error)
	UpdateStatus(ctx context.Context, followerID, followingID primitive.ObjectID, status string) error
}

// FollowUseCase interface defines business logic for follows
type FollowUseCase interface {
	Follow(ctx context.Context, followerID, followingID primitive.ObjectID) error
	Unfollow(ctx context.Context, followerID, followingID primitive.ObjectID) error
	Block(ctx context.Context, userID, blockedID primitive.ObjectID) error
	Unblock(ctx context.Context, userID, blockedID primitive.ObjectID) error
	FindFollowers(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]Follow, error)
	FindFollowing(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]Follow, error)
	IsFollowing(ctx context.Context, followerID, followingID primitive.ObjectID) (bool, error)
	IsBlocked(ctx context.Context, userID, blockedID primitive.ObjectID) (bool, error)
}
