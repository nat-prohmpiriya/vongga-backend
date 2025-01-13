package handler

import (
	"vongga_api/internal/domain"
	"vongga_api/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
)

type ReactionHandler struct {
	reactionUseCase domain.ReactionUseCase
	tracer          trace.Tracer
}

func NewReactionHandler(router fiber.Router, ru domain.ReactionUseCase, tracer trace.Tracer) *ReactionHandler {
	handler := &ReactionHandler{
		reactionUseCase: ru,
		tracer:          tracer,
	}

	router.Post("", handler.CreateReaction)
	router.Delete("/:id", handler.DeleteReaction)
	router.Get("/post/:postId", handler.FindManyPostReactions)
	router.Get("/comment/:commentId", handler.FindManyCommentReactions)

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
	ctx, span := h.tracer.Start(c.Context(), "ReactionHandler.CreateReaction")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, errr := utils.FindUserIDFromContext(c)
	if errr != nil {
		logger.Output("unauthorized access attempt 1", errr)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req domain.CreateReactionRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Input(req)
		logger.Output("invalid request body format 2", err)
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	logger.Input(map[string]interface{}{
		"userID":  userID,
		"request": req,
	})

	var commentID *primitive.ObjectID
	var postID primitive.ObjectID
	var err error

	if req.CommentID != "" {
		id, err := primitive.ObjectIDFromHex(req.CommentID)
		if err != nil {
			logger.Output("invalid comment ID format 3", err)
			return utils.SendError(c, fiber.StatusBadRequest, "Invalid comment ID")
		}
		commentID = &id
	}

	if req.PostID != "" {
		postID, err = primitive.ObjectIDFromHex(req.PostID)
		if err != nil {
			logger.Output("invalid post ID format 4", err)
			return utils.SendError(c, fiber.StatusBadRequest, "Invalid post ID")
		}
	}

	reaction, err := h.reactionUseCase.CreateReaction(ctx, userID, postID, commentID, req.Type)
	if err != nil {
		logger.Output("failed to create reaction 5", err)
		return utils.HandleError(c, err)
	}

	logger.Output(reaction, nil)
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
	ctx, span := h.tracer.Start(c.Context(), "ReactionHandler.DeleteReaction")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	reactionID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		logger.Input(c.Params("id"))
		logger.Output("invalid reaction ID format 1", err)
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid reaction ID")
	}
	logger.Input(reactionID)

	if err := h.reactionUseCase.DeleteReaction(ctx, reactionID); err != nil {
		logger.Output("failed to delete reaction 2", err)
		return utils.HandleError(c, err)
	}

	logger.Output("Reaction deleted successfully", nil)
	return utils.SendSuccess(c, "Reaction deleted successfully")
}

// FindManyPostReactions lists reactions for a post
// @Summary FindMany post reactions
// @Description Find a list of reactions for a specific post
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
func (h *ReactionHandler) FindManyPostReactions(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "ReactionHandler.FindManyPostReactions")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	postID, err := primitive.ObjectIDFromHex(c.Params("postId"))
	if err != nil {
		logger.Input(c.Params("postId"))
		logger.Output("invalid post ID format 1", err)
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid post ID")
	}

	limit, offset := utils.FindPaginationParams(c)
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	logger.Input(map[string]interface{}{
		"postId": postID,
		"limit":  limit,
		"offset": offset,
	})

	reactions, err := h.reactionUseCase.FindManyReactions(ctx, postID, false, limit, offset)
	if err != nil {
		logger.Output("failed to find post reactions 2", err)
		return utils.HandleError(c, err)
	}

	logger.Output(reactions, nil)
	return c.JSON(reactions)
}

// FindManyCommentReactions lists reactions for a comment
// @Summary FindMany comment reactions
// @Description Find a list of reactions for a specific comment
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
func (h *ReactionHandler) FindManyCommentReactions(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "ReactionHandler.FindManyCommentReactions")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	commentID, err := primitive.ObjectIDFromHex(c.Params("commentId"))
	if err != nil {
		logger.Input(c.Params("commentId"))
		logger.Output("invalid comment ID format 1", err)
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid comment ID")
	}

	limit, offset := utils.FindPaginationParams(c)
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	logger.Input(map[string]interface{}{
		"commentId": commentID,
		"limit":     limit,
		"offset":    offset,
	})

	reactions, err := h.reactionUseCase.FindManyReactions(ctx, commentID, true, limit, offset)
	if err != nil {
		logger.Output("failed to find comment reactions 2", err)
		return utils.HandleError(c, err)
	}

	logger.Output(reactions, nil)
	return c.JSON(reactions)
}
