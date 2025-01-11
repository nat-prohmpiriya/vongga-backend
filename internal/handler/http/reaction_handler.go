package handler

import (
	"vongga-api/internal/domain"
	"vongga-api/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReactionHandler struct {
	reactionUseCase domain.ReactionUseCase
}

func NewReactionHandler(router fiber.Router, ru domain.ReactionUseCase) *ReactionHandler {
	handler := &ReactionHandler{
		reactionUseCase: ru,
	}

	router.Post("/", handler.CreateReaction)
	router.Delete("/:id", handler.DeleteReaction)
	router.Get("/post/:postId", handler.ListPostReactions)
	router.Get("/comment/:commentId", handler.ListCommentReactions)

	return handler
}

// CreateReaction creates a new reaction
// @Summary Create a new reaction
// @Description Create a new reaction on a post or comment
// @Tags reactions
// @Accept json
// @Produce json
// @Param reaction body domain.CreateReactionRequest true "Reaction details"
// @Security BearerAuth
// @Success 201 {object} domain.Reaction
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Router /reactions [post]
func (h *ReactionHandler) CreateReaction(c *fiber.Ctx) error {
	logger := utils.NewLogger("ReactionHandler.CreateReaction")

	userID, errr := utils.GetUserIDFromContext(c)
	if errr != nil {
		logger.LogOutput(nil, errr)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req domain.CreateReactionRequest
	if err := c.BodyParser(&req); err != nil {
		logger.LogInput(req)
		logger.LogOutput(nil, err)
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	logger.LogInput(userID, req)

	var commentID *primitive.ObjectID
	var postID primitive.ObjectID
	var err error

	if req.CommentID != "" {
		id, err := primitive.ObjectIDFromHex(req.CommentID)
		if err != nil {
			logger.LogOutput(nil, err)
			return utils.SendError(c, fiber.StatusBadRequest, "Invalid comment ID")
		}
		commentID = &id
	}

	if req.PostID != "" {
		postID, err = primitive.ObjectIDFromHex(req.PostID)
		if err != nil {
			logger.LogOutput(nil, err)
			return utils.SendError(c, fiber.StatusBadRequest, "Invalid post ID")
		}
	}

	reaction, err := h.reactionUseCase.CreateReaction(userID, postID, commentID, req.Type)
	if err != nil {
		logger.LogOutput(nil, err)
		return utils.HandleError(c, err)
	}

	logger.LogOutput(reaction, nil)
	return c.Status(fiber.StatusCreated).JSON(reaction)
}

// DeleteReaction deletes a reaction
// @Summary Delete a reaction
// @Description Delete a reaction by ID
// @Tags reactions
// @Accept json
// @Produce json
// @Param id path string true "Reaction ID"
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Router /reactions/{id} [delete]
func (h *ReactionHandler) DeleteReaction(c *fiber.Ctx) error {
	logger := utils.NewLogger("ReactionHandler.DeleteReaction")

	reactionID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		logger.LogInput(c.Params("id"))
		logger.LogOutput(nil, err)
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid reaction ID")
	}
	logger.LogInput(reactionID)

	if err := h.reactionUseCase.DeleteReaction(reactionID); err != nil {
		logger.LogOutput(nil, err)
		return utils.HandleError(c, err)
	}

	logger.LogOutput("Reaction deleted successfully", nil)
	return utils.SendSuccess(c, "Reaction deleted successfully")
}

// ListPostReactions lists reactions for a post
// @Summary List post reactions
// @Description Get a list of reactions for a specific post
// @Tags reactions
// @Accept json
// @Produce json
// @Param postId path string true "Post ID"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Security BearerAuth
// @Success 200 {array} domain.Reaction
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Router /reactions/post/{postId} [get]
func (h *ReactionHandler) ListPostReactions(c *fiber.Ctx) error {
	logger := utils.NewLogger("ReactionHandler.ListPostReactions")

	postID, err := primitive.ObjectIDFromHex(c.Params("postId"))
	if err != nil {
		logger.LogInput(c.Params("postId"))
		logger.LogOutput(nil, err)
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid post ID")
	}

	limit, offset := utils.GetPaginationParams(c)
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	logger.LogInput(postID, limit, offset)

	reactions, err := h.reactionUseCase.ListReactions(postID, false, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return utils.HandleError(c, err)
	}

	logger.LogOutput(reactions, nil)
	return c.JSON(reactions)
}

// ListCommentReactions lists reactions for a comment
// @Summary List comment reactions
// @Description Get a list of reactions for a specific comment
// @Tags reactions
// @Accept json
// @Produce json
// @Param commentId path string true "Comment ID"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Security BearerAuth
// @Success 200 {array} domain.Reaction
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Router /reactions/comment/{commentId} [get]
func (h *ReactionHandler) ListCommentReactions(c *fiber.Ctx) error {
	logger := utils.NewLogger("ReactionHandler.ListCommentReactions")

	commentID, err := primitive.ObjectIDFromHex(c.Params("commentId"))
	if err != nil {
		logger.LogInput(c.Params("commentId"))
		logger.LogOutput(nil, err)
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid comment ID")
	}

	limit, offset := utils.GetPaginationParams(c)
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	logger.LogInput(commentID, limit, offset)

	reactions, err := h.reactionUseCase.ListReactions(commentID, true, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return utils.HandleError(c, err)
	}

	logger.LogOutput(reactions, nil)
	return c.JSON(reactions)
}
