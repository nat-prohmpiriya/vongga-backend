package handler

import (
	"fmt"

	"vongga_api/internal/domain"
	"vongga_api/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
)

type PostHandler struct {
	postUseCase domain.PostUseCase
	tracer      trace.Tracer
}

func NewPostHandler(router fiber.Router, pu domain.PostUseCase, tracer trace.Tracer) *PostHandler {
	handler := &PostHandler{
		postUseCase: pu,
		tracer:      tracer,
	}

	router.Post("", handler.CreatePost)
	router.Get("", handler.FindManyPosts)
	router.Get("/:id", handler.FindPost)
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
	ctx, span := h.tracer.Start(c.UserContext(), "PostHandler.CreatePost")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	var req CreatePostRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Output("error parsing request body 1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 2", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	post, err := h.postUseCase.CreatePost(
		ctx,
		userID,
		req.Content,
		req.Media,
		req.Tags,
		req.Location,
		req.Visibility,
		req.SubPosts,
	)
	if err != nil {
		logger.Output("error creating post 3", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(post, nil)
	return c.Status(fiber.StatusCreated).JSON(post)
}

func (h *PostHandler) UpdatePost(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "PostHandler.UpdatePost")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	postID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		logger.Output("error parsing post ID 1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	var req UpdatePostRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Output("error parsing request body 2", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	input := map[string]interface{}{
		"postID":  postID,
		"request": req,
	}
	logger.Input(input)

	post, err := h.postUseCase.UpdatePost(ctx, postID, req.Content, req.Media, req.Tags, req.Location, req.Visibility)
	if err != nil {
		logger.Output("error updating post 3", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(post, nil)
	return c.JSON(post)
}

func (h *PostHandler) DeletePost(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "PostHandler.DeletePost")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	postID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		logger.Output("error parsing post ID 1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}
	logger.Input(postID)

	err = h.postUseCase.DeletePost(ctx, postID)
	if err != nil {
		logger.Output("error deleting post 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output("Post deleted successfully", nil)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *PostHandler) FindPost(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "PostHandler.FindPost")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	postID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		logger.Output("error parsing post ID 1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	includeSubPosts := c.Query("includeSubPosts") == "true"
	input := map[string]interface{}{
		"postID":          postID,
		"includeSubPosts": includeSubPosts,
	}
	logger.Input(input)

	post, err := h.postUseCase.FindPost(ctx, postID, includeSubPosts)
	if err != nil {
		logger.Output("error finding post 2", err)
		if domain.IsNotFoundError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(post, nil)
	return c.JSON(post)
}

func (h *PostHandler) FindManyPosts(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "PostHandler.FindManyPosts")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userIDStr := c.Query("userId")
	if userIDStr == "" {
		logger.Output("missing user ID query parameter 1", fmt.Errorf("missing userId query parameter"))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing user ID",
		})
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		logger.Output("error parsing user ID 2", err)
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
	logger.Input(input)

	posts, err := h.postUseCase.FindManyPosts(ctx, userID, limit, offset, includeSubPosts, hasMedia, mediaType)
	if err != nil {
		logger.Output("error finding many posts 3", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(posts, nil)
	return c.JSON(posts)
}
