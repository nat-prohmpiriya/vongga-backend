package handler

import (
	"vongga-api/internal/domain"
	"vongga-api/utils"

	"github.com/gofiber/fiber/v2"
)

type StoryHandler struct {
	storyUseCase domain.StoryUseCase
}

func NewStoryHandler(router fiber.Router, storyUseCase domain.StoryUseCase) *StoryHandler {
	handler := &StoryHandler{
		storyUseCase: storyUseCase,
	}

	router.Post("/", handler.CreateStory)
	router.Find("/active", handler.FindActiveStories)
	router.Find("/user/:userId", handler.FindUserStories)
	router.Find("/:storyId", handler.FindStoryByID)
	router.Post("/:storyId/view", handler.ViewStory)
	router.Delete("/:storyId", handler.DeleteStory)

	return handler
}

func (h *StoryHandler) CreateStory(c *fiber.Ctx) error {
	logger := utils.NewLogger("StoryHandler.CreateStory")

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
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
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate required fields
	if req.MediaURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "media URL is required",
		})
	}

	if req.MediaType != domain.Image && req.MediaType != domain.Video {
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

	logger.LogInput(story)
	err = h.storyUseCase.CreateStory(story)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(story, nil)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"story": story,
	})
}

func (h *StoryHandler) FindStoryByID(c *fiber.Ctx) error {
	logger := utils.NewLogger("StoryHandler.FindStoryByID")

	storyID := c.Params("storyId")
	logger.LogInput(storyID)

	if utils.IsUndefined(storyID) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "storyId is required",
		})
	}

	story, err := h.storyUseCase.FindStoryByID(storyID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(story, nil)
	return c.JSON(fiber.Map{
		"story": story,
	})
}

func (h *StoryHandler) FindUserStories(c *fiber.Ctx) error {
	logger := utils.NewLogger("StoryHandler.FindUserStories")

	userID := c.Params("userId")
	logger.LogInput(userID)

	stories, err := h.storyUseCase.FindUserStories(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(stories, nil)
	return c.JSON(fiber.Map{
		"stories": stories,
	})
}

func (h *StoryHandler) FindActiveStories(c *fiber.Ctx) error {
	logger := utils.NewLogger("StoryHandler.FindActiveStories")

	stories, err := h.storyUseCase.FindActiveStories()
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(stories, nil)
	return c.JSON(fiber.Map{
		"stories": stories,
	})
}

func (h *StoryHandler) ViewStory(c *fiber.Ctx) error {
	logger := utils.NewLogger("StoryHandler.ViewStory")

	storyID := c.Params("storyId")
	viewerID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	input := map[string]interface{}{
		"storyId":  storyID,
		"viewerId": viewerID,
	}
	logger.LogInput(input)

	err = h.storyUseCase.ViewStory(storyID, viewerID.Hex())
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(nil, nil)
	return c.SendStatus(fiber.StatusOK)
}

func (h *StoryHandler) DeleteStory(c *fiber.Ctx) error {
	logger := utils.NewLogger("StoryHandler.DeleteStory")

	storyID := c.Params("storyId")
	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	input := map[string]interface{}{
		"storyId": storyID,
		"userId":  userID,
	}
	logger.LogInput(input)

	err = h.storyUseCase.DeleteStory(storyID, userID.Hex())
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(nil, nil)
	return c.SendStatus(fiber.StatusOK)
}
