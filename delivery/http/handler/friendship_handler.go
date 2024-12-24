package handler

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type friendshipHandler struct {
	friendshipUseCase domain.FriendshipUseCase
}

// NewFriendshipHandler creates a new instance of FriendshipHandler
func NewFriendshipHandler(fu domain.FriendshipUseCase) *friendshipHandler {
	return &friendshipHandler{
		friendshipUseCase: fu,
	}
}

// SendFriendRequest handles sending a friend request
func (h *friendshipHandler) SendFriendRequest(c *fiber.Ctx) error {
	logger := utils.NewLogger("FriendshipHandler.SendFriendRequest")

	userID := c.Locals("userID").(string)
	friendID := c.Params("id")

	logger.LogInput(map[string]interface{}{
		"userID":  userID,
		"friendID": friendID,
	})

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	friendObjID, err := primitive.ObjectIDFromHex(friendID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid friend ID",
		})
	}

	err = h.friendshipUseCase.SendFriendRequest(userObjID, friendObjID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput("Friend request sent successfully", nil)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Friend request sent successfully",
	})
}

// AcceptFriendRequest handles accepting a friend request
func (h *friendshipHandler) AcceptFriendRequest(c *fiber.Ctx) error {
	logger := utils.NewLogger("FriendshipHandler.AcceptFriendRequest")

	userID := c.Locals("userID").(string)
	friendID := c.Params("id")

	logger.LogInput(map[string]interface{}{
		"userID":  userID,
		"friendID": friendID,
	})

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	friendObjID, err := primitive.ObjectIDFromHex(friendID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid friend ID",
		})
	}

	err = h.friendshipUseCase.AcceptFriendRequest(userObjID, friendObjID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput("Friend request accepted", nil)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Friend request accepted",
	})
}

// RejectFriendRequest handles rejecting a friend request
func (h *friendshipHandler) RejectFriendRequest(c *fiber.Ctx) error {
	logger := utils.NewLogger("FriendshipHandler.RejectFriendRequest")

	userID := c.Locals("userID").(string)
	friendID := c.Params("id")

	logger.LogInput(map[string]interface{}{
		"userID":  userID,
		"friendID": friendID,
	})

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	friendObjID, err := primitive.ObjectIDFromHex(friendID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid friend ID",
		})
	}

	err = h.friendshipUseCase.RejectFriendRequest(userObjID, friendObjID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput("Friend request rejected", nil)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Friend request rejected",
	})
}

// CancelFriendRequest handles canceling a sent friend request
func (h *friendshipHandler) CancelFriendRequest(c *fiber.Ctx) error {
	logger := utils.NewLogger("FriendshipHandler.CancelFriendRequest")

	userID := c.Locals("userID").(string)
	friendID := c.Params("id")

	logger.LogInput(map[string]interface{}{
		"userID":  userID,
		"friendID": friendID,
	})

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	friendObjID, err := primitive.ObjectIDFromHex(friendID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid friend ID",
		})
	}

	err = h.friendshipUseCase.CancelFriendRequest(userObjID, friendObjID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput("Friend request canceled", nil)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Friend request canceled",
	})
}

// Unfriend handles removing a friend
func (h *friendshipHandler) Unfriend(c *fiber.Ctx) error {
	logger := utils.NewLogger("FriendshipHandler.Unfriend")

	userID := c.Locals("userID").(string)
	friendID := c.Params("id")

	logger.LogInput(map[string]interface{}{
		"userID":  userID,
		"friendID": friendID,
	})

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	friendObjID, err := primitive.ObjectIDFromHex(friendID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid friend ID",
		})
	}

	err = h.friendshipUseCase.Unfriend(userObjID, friendObjID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput("Successfully unfriended", nil)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Successfully unfriended",
	})
}

// GetFriends handles getting a user's friends
func (h *friendshipHandler) GetFriends(c *fiber.Ctx) error {
	logger := utils.NewLogger("FriendshipHandler.GetFriends")

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

	friends, err := h.friendshipUseCase.GetFriends(userObjID, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get friends",
		})
	}

	logger.LogOutput(friends, nil)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"friends": friends,
	})
}

// GetPendingRequests handles getting pending friend requests
func (h *friendshipHandler) GetPendingRequests(c *fiber.Ctx) error {
	logger := utils.NewLogger("FriendshipHandler.GetPendingRequests")

	userID := c.Locals("userID").(string)
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

	requests, err := h.friendshipUseCase.GetPendingRequests(userObjID, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get pending requests",
		})
	}

	logger.LogOutput(requests, nil)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"requests": requests,
	})
}

// RegisterRoutes registers the friendship routes
func (h *friendshipHandler) RegisterRoutes(app *fiber.App) {
	friendGroup := app.Group("/api/friends")

	friendGroup.Post("/:id", h.SendFriendRequest)
	friendGroup.Post("/:id/accept", h.AcceptFriendRequest)
	friendGroup.Post("/:id/reject", h.RejectFriendRequest)
	friendGroup.Delete("/:id/cancel", h.CancelFriendRequest)
	friendGroup.Delete("/:id", h.Unfriend)
	friendGroup.Get("/:id", h.GetFriends)
	friendGroup.Get("/requests", h.GetPendingRequests)
}
