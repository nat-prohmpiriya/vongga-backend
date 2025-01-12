package handler

import (
	"fmt"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostHandler struct {
	postUseCase domain.PostUseCase
}

func NewPostHandler(router fiber.Router, pu domain.PostUseCase) *PostHandler {
	handler := &PostHandler{
		postUseCase: pu,
	}

	router.Post("/", handler.CreatePost)
	router.Find("/", handler.FindManyPosts)
	router.Find("/:id", handler.FindPost)
	router.Put("/:id", handler.UpdatePost)
	router.Delete("/:id", handler.DeletePost)

	return handler
}

type CreatePostRequest struct {
	Content    string                `json:"content"`
	Media      []domain.Media        `json:"media,omitempty"`
	Tags       []string              `json:"tags,omitempty"`
	Location   *domain.Location      `json:"location,omitempty"`
	Visibility string                `json:"visibility"`
	SubPosts   []domain.SubPostInput `json:"subPosts,omitempty"`
}

type UpdatePostRequest struct {
	Content    string           `json:"content"`
	Media      []domain.Media   `json:"media,omitempty"`
	Tags       []string         `json:"tags,omitempty"`
	Location   *domain.Location `json:"location,omitempty"`
	Visibility string           `json:"visibility"`
}

func (h *PostHandler) CreatePost(c *fiber.Ctx) error {
	logger := utils.NewLogger("PostHandler.CreatePost")

	var req CreatePostRequest
	if err := c.BodyParser(&req); err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	post, err := h.postUseCase.CreatePost(
		userID,
		req.Content,
		req.Media,
		req.Tags,
		req.Location,
		req.Visibility,
		req.SubPosts,
	)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(post, nil)
	return c.Status(fiber.StatusCreated).JSON(post)
}

func (h *PostHandler) UpdatePost(c *fiber.Ctx) error {
	logger := utils.NewLogger("PostHandler.UpdatePost")

	postID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	var req UpdatePostRequest
	if err := c.BodyParser(&req); err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	input := map[string]interface{}{
		"postID":  postID,
		"request": req,
	}
	logger.LogInput(input)

	post, err := h.postUseCase.UpdatePost(postID, req.Content, req.Media, req.Tags, req.Location, req.Visibility)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(post, nil)
	return c.JSON(post)
}

func (h *PostHandler) DeletePost(c *fiber.Ctx) error {
	logger := utils.NewLogger("PostHandler.DeletePost")

	postID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}
	logger.LogInput(postID)

	err = h.postUseCase.DeletePost(postID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput("Post deleted successfully", nil)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *PostHandler) FindPost(c *fiber.Ctx) error {
	logger := utils.NewLogger("PostHandler.FindPost")

	postID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	includeSubPosts := c.Query("includeSubPosts") == "true"
	input := map[string]interface{}{
		"postID":          postID,
		"includeSubPosts": includeSubPosts,
	}
	logger.LogInput(input)

	post, err := h.postUseCase.FindPost(postID, includeSubPosts)
	if err != nil {
		logger.LogOutput(nil, err)
		if domain.IsNotFoundError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(post, nil)
	return c.JSON(post)
}

func (h *PostHandler) FindManyPosts(c *fiber.Ctx) error {
	logger := utils.NewLogger("PostHandler.FindManyPosts")

	userIDStr := c.Query("userId")
	if userIDStr == "" {
		logger.LogOutput(nil, fmt.Errorf("missing userId query parameter"))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing user ID",
		})
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	limit := c.QueryInt("limit", 0)
	offset := c.QueryInt("offset", 0)
	includeSubPosts := c.Query("includeSubPosts") == "true"
	hasMedia := c.Query("hasMedia") == "true"
	mediaType := c.Query("mediaType")

	// Validate mediaType if hasMedia is true
	if hasMedia && mediaType != "" && mediaType != domain.MediaTypeImage && mediaType != domain.MediaTypeVideo {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid media type. Must be 'image' or 'video'",
		})
	}

	input := map[string]interface{}{
		"userID":          userID,
		"limit":           limit,
		"offset":          offset,
		"includeSubPosts": includeSubPosts,
		"hasMedia":        hasMedia,
		"mediaType":       mediaType,
	}
	logger.LogInput(input)

	posts, err := h.postUseCase.FindManyPosts(userID, limit, offset, includeSubPosts, hasMedia, mediaType)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(posts, nil)
	return c.JSON(posts)
}
