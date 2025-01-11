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
	router.Get("/subposts/:id", handler.GetSubPost)
	router.Get("/:postId/subposts", handler.ListSubPosts)
	router.Put("/:postId/subposts/reorder", handler.ReorderSubPosts)
}

type CreateSubPostRequest struct {
	Content string         `json:"content"`
	Media   []domain.Media `json:"media,omitempty"`
	Order   int            `json:"order"`
}

func (h *SubPostHandler) CreateSubPost(c *fiber.Ctx) error {
	logger := utils.NewLogger("SubPostHandler.CreateSubPost")

	parentID, err := primitive.ObjectIDFromHex(c.Params("postId"))
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	var req CreateSubPostRequest
	if err := c.BodyParser(&req); err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// TODO: Get userID from auth context
	userID := primitive.NewObjectID()

	input := map[string]interface{}{
		"parentID": parentID,
		"userID":   userID,
		"request":  req,
	}
	logger.LogInput(input)

	subPost, err := h.subPostUseCase.CreateSubPost(parentID, userID, req.Content, req.Media, req.Order)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(subPost, nil)
	return c.JSON(subPost)
}

type UpdateSubPostRequest struct {
	Content string         `json:"content"`
	Media   []domain.Media `json:"media,omitempty"`
}

func (h *SubPostHandler) UpdateSubPost(c *fiber.Ctx) error {
	logger := utils.NewLogger("SubPostHandler.UpdateSubPost")

	subPostID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid subpost ID",
		})
	}

	var req UpdateSubPostRequest
	if err := c.BodyParser(&req); err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	input := map[string]interface{}{
		"subPostID": subPostID,
		"request":   req,
	}
	logger.LogInput(input)

	subPost, err := h.subPostUseCase.UpdateSubPost(subPostID, req.Content, req.Media)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(subPost, nil)
	return c.JSON(subPost)
}

func (h *SubPostHandler) DeleteSubPost(c *fiber.Ctx) error {
	logger := utils.NewLogger("SubPostHandler.DeleteSubPost")

	subPostID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid subpost ID",
		})
	}
	logger.LogInput(subPostID)

	err = h.subPostUseCase.DeleteSubPost(subPostID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput("SubPost deleted successfully", nil)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *SubPostHandler) GetSubPost(c *fiber.Ctx) error {
	logger := utils.NewLogger("SubPostHandler.GetSubPost")

	subPostID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid subpost ID",
		})
	}
	logger.LogInput(subPostID)

	subPost, err := h.subPostUseCase.GetSubPost(subPostID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(subPost, nil)
	return c.JSON(subPost)
}

func (h *SubPostHandler) ListSubPosts(c *fiber.Ctx) error {
	logger := utils.NewLogger("SubPostHandler.ListSubPosts")

	parentID, err := primitive.ObjectIDFromHex(c.Params("postId"))
	if err != nil {
		logger.LogOutput(nil, err)
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
	logger.LogInput(input)

	subPosts, err := h.subPostUseCase.ListSubPosts(parentID, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(subPosts, nil)
	return c.JSON(subPosts)
}

type ReorderSubPostsRequest struct {
	Orders map[string]int `json:"orders"` // subPostId -> order
}

func (h *SubPostHandler) ReorderSubPosts(c *fiber.Ctx) error {
	logger := utils.NewLogger("SubPostHandler.ReorderSubPosts")

	parentID, err := primitive.ObjectIDFromHex(c.Params("postId"))
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post ID",
		})
	}

	var req ReorderSubPostsRequest
	if err := c.BodyParser(&req); err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Convert string IDs to ObjectIDs
	orders := make(map[primitive.ObjectID]int)
	for idStr, order := range req.Orders {
		id, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			logger.LogOutput(nil, err)
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
	logger.LogInput(input)

	err = h.subPostUseCase.ReorderSubPosts(parentID, orders)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput("SubPosts reordered successfully", nil)
	return c.SendStatus(fiber.StatusNoContent)
}
