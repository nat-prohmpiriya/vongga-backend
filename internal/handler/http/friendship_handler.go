package handler

import (
	"fmt"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FriendshipHandler struct {
	friendshipUseCase domain.FriendshipUseCase
}

func NewFriendshipHandler(router fiber.Router, fu domain.FriendshipUseCase) *FriendshipHandler {
	handler := &FriendshipHandler{
		friendshipUseCase: fu,
	}

	router.Post("/request/:userId", handler.SendFriendRequest)
	router.Post("/accept/:userId", handler.AcceptFriendRequest)
	router.Post("/reject/:userId", handler.RejectFriendRequest)
	router.Delete("/:userId", handler.RemoveFriend)
	router.Find("/", handler.FindManyFriends)
	router.Find("/requests", handler.FindManyFriendRequests)

	return handler
}

func (h *FriendshipHandler) SendFriendRequest(c *fiber.Ctx) error {
	logger := utils.NewLogger("FriendshipHandler.SendFriendRequest")

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	targetID, err := primitive.ObjectIDFromHex(c.Params("userId"))
	if err != nil {
		logger.LogOutput(nil, err)
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid target user ID")
	}

	if targetID == userID {
		logger.LogOutput(nil, fmt.Errorf("cannot send friend request to self"))
		return utils.SendError(c, fiber.StatusBadRequest, "Cannot send friend request to yourself")
	}

	logger.LogInput(userID, targetID)
	if err := h.friendshipUseCase.SendFriendRequest(userID, targetID); err != nil {
		logger.LogOutput(nil, err)
		return utils.HandleError(c, err)
	}

	logger.LogOutput("Friend request sent successfully", nil)
	return utils.SendSuccess(c, "Friend request sent successfully")
}

func (h *FriendshipHandler) AcceptFriendRequest(c *fiber.Ctx) error {
	logger := utils.NewLogger("FriendshipHandler.AcceptFriendRequest")

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	targetID, err := primitive.ObjectIDFromHex(c.Params("userId"))
	if err != nil {
		logger.LogOutput(nil, err)
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	logger.LogInput(userID, targetID)
	if err := h.friendshipUseCase.AcceptFriendRequest(userID, targetID); err != nil {
		logger.LogOutput(nil, err)
		return utils.HandleError(c, err)
	}

	logger.LogOutput("Friend request accepted successfully", nil)
	return utils.SendSuccess(c, "Friend request accepted successfully")
}

func (h *FriendshipHandler) RejectFriendRequest(c *fiber.Ctx) error {
	logger := utils.NewLogger("FriendshipHandler.RejectFriendRequest")

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	targetID, err := primitive.ObjectIDFromHex(c.Params("userId"))
	if err != nil {
		logger.LogOutput(nil, err)
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	logger.LogInput(userID, targetID)
	if err := h.friendshipUseCase.RejectFriendRequest(userID, targetID); err != nil {
		logger.LogOutput(nil, err)
		return utils.HandleError(c, err)
	}

	logger.LogOutput("Friend request rejected successfully", nil)
	return utils.SendSuccess(c, "Friend request rejected successfully")
}

func (h *FriendshipHandler) RemoveFriend(c *fiber.Ctx) error {
	logger := utils.NewLogger("FriendshipHandler.RemoveFriend")

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	targetID, err := primitive.ObjectIDFromHex(c.Params("userId"))
	if err != nil {
		logger.LogOutput(nil, err)
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	logger.LogInput(userID, targetID)
	if err := h.friendshipUseCase.RemoveFriend(userID, targetID); err != nil {
		logger.LogOutput(nil, err)
		return utils.HandleError(c, err)
	}

	logger.LogOutput("Friend removed successfully", nil)
	return utils.SendSuccess(c, "Friend removed successfully")
}

func (h *FriendshipHandler) FindManyFriends(c *fiber.Ctx) error {
	logger := utils.NewLogger("FriendshipHandler.FindManyFriends")

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	limit, offset := utils.FindPaginationParams(c)
	logger.LogInput(userID, limit, offset)

	friends, err := h.friendshipUseCase.FindManyFriends(userID, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return utils.HandleError(c, err)
	}

	logger.LogOutput(friends, nil)
	return c.JSON(friends)
}

func (h *FriendshipHandler) FindManyFriendRequests(c *fiber.Ctx) error {
	logger := utils.NewLogger("FriendshipHandler.FindManyFriendRequests")

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	limit, offset := utils.FindPaginationParams(c)
	logger.LogInput(userID, limit, offset)

	requests, err := h.friendshipUseCase.FindManyFriendRequests(userID, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return utils.HandleError(c, err)
	}

	logger.LogOutput(requests, nil)
	return c.JSON(requests)
}
