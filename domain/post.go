package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	ID             primitive.ObjectID   `bson:"id,omitempty" json:"id"`
	UserID         primitive.ObjectID   `bson:"userId" json:"userId"`
	Content        string               `bson:"content" json:"content"`
	Media          []Media              `bson:"media" json:"media"`
	ReactionCounts map[string]int       `bson:"reactionCounts" json:"reactionCounts"`
	CommentCount   int                  `bson:"commentCount" json:"commentCount"`
	SubPostCount   int                  `bson:"subPostCount" json:"subPostCount"`
	CreatedAt      time.Time            `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time            `bson:"updatedAt" json:"updatedAt"`
	Tags           []string             `bson:"tags" json:"tags"`
	Location       *Location            `bson:"location,omitempty" json:"location,omitempty"`
	Visibility     string               `bson:"visibility" json:"visibility"`
	ShareCount     int                  `bson:"shareCount" json:"shareCount"`
	ViewCount      int                  `bson:"viewCount" json:"viewCount"`
	IsEdited       bool                 `bson:"isEdited" json:"isEdited"`
	EditHistory    []EditLog            `bson:"editHistory" json:"editHistory"`
}

type SubPost struct {
	ID             primitive.ObjectID   `bson:"id,omitempty" json:"id"`
	ParentID       primitive.ObjectID   `bson:"parentId" json:"parentId"`
	UserID         primitive.ObjectID   `bson:"userId" json:"userId"`
	Content        string               `bson:"content" json:"content"`
	Media          []Media              `bson:"media" json:"media"`
	ReactionCounts map[string]int       `bson:"reactionCounts" json:"reactionCounts"`
	CommentCount   int                  `bson:"commentCount" json:"commentCount"`
	CreatedAt      time.Time            `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time            `bson:"updatedAt" json:"updatedAt"`
	Order          int                  `bson:"order" json:"order"`
}

type Media struct {
	Type         string  `bson:"type" json:"type"`
	URL          string  `bson:"url" json:"url"`
	ThumbnailURL string  `bson:"thumbnailUrl,omitempty" json:"thumbnailUrl,omitempty"`
	Description  string  `bson:"description,omitempty" json:"description,omitempty"`
	Size         int64   `bson:"size" json:"size"`
	Duration     float64 `bson:"duration,omitempty" json:"duration,omitempty"`
}

type Location struct {
	Type        string    `bson:"type" json:"type"`
	Coordinates []float64 `bson:"coordinates" json:"coordinates"`
	PlaceName   string    `bson:"placeName" json:"placeName"`
	Address     string    `bson:"address,omitempty" json:"address,omitempty"`
}

type EditLog struct {
	Content   string    `bson:"content" json:"content"`
	Media     []Media   `bson:"media" json:"media"`
	Tags      []string  `bson:"tags" json:"tags"`
	Location  *Location `bson:"location,omitempty" json:"location,omitempty"`
	EditedAt  time.Time `bson:"editedAt" json:"editedAt"`
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
	Post     *Post     `json:"post"`
	SubPosts []SubPost `json:"subPosts,omitempty"`
}
