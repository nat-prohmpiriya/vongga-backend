package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	ID             primitive.ObjectID   `bson:"id,omitempty"`
	UserID         primitive.ObjectID   `bson:"userId"`
	Content        string               `bson:"content"`
	Media          []Media              `bson:"media"`
	ReactionCounts map[string]int       `bson:"reactionCounts"`
	CommentCount   int                  `bson:"commentCount"`
	SubPostCount   int                  `bson:"subPostCount"`
	CreatedAt      time.Time            `bson:"createdAt"`
	UpdatedAt      time.Time            `bson:"updatedAt"`
	Tags           []string             `bson:"tags"`
	Location       *Location            `bson:"location,omitempty"`
	Visibility     string               `bson:"visibility"`
	ShareCount     int                  `bson:"shareCount"`
	ViewCount      int                  `bson:"viewCount"`
	IsEdited       bool                 `bson:"isEdited"`
	EditHistory    []EditLog            `bson:"editHistory"`
}

type SubPost struct {
	ID             primitive.ObjectID   `bson:"id,omitempty"`
	ParentID       primitive.ObjectID   `bson:"parentId"`
	UserID         primitive.ObjectID   `bson:"userId"`
	Content        string               `bson:"content"`
	Media          []Media              `bson:"media"`
	ReactionCounts map[string]int       `bson:"reactionCounts"`
	CommentCount   int                  `bson:"commentCount"`
	CreatedAt      time.Time            `bson:"createdAt"`
	UpdatedAt      time.Time            `bson:"updatedAt"`
	Order          int                  `bson:"order"`
}

type Media struct {
	Type         string  `bson:"type"`
	URL          string  `bson:"url"`
	ThumbnailURL string  `bson:"thumbnailUrl,omitempty"`
	Description  string  `bson:"description,omitempty"`
	Size         int64   `bson:"size"`
	Duration     float64 `bson:"duration,omitempty"`
}

type Location struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"`
	PlaceName   string    `bson:"placeName,omitempty"`
	Address     string    `bson:"address,omitempty"`
}

type EditLog struct {
	Content   string    `bson:"content"`
	Media     []Media   `bson:"media"`
	Tags      []string  `bson:"tags"`
	Location  *Location `bson:"location,omitempty"`
	EditedAt  time.Time `bson:"editedAt"`
}

// Repository interface
type PostRepository interface {
	Create(post *Post) error
	Update(post *Post) error
	Delete(id primitive.ObjectID) error
	FindByID(id primitive.ObjectID) (*Post, error)
	FindByUserID(userID primitive.ObjectID, limit, offset int) ([]Post, error)
}

type SubPostRepository interface {
	Create(subPost *SubPost) error
	Update(subPost *SubPost) error
	Delete(id primitive.ObjectID) error
	FindByID(id primitive.ObjectID) (*SubPost, error)
	FindByParentID(parentID primitive.ObjectID, limit, offset int) ([]SubPost, error)
	UpdateOrder(parentID primitive.ObjectID, orders map[primitive.ObjectID]int) error
}

// UseCase interface
type PostUseCase interface {
	CreatePost(userID primitive.ObjectID, content string, media []Media, tags []string, location *Location, visibility string) (*Post, error)
	UpdatePost(postID primitive.ObjectID, content string, media []Media, tags []string, location *Location, visibility string) (*Post, error)
	DeletePost(postID primitive.ObjectID) error
	GetPost(postID primitive.ObjectID, includeSubPosts bool) (*PostWithDetails, error)
	ListPosts(userID primitive.ObjectID, limit, offset int, includeSubPosts bool) ([]PostWithDetails, error)
}

type SubPostUseCase interface {
	CreateSubPost(parentID, userID primitive.ObjectID, content string, media []Media, order int) (*SubPost, error)
	UpdateSubPost(subPostID primitive.ObjectID, content string, media []Media) (*SubPost, error)
	DeleteSubPost(subPostID primitive.ObjectID) error
	GetSubPost(subPostID primitive.ObjectID) (*SubPost, error)
	ListSubPosts(parentID primitive.ObjectID, limit, offset int) ([]SubPost, error)
	ReorderSubPosts(parentID primitive.ObjectID, orders map[primitive.ObjectID]int) error
}

// PostWithDetails includes Post and its related data
type PostWithDetails struct {
	Post     *Post      `json:"post"`
	SubPosts []SubPost  `json:"subPosts,omitempty"`
}
