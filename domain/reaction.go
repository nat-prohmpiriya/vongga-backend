package domain

import (
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
