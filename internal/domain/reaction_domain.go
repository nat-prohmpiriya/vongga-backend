package domain

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateReactionRequest represents the request payload for creating a reaction
type CreateReactionRequest struct {
	PostID    string `json:"postId" validate:"required_without=CommentID"`
	CommentID string `json:"commentId" validate:"required_without=PostID"`
	Type      string `json:"type" validate:"required,oneof=like love haha wow sad angry"`
}

type Reaction struct {
	BaseModel `bson:",inline"`
	PostID    primitive.ObjectID  `bson:"postId" json:"postId"`
	CommentID *primitive.ObjectID `bson:"commentId,omitempty" json:"commentId,omitempty"`
	UserID    primitive.ObjectID  `bson:"userId" json:"userId"`
	Type      string              `bson:"type" json:"type"`
}

// Repository interface
type ReactionRepository interface {
	Create(ctx context.Context, reaction *Reaction) error
	Update(ctx context.Context, reaction *Reaction) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*Reaction, error)
	FindByPostID(ctx context.Context, postID primitive.ObjectID, limit, offset int) ([]Reaction, error)
	FindByCommentID(ctx context.Context, commentID primitive.ObjectID, limit, offset int) ([]Reaction, error)
	FindByUserAndTarget(ctx context.Context, userID, postID primitive.ObjectID, commentID *primitive.ObjectID) (*Reaction, error)
}

// UseCase interface
type ReactionUseCase interface {
	CreateReaction(ctx context.Context, userID, postID primitive.ObjectID, commentID *primitive.ObjectID, reactionType string) (*Reaction, error)
	DeleteReaction(ctx context.Context, reactionID primitive.ObjectID) error
	FindReaction(ctx context.Context, reactionID primitive.ObjectID) (*Reaction, error)
	FindManyReactions(ctx context.Context, targetID primitive.ObjectID, isComment bool, limit, offset int) ([]Reaction, error)
}
