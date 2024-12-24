package handler

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type followHandler struct {
	followUseCase domain.FollowUseCase
}

// NewFollowHandler creates a new instance of FollowHandler
func NewFollowHandler(fu domain.FollowUseCase) *followHandler {
	return &followHandler{
		followUseCase: fu,
	}
}

// Follow handles following a user
func (h *followHandler) Follow(c *fiber.Ctx) error {
	logger := utils.NewLogger("followHandler.Follow")

	userID := c.Locals("userID").(string)
	followingID := c.Params("id")

	logger.LogInput(map[string]interface{}{
		"userID":     userID,
		"followingID": followingID,
	})

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	followingObjID, err := primitive.ObjectIDFromHex(followingID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid following ID",
		})
	}

	err = h.followUseCase.Follow(userObjID, followingObjID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput("Successfully followed user", nil)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Successfully followed user",
	})
}

// Unfollow handles unfollowing a user
func (h *followHandler) Unfollow(c *fiber.Ctx) error {
	logger := utils.NewLogger("followHandler.Unfollow")

	userID := c.Locals("userID").(string)
	followingID := c.Params("id")

	logger.LogInput(map[string]interface{}{
		"userID":     userID,
		"followingID": followingID,
	})

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	followingObjID, err := primitive.ObjectIDFromHex(followingID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid following ID",
		})
	}

	err = h.followUseCase.Unfollow(userObjID, followingObjID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput("Successfully unfollowed user", nil)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Successfully unfollowed user",
	})
}

// Block handles blocking a user
func (h *followHandler) Block(c *fiber.Ctx) error {
	logger := utils.NewLogger("followHandler.Block")

	userID := c.Locals("userID").(string)
	blockedID := c.Params("id")

	logger.LogInput(map[string]interface{}{
		"userID":   userID,
		"blockedID": blockedID,
	})

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	blockedObjID, err := primitive.ObjectIDFromHex(blockedID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid blocked user ID",
		})
	}

	err = h.followUseCase.Block(userObjID, blockedObjID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput("Successfully blocked user", nil)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Successfully blocked user",
	})
}

// Unblock handles unblocking a user
func (h *followHandler) Unblock(c *fiber.Ctx) error {
	logger := utils.NewLogger("followHandler.Unblock")

	userID := c.Locals("userID").(string)
	blockedID := c.Params("id")

	logger.LogInput(map[string]interface{}{
		"userID":   userID,
		"blockedID": blockedID,
	})

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	blockedObjID, err := primitive.ObjectIDFromHex(blockedID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid blocked user ID",
		})
	}

	err = h.followUseCase.Unblock(userObjID, blockedObjID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput("Successfully unblocked user", nil)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Successfully unblocked user",
	})
}

// GetFollowers handles getting a user's followers
func (h *followHandler) GetFollowers(c *fiber.Ctx) error {
	logger := utils.NewLogger("followHandler.GetFollowers")

	userID := c.Params("id")
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	logger.LogInput(map[string]interface{}{
		"userID": userID,
		"limit":  limit,
		"offset": offset,
	})

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	followers, err := h.followUseCase.GetFollowers(userObjID, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get followers",
		})
	}

	logger.LogOutput(followers, nil)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"followers": followers,
	})
}

// GetFollowing handles getting users that a user is following
func (h *followHandler) GetFollowing(c *fiber.Ctx) error {
	logger := utils.NewLogger("followHandler.GetFollowing")

	userID := c.Params("id")
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	logger.LogInput(map[string]interface{}{
		"userID": userID,
		"limit":  limit,
		"offset": offset,
	})

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	following, err := h.followUseCase.GetFollowing(userObjID, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get following",
		})
	}

	logger.LogOutput(following, nil)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"following": following,
	})
}

// RegisterRoutes registers the follow routes
func (h *followHandler) RegisterRoutes(app *fiber.App) {
	followGroup := app.Group("/api/follows")

	followGroup.Post("/:id", h.Follow)
	followGroup.Delete("/:id", h.Unfollow)
	followGroup.Post("/:id/block", h.Block)
	followGroup.Delete("/:id/block", h.Unblock)
	followGroup.Get("/:id/followers", h.GetFollowers)
	followGroup.Get("/:id/following", h.GetFollowing)
}
