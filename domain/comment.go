package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Comment struct {
	ID             primitive.ObjectID   `bson:"id,omitempty"`
	PostID         primitive.ObjectID   `bson:"postId"`
	UserID         primitive.ObjectID   `bson:"userId"`
	Content        string               `bson:"content"`
	Media          []Media              `bson:"media,omitempty"`
	ReactionCounts map[string]int       `bson:"reactionCounts"`
	CreatedAt      time.Time            `bson:"createdAt"`
	UpdatedAt      time.Time            `bson:"updatedAt"`
	ReplyTo        *primitive.ObjectID  `bson:"replyTo,omitempty"`
}

// Repository interface
type CommentRepository interface {
	Create(comment *Comment) error
	Update(comment *Comment) error
	Delete(id primitive.ObjectID) error
	FindByID(id primitive.ObjectID) (*Comment, error)
	FindByPostID(postID primitive.ObjectID, limit, offset int) ([]Comment, error)
}

// UseCase interface
type CommentUseCase interface {
	CreateComment(userID, postID primitive.ObjectID, content string, media []Media, replyTo *primitive.ObjectID) (*Comment, error)
	UpdateComment(commentID primitive.ObjectID, content string, media []Media) (*Comment, error)
	DeleteComment(commentID primitive.ObjectID) error
	GetComment(commentID primitive.ObjectID) (*Comment, error)
	ListComments(postID primitive.ObjectID, limit, offset int) ([]Comment, error)
}
