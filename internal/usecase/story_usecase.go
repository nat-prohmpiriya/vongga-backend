package usecase

import (
	"context"
	"fmt"
	"time"

	"vongga_api/internal/domain"
	"vongga_api/utils"

	"go.opentelemetry.io/otel/trace"
)

type storyUseCase struct {
	storyRepo domain.StoryRepository
	userRepo  domain.UserRepository
	tracer    trace.Tracer
}

func NewStoryUseCase(storyRepo domain.StoryRepository, userRepo domain.UserRepository, tracer trace.Tracer) domain.StoryUseCase {
	return &storyUseCase{
		storyRepo: storyRepo,
		userRepo:  userRepo,
		tracer:    tracer,
	}
}

func (u *storyUseCase) CreateStory(ctx context.Context, story *domain.Story) error {
	ctx, span := u.tracer.Start(ctx, "StoryUseCase.CreateStory")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	logger.Input(story)

	// Validate user exists
	user, err := u.userRepo.FindByID(ctx, story.UserID)
	if err != nil {
		logger.Output("error finding user 1", err)
		return err
	}
	if user == nil {
		err = fmt.Errorf("user not found")
		logger.Output("error user not found 2", err)
		return err
	}

	// Validate media
	if story.Media.URL == "" {
		err = fmt.Errorf("media URL is required")
		logger.Output("error media URL is required 3", err)
		return err
	}

	if story.Media.Type != domain.Image && story.Media.Type != domain.Video {
		err = fmt.Errorf("invalid media type")
		logger.Output("error invalid media type 4", err)
		return err
	}

	// Create story
	err = u.storyRepo.Create(ctx, story)
	if err != nil {
		logger.Output("error creating story 5", err)
		return err
	}

	logger.Output(story, nil)
	return nil
}

func (u *storyUseCase) FindStoryByID(ctx context.Context, id string) (*domain.StoryResponse, error) {
	ctx, span := u.tracer.Start(ctx, "StoryUseCase.FindStoryByID")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	logger.Input(id)

	story, err := u.storyRepo.FindByID(ctx, id)
	if err != nil {
		logger.Output("error finding story 1", err)
		return nil, err
	}

	if story == nil {
		err = fmt.Errorf("story not found")
		logger.Output("error story not found 2", err)
		return nil, err
	}

	// Find user information
	user, err := u.userRepo.FindByID(ctx, story.UserID)
	if err != nil {
		logger.Output("error finding user 3", err)
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

func (u *storyUseCase) FindUserStories(ctx context.Context, userID string) ([]*domain.StoryResponse, error) {
	ctx, span := u.tracer.Start(ctx, "StoryUseCase.FindUserStories")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	logger.Input(userID)

	// Validate user exists
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		logger.Output("error finding user 1", err)
		return nil, err
	}
	if user == nil {
		err = fmt.Errorf("user not found")
		logger.Output("error user not found 2", err)
		return nil, err
	}

	stories, err := u.storyRepo.FindByUserID(ctx, userID)
	if err != nil {
		logger.Output("error finding stories 3", err)
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

func (u *storyUseCase) FindActiveStories(ctx context.Context) ([]*domain.StoryResponse, error) {
	ctx, span := u.tracer.Start(ctx, "StoryUseCase.FindActiveStories")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	stories, err := u.storyRepo.FindActiveStories(ctx)
	if err != nil {
		logger.Output("error finding active stories 1", err)
		return nil, err
	}

	var responses []*domain.StoryResponse
	for _, story := range stories {
		user, err := u.userRepo.FindByID(ctx, story.UserID)
		if err != nil {
			logger.Output("error finding user 2", err)
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

func (u *storyUseCase) ViewStory(ctx context.Context, storyID string, viewerID string) error {
	ctx, span := u.tracer.Start(ctx, "StoryUseCase.ViewStory")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"storyID":  storyID,
		"viewerID": viewerID,
	}
	logger.Input(input)

	// Validate story exists and is active
	story, err := u.storyRepo.FindByID(ctx, storyID)
	if err != nil {
		logger.Output("error finding story 1", err)
		return err
	}
	if story == nil {
		err = fmt.Errorf("story not found")
		logger.Output("error story not found 2", err)
		return err
	}

	// Validate viewer exists
	viewer, err := u.userRepo.FindByID(ctx, viewerID)
	if err != nil {
		logger.Output("error finding viewer 3", err)
		return err
	}
	if viewer == nil {
		err = fmt.Errorf("viewer not found")
		logger.Output("error viewer not found 4", err)
		return err
	}

	// Check if story has expired
	if time.Now().After(story.ExpiresAt) {
		err = fmt.Errorf("story has expired")
		logger.Output("error story has expired 5", err)
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

	err = u.storyRepo.AddViewer(ctx, storyID, newViewer)
	if err != nil {
		logger.Output("error adding viewer 6", err)
		return err
	}

	logger.Output("Story viewed successfully", nil)
	return nil
}

func (u *storyUseCase) DeleteStory(ctx context.Context, storyID string, userID string) error {
	ctx, span := u.tracer.Start(ctx, "StoryUseCase.DeleteStory")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"storyID": storyID,
		"userID":  userID,
	}
	logger.Input(input)

	// Validate story exists and belongs to user
	story, err := u.storyRepo.FindByID(ctx, storyID)
	if err != nil {
		logger.Output("error finding story 1", err)
		return err
	}
	if story == nil {
		err = fmt.Errorf("story not found")
		logger.Output("error story not found 2", err)
		return err
	}

	if story.UserID != userID {
		err = fmt.Errorf("unauthorized to delete this story")
		logger.Output("error unauthorized to delete story 3", err)
		return err
	}

	err = u.storyRepo.DeleteStory(ctx, storyID)
	if err != nil {
		logger.Output("error deleting story 4", err)
		return err
	}

	logger.Output("Story deleted successfully", nil)
	return nil
}

func (u *storyUseCase) ArchiveExpiredStories(ctx context.Context) error {
	ctx, span := u.tracer.Start(ctx, "StoryUseCase.ArchiveExpiredStories")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	err := u.storyRepo.ArchiveExpiredStories(ctx)
	if err != nil {
		logger.Output("error archiving expired stories 1", err)
		return err
	}

	logger.Output("Expired stories archived successfully", nil)
	return nil
}
