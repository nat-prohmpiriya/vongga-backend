package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateReactionRequest represents the request payload for creating a reaction
type CreateReactionRequest struct {
	PostID    string `json:"postId" validate:"required"`
	CommentID string `json:"commentId,omitempty"`
	Type      string `json:"type" validate:"required,oneof=like love haha wow sad angry"`
}

type Reaction struct {
	ID        primitive.ObjectID  `bson:"id,omitempty"`
	PostID    primitive.ObjectID  `bson:"postId"`
	CommentID *primitive.ObjectID `bson:"commentId,omitempty"`
	UserID    primitive.ObjectID  `bson:"userId"`
	Type      string             `bson:"type"`
	CreatedAt time.Time          `bson:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt"`
}

// Repository interface
type ReactionRepository interface {
	Create(reaction *Reaction) error
	Update(reaction *Reaction) error
	Delete(id primitive.ObjectID) error
	FindByID(id primitive.ObjectID) (*Reaction, error)
	FindByPostID(postID primitive.ObjectID, limit, offset int) ([]Reaction, error)
	FindByCommentID(commentID primitive.ObjectID, limit, offset int) ([]Reaction, error)
	FindByUserAndTarget(userID, postID primitive.ObjectID, commentID *primitive.ObjectID) (*Reaction, error)
}

// UseCase interface
type ReactionUseCase interface {
	CreateReaction(userID, postID primitive.ObjectID, commentID *primitive.ObjectID, reactionType string) (*Reaction, error)
	DeleteReaction(reactionID primitive.ObjectID) error
	GetReaction(reactionID primitive.ObjectID) (*Reaction, error)
	ListReactions(targetID primitive.ObjectID, isComment bool, limit, offset int) ([]Reaction, error)
}
