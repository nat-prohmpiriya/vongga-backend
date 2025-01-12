package handler

import (
	"strconv"

	"vongga_api/internal/domain"
	"vongga_api/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
)

type SubPostHandler struct {
	subPostUseCase domain.SubPostUseCase
	tracer         trace.Tracer
}

func NewSubPostHandler(router fiber.Router, su domain.SubPostUseCase, tracer trace.Tracer) {
	handler := &SubPostHandler{
		subPostUseCase: su,
		tracer:         tracer,
	}

	router.Post("/:postId/subposts", handler.CreateSubPost)
	router.Put("/subposts/:id", handler.UpdateSubPost)
	router.Delete("/subposts/:id", handler.DeleteSubPost)
	router.Get("/subposts/:id", handler.FindSubPost)
	router.Get("/:postId/subposts", handler.FindManySubPosts)
	router.Put("/:postId/subposts/reorder", handler.ReorderSubPosts)
}

type CreateSubPostRequest struct {
	Content string         `json:"content"`
	Media   []domain.Media `json:"media,omitempty"`
	Order   int            `json:"order"`
}

func (h *SubPostHandler) CreateSubPost(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "SubPostHandler.CreateSubPost")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	parentID, err := primitive.ObjectIDFromHex(c.Params("postId"))
	if err != nil {
		logger.Output("invalid post ID format 1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	var req CreateSubPostRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Output("invalid request body format 2", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// TODO: Find userID from auth context
	userID := primitive.NewObjectID()

	logger.Input(map[string]interface{}{
		"parentID": parentID,
		"userID":   userID,
		"request":  req,
	})

	subPost, err := h.subPostUseCase.CreateSubPost(ctx, parentID, userID, req.Content, req.Media, req.Order)
	if err != nil {
		logger.Output("failed to create subpost 3", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(subPost, nil)
	return c.JSON(subPost)
}

type UpdateSubPostRequest struct {
	Content string         `json:"content"`
	Media   []domain.Media `json:"media,omitempty"`
}

func (h *SubPostHandler) UpdateSubPost(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "SubPostHandler.UpdateSubPost")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	subPostID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		logger.Output("invalid subpost ID format 1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid subpost ID",
		})
	}

	var req UpdateSubPostRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Output("invalid request body format 2", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Input(map[string]interface{}{
		"subPostID": subPostID,
		"request":   req,
	})

	subPost, err := h.subPostUseCase.UpdateSubPost(ctx, subPostID, req.Content, req.Media)
	if err != nil {
		logger.Output("failed to update subpost 3", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(subPost, nil)
	return c.JSON(subPost)
}

func (h *SubPostHandler) DeleteSubPost(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "SubPostHandler.DeleteSubPost")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	subPostID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		logger.Output("invalid subpost ID format 1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid subpost ID",
		})
	}
	logger.Input(map[string]interface{}{
		"subPostID": subPostID,
	})

	err = h.subPostUseCase.DeleteSubPost(ctx, subPostID)
	if err != nil {
		logger.Output("failed to delete subpost 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output("subpost deleted successfully", nil)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *SubPostHandler) FindSubPost(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "SubPostHandler.FindSubPost")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	subPostID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		logger.Output("invalid subpost ID format 1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid subpost ID",
		})
	}
	logger.Input(map[string]interface{}{
		"subPostID": subPostID,
	})

	subPost, err := h.subPostUseCase.FindSubPost(ctx, subPostID)
	if err != nil {
		logger.Output("failed to find subpost 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(subPost, nil)
	return c.JSON(subPost)
}

func (h *SubPostHandler) FindManySubPosts(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "SubPostHandler.FindManySubPosts")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	parentID, err := primitive.ObjectIDFromHex(c.Params("postId"))
	if err != nil {
		logger.Output("invalid post ID format 1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	logger.Input(map[string]interface{}{
		"parentID": parentID,
		"limit":    limit,
		"offset":   offset,
	})

	subPosts, err := h.subPostUseCase.FindManySubPosts(ctx, parentID, limit, offset)
	if err != nil {
		logger.Output("failed to find subposts 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(subPosts, nil)
	return c.JSON(subPosts)
}

type ReorderSubPostsRequest struct {
	Orders map[string]int `json:"orders"` // subPostId -> order
}

func (h *SubPostHandler) ReorderSubPosts(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "SubPostHandler.ReorderSubPosts")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	parentID, err := primitive.ObjectIDFromHex(c.Params("postId"))
	if err != nil {
		logger.Output("invalid post ID format 1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	var req ReorderSubPostsRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Output("invalid request body format 2", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Convert string IDs to ObjectIDs
	orders := make(map[primitive.ObjectID]int)
	for idStr, order := range req.Orders {
		id, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			logger.Output("invalid subpost ID in orders 3", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid subpost ID in orders",
			})
		}
		orders[id] = order
	}

	logger.Input(map[string]interface{}{
		"parentID": parentID,
		"orders":   orders,
	})

	err = h.subPostUseCase.ReorderSubPosts(ctx, parentID, orders)
	if err != nil {
		logger.Output("failed to reorder subposts 4", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output("subposts reordered successfully", nil)
	return c.SendStatus(fiber.StatusNoContent)
}
