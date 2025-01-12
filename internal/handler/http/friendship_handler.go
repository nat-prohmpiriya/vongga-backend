package handler

import (
	"fmt"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
)

type FriendshipHandler struct {
	friendshipUseCase domain.FriendshipUseCase
	tracer            trace.Tracer
}

func NewFriendshipHandler(router fiber.Router, fu domain.FriendshipUseCase, tracer trace.Tracer) *FriendshipHandler {
	handler := &FriendshipHandler{
		friendshipUseCase: fu,
		tracer:            tracer,
	}

	router.Post("/request/:userId", handler.SendFriendRequest)
	router.Post("/accept/:userId", handler.AcceptFriendRequest)
	router.Post("/reject/:userId", handler.RejectFriendRequest)
	router.Delete("/:userId", handler.RemoveFriend)
	router.Get("/", handler.FindManyFriends)
	router.Get("/requests", handler.FindManyFriendRequests)

	return handler
}

func (h *FriendshipHandler) SendFriendRequest(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "FriendshipHandler.SendFriendRequest")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	targetID, err := primitive.ObjectIDFromHex(c.Params("userId"))
	if err != nil {
		logger.Output("error parsing target ID 2", err)
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid target user ID")
	}

	if targetID == userID {
		logger.Output("cannot send friend request to self 3", fmt.Errorf("cannot send friend request to self"))
		return utils.SendError(c, fiber.StatusBadRequest, "Cannot send friend request to yourself")
	}

	logger.Input(map[string]interface{}{
		"userID":   userID,
		"targetID": targetID,
	})
	if err := h.friendshipUseCase.SendFriendRequest(ctx, userID, targetID); err != nil {
		logger.Output("error sending friend request 4", err)
		return utils.HandleError(c, err)
	}

	logger.Output("Friend request sent successfully", nil)
	return utils.SendSuccess(c, "Friend request sent successfully")
}

func (h *FriendshipHandler) AcceptFriendRequest(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "FriendshipHandler.AcceptFriendRequest")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	targetID, err := primitive.ObjectIDFromHex(c.Params("userId"))
	if err != nil {
		logger.Output("error parsing target ID 2", err)
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	logger.Input(map[string]interface{}{
		"userID":   userID,
		"targetID": targetID,
	})
	if err := h.friendshipUseCase.AcceptFriendRequest(ctx, userID, targetID); err != nil {
		logger.Output("error accepting friend request 3", err)
		return utils.HandleError(c, err)
	}

	logger.Output("Friend request accepted successfully", nil)
	return utils.SendSuccess(c, "Friend request accepted successfully")
}

func (h *FriendshipHandler) RejectFriendRequest(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "FriendshipHandler.RejectFriendRequest")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	targetID, err := primitive.ObjectIDFromHex(c.Params("userId"))
	if err != nil {
		logger.Output("error parsing target ID 2", err)
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	logger.Input(map[string]interface{}{
		"userID":   userID,
		"targetID": targetID,
	})
	if err := h.friendshipUseCase.RejectFriendRequest(ctx, userID, targetID); err != nil {
		logger.Output("error rejecting friend request 3", err)
		return utils.HandleError(c, err)
	}

	logger.Output("Friend request rejected successfully", nil)
	return utils.SendSuccess(c, "Friend request rejected successfully")
}

func (h *FriendshipHandler) RemoveFriend(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "FriendshipHandler.RemoveFriend")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	targetID, err := primitive.ObjectIDFromHex(c.Params("userId"))
	if err != nil {
		logger.Output("error parsing target ID 2", err)
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	logger.Input(map[string]interface{}{
		"userID":   userID,
		"targetID": targetID,
	})
	if err := h.friendshipUseCase.RemoveFriend(ctx, userID, targetID); err != nil {
		logger.Output("error removing friend 3", err)
		return utils.HandleError(c, err)
	}

	logger.Output("Friend removed successfully", nil)
	return utils.SendSuccess(c, "Friend removed successfully")
}

func (h *FriendshipHandler) FindManyFriends(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "FriendshipHandler.FindManyFriends")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	limit, offset := utils.FindPaginationParams(c)
	logger.Input(map[string]interface{}{
		"userID": userID,
		"limit":  limit,
		"offset": offset,
	})

	friends, err := h.friendshipUseCase.FindManyFriends(ctx, userID, limit, offset)
	if err != nil {
		logger.Output("error finding friends 2", err)
		return utils.HandleError(c, err)
	}

	logger.Output(friends, nil)
	return c.JSON(friends)
}

func (h *FriendshipHandler) FindManyFriendRequests(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.UserContext(), "FriendshipHandler.FindManyFriendRequests")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	limit, offset := utils.FindPaginationParams(c)
	logger.Input(map[string]interface{}{
		"userID": userID,
		"limit":  limit,
		"offset": offset,
	})

	requests, err := h.friendshipUseCase.FindManyFriendRequests(ctx, userID, limit, offset)
	if err != nil {
		logger.Output("error finding friend requests 2", err)
		return utils.HandleError(c, err)
	}

	logger.Output(requests, nil)
	return c.JSON(requests)
}
