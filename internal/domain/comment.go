package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Comment struct {
	BaseModel      `bson:",inline"`
	PostID         primitive.ObjectID  `bson:"postId" json:"postId"`
	UserID         primitive.ObjectID  `bson:"userId" json:"userId"`
	Content        string              `bson:"content" json:"content"`
	Media          []Media             `bson:"media,omitempty" json:"media,omitempty"`
	ReactionCounts map[string]int      `bson:"reactionCounts" json:"reactionCounts"`
	ReplyTo        *primitive.ObjectID `bson:"replyTo,omitempty" json:"replyTo,omitempty"`
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

// CommentUser represents limited user data for comment owner
type CommentUser struct {
	ID           primitive.ObjectID `json:"userId"`
	Username     string            `json:"username"`
	DisplayName  string            `json:"displayName"`
	PhotoProfile string            `json:"photoProfile"`
	FirstName    string            `json:"firstName"`
	LastName     string            `json:"lastName"`
}

// CommentWithUser includes Comment and its related user data
type CommentWithUser struct {
	*Comment
	User *CommentUser `json:"user"`
}
