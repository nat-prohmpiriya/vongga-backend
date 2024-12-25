package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CommentHandler struct {
	commentUseCase domain.CommentUseCase
	userUseCase    domain.UserUseCase
}

func NewCommentHandler(router fiber.Router, cu domain.CommentUseCase, uu domain.UserUseCase) *CommentHandler {
	handler := &CommentHandler{
		commentUseCase: cu,
		userUseCase:    uu,
	}

	router.Post("/posts/:postId", handler.CreateComment)
	router.Put("/:id", handler.UpdateComment)
	router.Delete("/:id", handler.DeleteComment)
	router.Get("/posts/:postId", handler.ListComments)
	router.Get("/:id", handler.GetComment)

	return handler
}

type CreateCommentRequest struct {
	Content string         `json:"content"`
	Media   []domain.Media `json:"media,omitempty"`
	ReplyTo *string        `json:"replyTo,omitempty"`
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

	// Get userID from auth context
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

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

	comments, err := h.commentUseCase.ListComments(postID, limit, offset)
	if err != nil {
		logger.LogOutput(input, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Create a slice to store comments with user information
	commentsWithUsers := make([]domain.CommentWithUser, 0, len(comments))

	// Fetch user information for each comment
	for _, comment := range comments {
		user, err := h.userUseCase.GetUserByID(comment.UserID.Hex())
		if err != nil {
			logger.LogOutput(input, err)
			continue
		}

		// Create a copy of the comment
		commentCopy := comment

		commentWithUser := domain.CommentWithUser{
			Comment: &commentCopy,
			User: &domain.CommentUser{
				ID:           user.ID,
				Username:     user.Username,
				DisplayName:  user.DisplayName,
				PhotoProfile: user.PhotoProfile,
				FirstName:    user.FirstName,
				LastName:     user.LastName,
			},
		}
		commentsWithUsers = append(commentsWithUsers, commentWithUser)
	}

	logger.LogOutput(commentsWithUsers, nil)
	return c.JSON(commentsWithUsers)
}
