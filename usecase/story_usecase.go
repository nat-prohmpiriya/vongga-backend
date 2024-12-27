package usecase

import (
	"fmt"
	"time"

	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
)

type storyUseCase struct {
	storyRepo domain.StoryRepository
	userRepo  domain.UserRepository
}

func NewStoryUseCase(storyRepo domain.StoryRepository, userRepo domain.UserRepository) domain.StoryUseCase {
	return &storyUseCase{
		storyRepo: storyRepo,
		userRepo:  userRepo,
	}
}

func (u *storyUseCase) CreateStory(story *domain.Story) error {
	logger := utils.NewLogger("StoryUseCase.CreateStory")
	logger.LogInput(story)

	// Validate user exists
	user, err := u.userRepo.FindByID(story.UserID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}
	if user == nil {
		err = fmt.Errorf("user not found")
		logger.LogOutput(nil, err)
		return err
	}

	// Validate media
	if story.Media.URL == "" {
		err = fmt.Errorf("media URL is required")
		logger.LogOutput(nil, err)
		return err
	}

	if story.Media.Type != domain.Image && story.Media.Type != domain.Video {
		err = fmt.Errorf("invalid media type")
		logger.LogOutput(nil, err)
		return err
	}

	// Create story
	err = u.storyRepo.Create(story)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(story, nil)
	return nil
}

func (u *storyUseCase) GetStoryByID(id string) (*domain.Story, error) {
	logger := utils.NewLogger("StoryUseCase.GetStoryByID")
	logger.LogInput(id)

	story, err := u.storyRepo.FindByID(id)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	if story == nil {
		err = fmt.Errorf("story not found")
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(story, nil)
	return story, nil
}

func (u *storyUseCase) GetUserStories(userID string) ([]*domain.Story, error) {
	logger := utils.NewLogger("StoryUseCase.GetUserStories")
	logger.LogInput(userID)

	// Validate user exists
	user, err := u.userRepo.FindByID(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}
	if user == nil {
		err = fmt.Errorf("user not found")
		logger.LogOutput(nil, err)
		return nil, err
	}

	stories, err := u.storyRepo.FindByUserID(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(stories, nil)
	return stories, nil
}

func (u *storyUseCase) GetActiveStories() ([]*domain.Story, error) {
	logger := utils.NewLogger("StoryUseCase.GetActiveStories")

	stories, err := u.storyRepo.FindActiveStories()
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(stories, nil)
	return stories, nil
}

func (u *storyUseCase) ViewStory(storyID string, viewerID string) error {
	logger := utils.NewLogger("StoryUseCase.ViewStory")
	input := map[string]interface{}{
		"storyID":  storyID,
		"viewerID": viewerID,
	}
	logger.LogInput(input)

	// Validate story exists and is active
	story, err := u.storyRepo.FindByID(storyID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}
	if story == nil {
		err = fmt.Errorf("story not found")
		logger.LogOutput(nil, err)
		return err
	}

	// Validate viewer exists
	viewer, err := u.userRepo.FindByID(viewerID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}
	if viewer == nil {
		err = fmt.Errorf("viewer not found")
		logger.LogOutput(nil, err)
		return err
	}

	// Check if story has expired
	if time.Now().After(story.ExpiresAt) {
		err = fmt.Errorf("story has expired")
		logger.LogOutput(nil, err)
		return err
	}

	// Check if user has already viewed the story
	for _, v := range story.Viewers {
		if v.UserID == viewerID {
			logger.LogOutput(nil, nil)
			return nil // Already viewed
		}
	}

	// Create new viewer
	newViewer := domain.StoryViewer{
		UserID:    viewerID,
		ViewedAt:  time.Now(),
		IsArchive: false,
	}

	err = u.storyRepo.AddViewer(storyID, newViewer)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}

func (u *storyUseCase) DeleteStory(storyID string, userID string) error {
	logger := utils.NewLogger("StoryUseCase.DeleteStory")
	input := map[string]interface{}{
		"storyID": storyID,
		"userID":  userID,
	}
	logger.LogInput(input)

	// Validate story exists and belongs to user
	story, err := u.storyRepo.FindByID(storyID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}
	if story == nil {
		err = fmt.Errorf("story not found")
		logger.LogOutput(nil, err)
		return err
	}

	if story.UserID != userID {
		err = fmt.Errorf("unauthorized to delete this story")
		logger.LogOutput(nil, err)
		return err
	}

	err = u.storyRepo.DeleteStory(storyID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}

func (u *storyUseCase) ArchiveExpiredStories() error {
	logger := utils.NewLogger("StoryUseCase.ArchiveExpiredStories")

	err := u.storyRepo.ArchiveExpiredStories()
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(nil, nil)
	return nil
}
