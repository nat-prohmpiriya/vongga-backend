package handler

import (
	"net/http"
	"strconv"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FollowHandler struct {
	followUseCase domain.FollowUseCase
}

func NewFollowHandler(router fiber.Router, fu domain.FollowUseCase) *FollowHandler {
	handler := &FollowHandler{
		followUseCase: fu,
	}

	router.Post("/:userId", handler.Follow)
	router.Delete("/:userId", handler.Unfollow)
	router.Get("/followers", handler.GetFollowers)
	router.Get("/following", handler.GetFollowing)
	router.Post("/block/:userId", handler.Block)
	router.Delete("/block/:userId", handler.Unblock)

	return handler
}

// Follow handles following a user
func (h *FollowHandler) Follow(c *fiber.Ctx) error {
	logger := utils.NewLogger("followHandler.Follow")

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	followingID := c.Params("userId")

	logger.LogInput(map[string]interface{}{
		"userID":      userID,
		"followingID": followingID,
	})

	followingObjID, err := primitive.ObjectIDFromHex(followingID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid following ID",
		})
	}

	err = h.followUseCase.Follow(userID, followingObjID)
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
func (h *FollowHandler) Unfollow(c *fiber.Ctx) error {
	logger := utils.NewLogger("followHandler.Unfollow")

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	followingID := c.Params("userId")

	logger.LogInput(map[string]interface{}{
		"userID":      userID,
		"followingID": followingID,
	})

	followingObjID, err := primitive.ObjectIDFromHex(followingID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid following ID",
		})
	}

	err = h.followUseCase.Unfollow(userID, followingObjID)
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
func (h *FollowHandler) Block(c *fiber.Ctx) error {
	logger := utils.NewLogger("followHandler.Block")

	userIDInterface := c.Locals("userId")
	if userIDInterface == nil {
		logger.LogOutput(nil, fiber.ErrUnauthorized)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		logger.LogOutput(nil, fiber.ErrUnauthorized)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid user authentication",
		})
	}

	blockedID := c.Params("userId")

	logger.LogInput(map[string]interface{}{
		"userID":    userID,
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
func (h *FollowHandler) Unblock(c *fiber.Ctx) error {
	logger := utils.NewLogger("followHandler.Unblock")

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	blockedID := c.Params("userId")

	logger.LogInput(map[string]interface{}{
		"userID":    userID,
		"blockedID": blockedID,
	})

	blockedObjID, err := primitive.ObjectIDFromHex(blockedID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid blocked user ID",
		})
	}

	err = h.followUseCase.Unblock(userID, blockedObjID)
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
func (h *FollowHandler) GetFollowers(c *fiber.Ctx) error {
	logger := utils.NewLogger("followHandler.GetFollowers")

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	logger.LogInput(map[string]interface{}{
		"userID": userID,
		"limit":  limit,
		"offset": offset,
	})

	followers, err := h.followUseCase.GetFollowers(userID, limit, offset)
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
func (h *FollowHandler) GetFollowing(c *fiber.Ctx) error {
	logger := utils.NewLogger("followHandler.GetFollowing")

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	logger.LogInput(map[string]interface{}{
		"userID": userID,
		"limit":  limit,
		"offset": offset,
	})

	following, err := h.followUseCase.GetFollowing(userID, limit, offset)
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
