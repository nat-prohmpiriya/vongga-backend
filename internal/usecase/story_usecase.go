package usecase

import (
	"fmt"
	"time"

	"vongga-api/internal/domain"
	"vongga-api/utils"
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
	logger := utils.NewTraceLogger("StoryUseCase.CreateStory")
	logger.Input(story)

	// Validate user exists
	user, err := u.userRepo.FindByID(story.UserID)
	if err != nil {
		logger.Output(nil, err)
		return err
	}
	if user == nil {
		err = fmt.Errorf("user not found")
		logger.Output(nil, err)
		return err
	}

	// Validate media
	if story.Media.URL == "" {
		err = fmt.Errorf("media URL is required")
		logger.Output(nil, err)
		return err
	}

	if story.Media.Type != domain.Image && story.Media.Type != domain.Video {
		err = fmt.Errorf("invalid media type")
		logger.Output(nil, err)
		return err
	}

	// Create story
	err = u.storyRepo.Create(story)
	if err != nil {
		logger.Output(nil, err)
		return err
	}

	logger.Output(story, nil)
	return nil
}

func (u *storyUseCase) FindStoryByID(id string) (*domain.StoryResponse, error) {
	logger := utils.NewTraceLogger("StoryUseCase.FindStoryByID")
	logger.Input(id)

	story, err := u.storyRepo.FindByID(id)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	if story == nil {
		err = fmt.Errorf("story not found")
		logger.Output(nil, err)
		return nil, err
	}

	// Find user information
	user, err := u.userRepo.FindByID(story.UserID)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	response := &domain.StoryResponse{
		Story: story,
	}
	response.User.ID = user.ID.Hex()
	response.User.Username = user.Username
	response.User.DisplayName = user.DisplayName
	response.User.PhotoProfile = user.PhotoProfile
	response.User.FirstName = user.FirstName
	response.User.LastName = user.LastName

	logger.Output(response, nil)
	return response, nil
}

func (u *storyUseCase) FindUserStories(userID string) ([]*domain.StoryResponse, error) {
	logger := utils.NewTraceLogger("StoryUseCase.FindUserStories")
	logger.Input(userID)

	// Validate user exists
	user, err := u.userRepo.FindByID(userID)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}
	if user == nil {
		err = fmt.Errorf("user not found")
		logger.Output(nil, err)
		return nil, err
	}

	stories, err := u.storyRepo.FindByUserID(userID)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	var responses []*domain.StoryResponse
	for _, story := range stories {
		response := &domain.StoryResponse{
			Story: story,
		}
		response.User.ID = user.ID.Hex()
		response.User.Username = user.Username
		response.User.DisplayName = user.DisplayName
		response.User.PhotoProfile = user.PhotoProfile
		response.User.FirstName = user.FirstName
		response.User.LastName = user.LastName
		responses = append(responses, response)
	}

	logger.Output(responses, nil)
	return responses, nil
}

func (u *storyUseCase) FindActiveStories() ([]*domain.StoryResponse, error) {
	logger := utils.NewTraceLogger("StoryUseCase.FindActiveStories")

	stories, err := u.storyRepo.FindActiveStories()
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	var responses []*domain.StoryResponse
	for _, story := range stories {
		user, err := u.userRepo.FindByID(story.UserID)
		if err != nil {
			logger.Output(nil, err)
			continue
		}

		response := &domain.StoryResponse{
			Story: story,
		}
		response.User.ID = user.ID.Hex()
		response.User.Username = user.Username
		response.User.DisplayName = user.DisplayName
		response.User.PhotoProfile = user.PhotoProfile
		response.User.FirstName = user.FirstName
		response.User.LastName = user.LastName
		responses = append(responses, response)
	}

	logger.Output(responses, nil)
	return responses, nil
}

func (u *storyUseCase) ViewStory(storyID string, viewerID string) error {
	logger := utils.NewTraceLogger("StoryUseCase.ViewStory")
	input := map[string]interface{}{
		"storyID":  storyID,
		"viewerID": viewerID,
	}
	logger.Input(input)

	// Validate story exists and is active
	story, err := u.storyRepo.FindByID(storyID)
	if err != nil {
		logger.Output(nil, err)
		return err
	}
	if story == nil {
		err = fmt.Errorf("story not found")
		logger.Output(nil, err)
		return err
	}

	// Validate viewer exists
	viewer, err := u.userRepo.FindByID(viewerID)
	if err != nil {
		logger.Output(nil, err)
		return err
	}
	if viewer == nil {
		err = fmt.Errorf("viewer not found")
		logger.Output(nil, err)
		return err
	}

	// Check if story has expired
	if time.Now().After(story.ExpiresAt) {
		err = fmt.Errorf("story has expired")
		logger.Output(nil, err)
		return err
	}

	// Check if user has already viewed the story
	for _, v := range story.Viewers {
		if v.UserID == viewerID {
			logger.Output(nil, nil)
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
		logger.Output(nil, err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *storyUseCase) DeleteStory(storyID string, userID string) error {
	logger := utils.NewTraceLogger("StoryUseCase.DeleteStory")
	input := map[string]interface{}{
		"storyID": storyID,
		"userID":  userID,
	}
	logger.Input(input)

	// Validate story exists and belongs to user
	story, err := u.storyRepo.FindByID(storyID)
	if err != nil {
		logger.Output(nil, err)
		return err
	}
	if story == nil {
		err = fmt.Errorf("story not found")
		logger.Output(nil, err)
		return err
	}

	if story.UserID != userID {
		err = fmt.Errorf("unauthorized to delete this story")
		logger.Output(nil, err)
		return err
	}

	err = u.storyRepo.DeleteStory(storyID)
	if err != nil {
		logger.Output(nil, err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}

func (u *storyUseCase) ArchiveExpiredStories() error {
	logger := utils.NewTraceLogger("StoryUseCase.ArchiveExpiredStories")

	err := u.storyRepo.ArchiveExpiredStories()
	if err != nil {
		logger.Output(nil, err)
		return err
	}

	logger.Output(nil, nil)
	return nil
}
