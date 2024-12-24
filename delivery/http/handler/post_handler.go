package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostHandler struct {
	postUseCase domain.PostUseCase
}

func NewPostHandler(router fiber.Router, pu domain.PostUseCase) {
	handler := &PostHandler{
		postUseCase: pu,
	}

	router.Post("/posts", handler.CreatePost)
	router.Put("/posts/:id", handler.UpdatePost)
	router.Delete("/posts/:id", handler.DeletePost)
	router.Get("/posts/:id", handler.GetPost)
	router.Get("/users/:userId/posts", handler.ListPosts)
}

type CreatePostRequest struct {
	Content    string          `json:"content"`
	Media      []domain.Media  `json:"media,omitempty"`
	Tags       []string        `json:"tags,omitempty"`
	Location   *domain.Location `json:"location,omitempty"`
	Visibility string          `json:"visibility"`
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
	logger.LogInput(req)

	// TODO: Get userID from auth context
	userID := primitive.NewObjectID()

	post, err := h.postUseCase.CreatePost(userID, req.Content, req.Media, req.Tags, req.Location, req.Visibility)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(post, nil)
	return c.Status(fiber.StatusCreated).JSON(post)
}

type UpdatePostRequest struct {
	Content    string          `json:"content"`
	Media      []domain.Media  `json:"media,omitempty"`
	Tags       []string        `json:"tags,omitempty"`
	Location   *domain.Location `json:"location,omitempty"`
	Visibility string          `json:"visibility"`
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

func (h *PostHandler) GetPost(c *fiber.Ctx) error {
	logger := utils.NewLogger("PostHandler.GetPost")

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

	post, err := h.postUseCase.GetPost(postID, includeSubPosts)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(post, nil)
	return c.JSON(post)
}

func (h *PostHandler) ListPosts(c *fiber.Ctx) error {
	logger := utils.NewLogger("PostHandler.ListPosts")

	userID, err := primitive.ObjectIDFromHex(c.Params("userId"))
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	limit := c.QueryInt("limit", 0)
	offset := c.QueryInt("offset", 0)
	includeSubPosts := c.Query("includeSubPosts") == "true"

	input := map[string]interface{}{
		"userID":          userID,
		"limit":           limit,
		"offset":          offset,
		"includeSubPosts": includeSubPosts,
	}
	logger.LogInput(input)

	posts, err := h.postUseCase.ListPosts(userID, limit, offset, includeSubPosts)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(posts, nil)
	return c.JSON(posts)
}
