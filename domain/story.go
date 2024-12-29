package domain

import (
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
	ViewersCount int          `bson:"viewersCount" json:"viewersCount"`
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
	Create(story *Story) error
	FindByID(id string) (*Story, error)
	FindByUserID(userID string) ([]*Story, error)
	FindActiveStories() ([]*Story, error)
	Update(story *Story) error
	AddViewer(storyID string, viewer StoryViewer) error
	DeleteStory(id string) error
	ArchiveExpiredStories() error
}

type StoryUseCase interface {
	CreateStory(story *Story) error
	GetStoryByID(id string) (*StoryResponse, error)
	GetUserStories(userID string) ([]*StoryResponse, error)
	GetActiveStories() ([]*StoryResponse, error)
	ViewStory(storyID string, viewerID string) error
	DeleteStory(storyID string, userID string) error
	ArchiveExpiredStories() error
}
