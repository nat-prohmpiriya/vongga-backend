package handler

import (
	"vongga_api/internal/domain"
	"vongga_api/utils"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

type StoryHandler struct {
	storyUseCase domain.StoryUseCase
	tracer       trace.Tracer
}

func NewStoryHandler(router fiber.Router, storyUseCase domain.StoryUseCase, tracer trace.Tracer) *StoryHandler {
	handler := &StoryHandler{
		storyUseCase: storyUseCase,
		tracer:       tracer,
	}

	router.Post("", handler.CreateStory)
	router.Get("/active", handler.FindActiveStories)
	router.Get("/user/:userId", handler.FindUserStories)
	router.Get("/:storyId", handler.FindStoryByID)
	router.Post("/:storyId/view", handler.ViewStory)
	router.Delete("/:storyId", handler.DeleteStory)

	return handler
}

func (h *StoryHandler) CreateStory(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "StoryHandler.CreateStory")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("unauthorized access attempt 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req struct {
		MediaURL      string           `json:"mediaUrl"`
		MediaType     domain.StoryType `json:"mediaType"`
		MediaDuration int              `json:"mediaDuration,omitempty"`
		Thumbnail     string           `json:"thumbnail,omitempty"`
		Caption       string           `json:"caption,omitempty"`
		Location      string           `json:"location,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		logger.Output("invalid request body format 2", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate required fields
	if req.MediaURL == "" {
		logger.Output("missing required media URL 3", nil)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "media URL is required",
		})
	}

	if req.MediaType != domain.Image && req.MediaType != domain.Video {
		logger.Output("invalid media type provided 4", nil)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid media type",
		})
	}

	story := &domain.Story{
		UserID: userID.Hex(),
		Media: domain.StoryMedia{
			URL:       req.MediaURL,
			Type:      req.MediaType,
			Duration:  req.MediaDuration,
			Thumbnail: req.Thumbnail,
		},
		Caption:  req.Caption,
		Location: req.Location,
	}

	logger.Input(map[string]interface{}{
		"story": story,
	})
	err = h.storyUseCase.CreateStory(ctx, story)
	if err != nil {
		logger.Output("failed to create story 5", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(story, nil)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"story": story,
	})
}

func (h *StoryHandler) FindStoryByID(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "StoryHandler.FindStoryByID")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	storyID := c.Params("storyId")
	logger.Input(map[string]interface{}{
		"storyId": storyID,
	})

	if utils.IsUndefined(storyID) {
		logger.Output("missing required storyId 1", nil)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "storyId is required",
		})
	}

	story, err := h.storyUseCase.FindStoryByID(ctx, storyID)
	if err != nil {
		logger.Output("failed to find story 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(story, nil)
	return c.JSON(fiber.Map{
		"story": story,
	})
}

func (h *StoryHandler) FindUserStories(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "StoryHandler.FindUserStories")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID := c.Params("userId")
	logger.Input(map[string]interface{}{
		"userId": userID,
	})

	stories, err := h.storyUseCase.FindUserStories(ctx, userID)
	if err != nil {
		logger.Output("failed to find user stories 1", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(stories, nil)
	return c.JSON(fiber.Map{
		"stories": stories,
	})
}

func (h *StoryHandler) FindActiveStories(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "StoryHandler.FindActiveStories")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	stories, err := h.storyUseCase.FindActiveStories(ctx)
	if err != nil {
		logger.Output("failed to find active stories 1", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(stories, nil)
	return c.JSON(fiber.Map{
		"stories": stories,
	})
}

func (h *StoryHandler) ViewStory(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "StoryHandler.ViewStory")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	storyID := c.Params("storyId")
	viewerID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("unauthorized access attempt 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	logger.Input(map[string]interface{}{
		"storyId":  storyID,
		"viewerId": viewerID,
	})

	err = h.storyUseCase.ViewStory(ctx, storyID, viewerID.Hex())
	if err != nil {
		logger.Output("failed to record story view 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output("story viewed successfully", nil)
	return c.SendStatus(fiber.StatusOK)
}

func (h *StoryHandler) DeleteStory(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "StoryHandler.DeleteStory")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	storyID := c.Params("storyId")
	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("unauthorized access attempt 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	logger.Input(map[string]interface{}{
		"storyId": storyID,
		"userId":  userID,
	})

	err = h.storyUseCase.DeleteStory(ctx, storyID, userID.Hex())
	if err != nil {
		logger.Output("failed to delete story 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output("story deleted successfully", nil)
	return c.SendStatus(fiber.StatusOK)
}
