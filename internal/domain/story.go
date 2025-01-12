package domain

import (
	"context"
	"time"
)

type StoryType string

const (
	Image StoryType = "image"
	Video StoryType = "video"
)

type StoryMedia struct {
	URL       string    `bson:"url" json:"url"`
	Type      StoryType `bson:"type" json:"type"`
	Duration  int       `bson:"duration" json:"duration"` // Duration in seconds for videos
	Thumbnail string    `bson:"thumbnail" json:"thumbnail"`
}

type StoryViewer struct {
	UserID    string    `bson:"userId" json:"userId"`
	ViewedAt  time.Time `bson:"viewedAt" json:"viewedAt"`
	IsArchive bool      `bson:"isArchive" json:"isArchive"`
}

type Story struct {
	BaseModel    `bson:",inline"`
	UserID       string        `bson:"userId" json:"userId"`
	Media        StoryMedia    `bson:"media" json:"media"`
	Caption      string        `bson:"caption" json:"caption"`
	Location     string        `bson:"location" json:"location"`
	ViewersCount int           `bson:"viewersCount" json:"viewersCount"`
	Viewers      []StoryViewer `bson:"viewers" json:"viewers"`
	ExpiresAt    time.Time     `bson:"expiresAt" json:"expiresAt"`
	IsArchive    bool          `bson:"isArchive" json:"isArchive"`
	IsActive     bool          `bson:"isActive" json:"isActive"`
}

type StoryResponse struct {
	*Story
	User struct {
		ID           string `json:"userId"`
		Username     string `json:"username"`
		DisplayName  string `json:"displayName"`
		PhotoProfile string `json:"photoProfile"`
		FirstName    string `json:"firstName"`
		LastName     string `json:"lastName"`
	} `json:"user"`
}

type StoryRepository interface {
	Create(ctx context.Context, story *Story) error
	FindByID(ctx context.Context, id string) (*Story, error)
	FindByUserID(ctx context.Context, userID string) ([]*Story, error)
	FindActiveStories(ctx context.Context) ([]*Story, error)
	Update(ctx context.Context, story *Story) error
	AddViewer(ctx context.Context, storyID string, viewer StoryViewer) error
	DeleteStory(ctx context.Context, id string) error
	ArchiveExpiredStories() error
}

type StoryUseCase interface {
	CreateStory(ctx context.Context, story *Story) error
	FindStoryByID(ctx context.Context, id string) (*StoryResponse, error)
	FindUserStories(ctx context.Context, userID string) ([]*StoryResponse, error)
	FindActiveStories(ctx context.Context) ([]*StoryResponse, error)
	ViewStory(ctx context.Context, storyID string, viewerID string) error
	DeleteStory(ctx context.Context, storyID string, userID string) error
	ArchiveExpiredStories(ctx context.Context) error
}
