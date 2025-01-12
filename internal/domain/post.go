package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	BaseModel      `bson:",inline"`
	UserID         primitive.ObjectID `bson:"userId" json:"userId"`
	Content        string             `bson:"content" json:"content"`
	Media          []Media            `bson:"media" json:"media"`
	ReactionCounts map[string]int     `bson:"reactionCounts" json:"reactionCounts"`
	CommentCount   int                `bson:"commentCount" json:"commentCount"`
	SubPostCount   int                `bson:"subPostCount" json:"subPostCount"`
	Tags           []string           `bson:"tags" json:"tags"`
	Location       *Location          `bson:"location,omitempty" json:"location,omitempty"`
	Visibility     string             `bson:"visibility" json:"visibility"`
	ShareCount     int                `bson:"shareCount" json:"shareCount"`
	ViewCount      int                `bson:"viewCount" json:"viewCount"`
	IsEdited       bool               `bson:"isEdited" json:"isEdited"`
	EditHistory    []EditLog          `bson:"editHistory" json:"editHistory"`
	AllowComments  bool               `bson:"allowComments" json:"allowComments"`
	AllowReactions bool               `bson:"allowReactions" json:"allowReactions"`
	PostType       string             `bson:"postType" json:"postType"`
}

type SubPost struct {
	BaseModel      `bson:",inline"`
	ParentID       primitive.ObjectID `bson:"parentId" json:"parentId"`
	UserID         primitive.ObjectID `bson:"userId" json:"userId"`
	Content        string             `bson:"content" json:"content"`
	Media          []Media            `bson:"media" json:"media"`
	ReactionCounts map[string]int     `bson:"reactionCounts" json:"reactionCounts"`
	CommentCount   int                `bson:"commentCount" json:"commentCount"`
	Order          int                `bson:"order" json:"order"`
	AllowComments  bool               `bson:"allowComments" json:"allowComments"`
	AllowReactions bool               `bson:"allowReactions" json:"allowReactions"`
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
	Content  string    `bson:"content" json:"content"`
	Media    []Media   `bson:"media" json:"media"`
	Tags     []string  `bson:"tags" json:"tags"`
	Location *Location `bson:"location,omitempty" json:"location,omitempty"`
	EditedAt time.Time `bson:"editedAt" json:"editedAt"`
}

type SubPostInput struct {
	Content string  `json:"content"`
	Media   []Media `json:"media,omitempty"`
	Order   int     `json:"order"`
}

const (
	MediaTypeImage = "image"
	MediaTypeVideo = "video"
)

// Repository interface
type PostRepository interface {
	Create(ctx context.Context, post *Post) error
	Update(ctx context.Context, post *Post) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*Post, error)
	FindByUserID(ctx context.Context, userID primitive.ObjectID, limit, offset int, hasMedia bool, mediaType string) ([]Post, error)
}

type SubPostRepository interface {
	Create(ctx context.Context, subPost *SubPost) error
	Update(ctx context.Context, subPost *SubPost) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*SubPost, error)
	FindByParentID(ctx context.Context, parentID primitive.ObjectID, limit, offset int) ([]SubPost, error)
	UpdateOrder(ctx context.Context, parentID primitive.ObjectID, orders map[primitive.ObjectID]int) error
}

// UseCase interface
type PostUseCase interface {
	CreatePost(ctx context.Context, userID primitive.ObjectID, content string, media []Media, tags []string, location *Location, visibility string, subPosts []SubPostInput) (*Post, error)
	UpdatePost(ctx context.Context, postID primitive.ObjectID, content string, media []Media, tags []string, location *Location, visibility string) (*Post, error)
	DeletePost(ctx context.Context, postID primitive.ObjectID) error
	FindPost(ctx context.Context, postID primitive.ObjectID, includeSubPosts bool) (*PostWithDetails, error)
	FindManyPosts(ctx context.Context, userID primitive.ObjectID, limit, offset int, includeSubPosts bool, hasMedia bool, mediaType string) ([]PostWithDetails, error)
}

type SubPostUseCase interface {
	CreateSubPost(ctx context.Context, parentID, userID primitive.ObjectID, content string, media []Media, order int) (*SubPost, error)
	UpdateSubPost(ctx context.Context, subPostID primitive.ObjectID, content string, media []Media) (*SubPost, error)
	DeleteSubPost(ctx context.Context, subPostID primitive.ObjectID) error
	FindSubPost(ctx context.Context, subPostID primitive.ObjectID) (*SubPost, error)
	FindManySubPosts(ctx context.Context, parentID primitive.ObjectID, limit, offset int) ([]SubPost, error)
	ReorderSubPosts(ctx context.Context, parentID primitive.ObjectID, orders map[primitive.ObjectID]int) error
}

// PostUser represents limited user data for post owner
type PostUser struct {
	ID           primitive.ObjectID `json:"userId"`
	Username     string             `json:"username"`
	DisplayName  string             `json:"displayName"`
	PhotoProfile string             `json:"photoProfile"`
	FirstName    string             `json:"firstName"`
	LastName     string             `json:"lastName"`
}

// PostWithDetails includes Post and its related data
type PostWithDetails struct {
	*Post
	User     *PostUser `json:"user"`
	SubPosts []SubPost `json:"subPosts,omitempty"`
}
