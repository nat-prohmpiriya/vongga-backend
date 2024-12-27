package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
)

type StoryHandler struct {
	storyUseCase domain.StoryUseCase
}

func NewStoryHandler(router fiber.Router, storyUseCase domain.StoryUseCase) *StoryHandler {
	handler := &StoryHandler{
		storyUseCase: storyUseCase,
	}

	router.Post("/", handler.CreateStory)
	router.Get("/active", handler.GetActiveStories)
	router.Get("/user/:userId", handler.GetUserStories)
	router.Get("/:storyId", handler.GetStoryByID)
	router.Post("/:storyId/view", handler.ViewStory)
	router.Delete("/:storyId", handler.DeleteStory)

	return handler
}

func (h *StoryHandler) CreateStory(c *fiber.Ctx) error {
	logger := utils.NewLogger("StoryHandler.CreateStory")

	userID := c.Locals("userId").(string)

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
		UserID: userID,
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
	err := h.storyUseCase.CreateStory(story)
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

func (h *StoryHandler) GetStoryByID(c *fiber.Ctx) error {
	logger := utils.NewLogger("StoryHandler.GetStoryByID")

	storyID := c.Params("storyId")
	logger.LogInput(storyID)

	story, err := h.storyUseCase.GetStoryByID(storyID)
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

func (h *StoryHandler) GetUserStories(c *fiber.Ctx) error {
	logger := utils.NewLogger("StoryHandler.GetUserStories")

	userID := c.Params("userId")
	logger.LogInput(userID)

	stories, err := h.storyUseCase.GetUserStories(userID)
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

func (h *StoryHandler) GetActiveStories(c *fiber.Ctx) error {
	logger := utils.NewLogger("StoryHandler.GetActiveStories")

	stories, err := h.storyUseCase.GetActiveStories()
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
	viewerID := c.Locals("userId").(string)

	input := map[string]interface{}{
		"storyId":  storyID,
		"viewerId": viewerID,
	}
	logger.LogInput(input)

	err := h.storyUseCase.ViewStory(storyID, viewerID)
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
	userID := c.Locals("userId").(string)

	input := map[string]interface{}{
		"storyId": storyID,
		"userId":  userID,
	}
	logger.LogInput(input)

	err := h.storyUseCase.DeleteStory(storyID, userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(nil, nil)
	return c.SendStatus(fiber.StatusOK)
}
