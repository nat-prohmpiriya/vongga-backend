package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CommentHandler struct {
	commentUseCase domain.CommentUseCase
}

func NewCommentHandler(router fiber.Router, cu domain.CommentUseCase) {
	handler := &CommentHandler{
		commentUseCase: cu,
	}

	router.Post("/posts/:postId/comments", handler.CreateComment)
	router.Put("/comments/:id", handler.UpdateComment)
	router.Delete("/comments/:id", handler.DeleteComment)
	router.Get("/comments/:id", handler.GetComment)
	router.Get("/posts/:postId/comments", handler.ListComments)
}

type CreateCommentRequest struct {
	Content string         `json:"content"`
	Media   []domain.Media `json:"media,omitempty"`
	ReplyTo *string       `json:"replyTo,omitempty"`
}

func (h *CommentHandler) CreateComment(c *fiber.Ctx) error {
	logger := utils.NewLogger("CommentHandler.CreateComment")

	postID, err := primitive.ObjectIDFromHex(c.Params("postId"))
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	var req CreateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// TODO: Get userID from auth context
	userID := primitive.NewObjectID()

	var replyTo *primitive.ObjectID
	if req.ReplyTo != nil {
		replyToID, err := primitive.ObjectIDFromHex(*req.ReplyTo)
		if err != nil {
			logger.LogOutput(nil, err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid reply to ID",
			})
		}
		replyTo = &replyToID
	}

	input := map[string]interface{}{
		"postID":  postID,
		"userID":  userID,
		"request": req,
	}
	logger.LogInput(input)

	comment, err := h.commentUseCase.CreateComment(userID, postID, req.Content, req.Media, replyTo)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(comment, nil)
	return c.Status(fiber.StatusCreated).JSON(comment)
}

type UpdateCommentRequest struct {
	Content string         `json:"content"`
	Media   []domain.Media `json:"media,omitempty"`
}

func (h *CommentHandler) UpdateComment(c *fiber.Ctx) error {
	logger := utils.NewLogger("CommentHandler.UpdateComment")

	commentID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid comment ID",
		})
	}

	var req UpdateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	input := map[string]interface{}{
		"commentID": commentID,
		"request":   req,
	}
	logger.LogInput(input)

	comment, err := h.commentUseCase.UpdateComment(commentID, req.Content, req.Media)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(comment, nil)
	return c.JSON(comment)
}

func (h *CommentHandler) DeleteComment(c *fiber.Ctx) error {
	logger := utils.NewLogger("CommentHandler.DeleteComment")

	commentID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid comment ID",
		})
	}
	logger.LogInput(commentID)

	err = h.commentUseCase.DeleteComment(commentID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput("Comment deleted successfully", nil)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *CommentHandler) GetComment(c *fiber.Ctx) error {
	logger := utils.NewLogger("CommentHandler.GetComment")

	commentID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid comment ID",
		})
	}
	logger.LogInput(commentID)

	comment, err := h.commentUseCase.GetComment(commentID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(comment, nil)
	return c.JSON(comment)
}

func (h *CommentHandler) ListComments(c *fiber.Ctx) error {
	logger := utils.NewLogger("CommentHandler.ListComments")

	postID, err := primitive.ObjectIDFromHex(c.Params("postId"))
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	limit := c.QueryInt("limit", 0)
	offset := c.QueryInt("offset", 0)

	input := map[string]interface{}{
		"postID": postID,
		"limit":  limit,
		"offset": offset,
	}
	logger.LogInput(input)

	comments, err := h.commentUseCase.ListComments(postID, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(comments, nil)
	return c.JSON(comments)
}
