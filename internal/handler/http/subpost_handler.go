package handler

import (
	"strconv"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubPostHandler struct {
	subPostUseCase domain.SubPostUseCase
}

func NewSubPostHandler(router fiber.Router, su domain.SubPostUseCase) {
	handler := &SubPostHandler{
		subPostUseCase: su,
	}

	router.Post("/:postId/subposts", handler.CreateSubPost)
	router.Put("/subposts/:id", handler.UpdateSubPost)
	router.Delete("/subposts/:id", handler.DeleteSubPost)
	router.Find("/subposts/:id", handler.FindSubPost)
	router.Find("/:postId/subposts", handler.FindManySubPosts)
	router.Put("/:postId/subposts/reorder", handler.ReorderSubPosts)
}

type CreateSubPostRequest struct {
	Content string         `json:"content"`
	Media   []domain.Media `json:"media,omitempty"`
	Order   int            `json:"order"`
}

func (h *SubPostHandler) CreateSubPost(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("SubPostHandler.CreateSubPost")

	parentID, err := primitive.ObjectIDFromHex(c.Params("postId"))
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	var req CreateSubPostRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// TODO: Find userID from auth context
	userID := primitive.NewObjectID()

	input := map[string]interface{}{
		"parentID": parentID,
		"userID":   userID,
		"request":  req,
	}
	logger.Input(input)

	subPost, err := h.subPostUseCase.CreateSubPost(parentID, userID, req.Content, req.Media, req.Order)
	if err != nil {
		logger.Output(nil, err)
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
	logger := utils.NewTraceLogger("SubPostHandler.UpdateSubPost")

	subPostID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid subpost ID",
		})
	}

	var req UpdateSubPostRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	input := map[string]interface{}{
		"subPostID": subPostID,
		"request":   req,
	}
	logger.Input(input)

	subPost, err := h.subPostUseCase.UpdateSubPost(subPostID, req.Content, req.Media)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(subPost, nil)
	return c.JSON(subPost)
}

func (h *SubPostHandler) DeleteSubPost(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("SubPostHandler.DeleteSubPost")

	subPostID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid subpost ID",
		})
	}
	logger.Input(subPostID)

	err = h.subPostUseCase.DeleteSubPost(subPostID)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output("SubPost deleted successfully", nil)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *SubPostHandler) FindSubPost(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("SubPostHandler.FindSubPost")

	subPostID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid subpost ID",
		})
	}
	logger.Input(subPostID)

	subPost, err := h.subPostUseCase.FindSubPost(subPostID)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(subPost, nil)
	return c.JSON(subPost)
}

func (h *SubPostHandler) FindManySubPosts(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("SubPostHandler.FindManySubPosts")

	parentID, err := primitive.ObjectIDFromHex(c.Params("postId"))
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	input := map[string]interface{}{
		"parentID": parentID,
		"limit":    limit,
		"offset":   offset,
	}
	logger.Input(input)

	subPosts, err := h.subPostUseCase.FindManySubPosts(parentID, limit, offset)
	if err != nil {
		logger.Output(nil, err)
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
	logger := utils.NewTraceLogger("SubPostHandler.ReorderSubPosts")

	parentID, err := primitive.ObjectIDFromHex(c.Params("postId"))
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	var req ReorderSubPostsRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Convert string IDs to ObjectIDs
	orders := make(map[primitive.ObjectID]int)
	for idStr, order := range req.Orders {
		id, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			logger.Output(nil, err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid subpost ID in orders",
			})
		}
		orders[id] = order
	}

	input := map[string]interface{}{
		"parentID": parentID,
		"orders":   orders,
	}
	logger.Input(input)

	err = h.subPostUseCase.ReorderSubPosts(parentID, orders)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output("SubPosts reordered successfully", nil)
	return c.SendStatus(fiber.StatusNoContent)
}
