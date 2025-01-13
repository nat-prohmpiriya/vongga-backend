package handler

import (
	"net/http"
	"strconv"

	"vongga_api/internal/domain"
	"vongga_api/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
)

type FollowHandler struct {
	followUseCase domain.FollowUseCase
	tracer        trace.Tracer
}

func NewFollowHandler(router fiber.Router, fu domain.FollowUseCase, tracer trace.Tracer) *FollowHandler {
	handler := &FollowHandler{
		followUseCase: fu,
		tracer:        tracer,
	}

	router.Post("/:userId", handler.Follow)
	router.Delete("/:userId", handler.Unfollow)
	router.Get("/followers", handler.FindFollowers)
	router.Get("/following", handler.FindFollowing)
	router.Post("/block/:userId", handler.Block)
	router.Delete("/block/:userId", handler.Unblock)

	return handler
}

// Follow handles following a user
func (h *FollowHandler) Follow(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "FollowHandler.Follow")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	followingID := c.Params("userId")

	logger.Input(map[string]interface{}{
		"userID":      userID,
		"followingID": followingID,
	})

	followingObjID, err := primitive.ObjectIDFromHex(followingID)
	if err != nil {
		logger.Output("error parsing following ID 2", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid following ID",
		})
	}

	err = h.followUseCase.Follow(ctx, userID, followingObjID)
	if err != nil {
		logger.Output("error following user 3", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output("Successfully followed user", nil)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Successfully followed user",
	})
}

// Unfollow handles unfollowing a user
func (h *FollowHandler) Unfollow(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "FollowHandler.Unfollow")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	followingID := c.Params("userId")

	logger.Input(map[string]interface{}{
		"userID":      userID,
		"followingID": followingID,
	})

	followingObjID, err := primitive.ObjectIDFromHex(followingID)
	if err != nil {
		logger.Output("error parsing following ID 2", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid following ID",
		})
	}

	err = h.followUseCase.Unfollow(ctx, userID, followingObjID)
	if err != nil {
		logger.Output("error unfollowing user 3", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output("Successfully unfollowed user", nil)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Successfully unfollowed user",
	})
}

// Block handles blocking a user
func (h *FollowHandler) Block(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "FollowHandler.Block")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userIDInterface := c.Locals("userId")
	if userIDInterface == nil {
		logger.Output(nil, fiber.ErrUnauthorized)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		logger.Output(nil, fiber.ErrUnauthorized)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid user authentication",
		})
	}

	blockedID := c.Params("userId")

	logger.Input(map[string]interface{}{
		"userID":    userID,
		"blockedID": blockedID,
	})

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logger.Output("error parsing user ID 3", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	blockedObjID, err := primitive.ObjectIDFromHex(blockedID)
	if err != nil {
		logger.Output("error parsing blocked user ID 4", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid blocked user ID",
		})
	}

	err = h.followUseCase.Block(ctx, userObjID, blockedObjID)
	if err != nil {
		logger.Output("error blocking user 5", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output("Successfully blocked user", nil)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Successfully blocked user",
	})
}

// Unblock handles unblocking a user
func (h *FollowHandler) Unblock(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "FollowHandler.Unblock")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	blockedID := c.Params("userId")

	logger.Input(map[string]interface{}{
		"userID":    userID,
		"blockedID": blockedID,
	})

	blockedObjID, err := primitive.ObjectIDFromHex(blockedID)
	if err != nil {
		logger.Output("error parsing blocked user ID 2", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid blocked user ID",
		})
	}

	err = h.followUseCase.Unblock(ctx, userID, blockedObjID)
	if err != nil {
		logger.Output("error unblocking user 3", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output("Successfully unblocked user", nil)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Successfully unblocked user",
	})
}

// FindFollowers handles getting a user's followers
func (h *FollowHandler) FindFollowers(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "FollowHandler.FindFollowers")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	logger.Input(map[string]interface{}{
		"userID": userID,
		"limit":  limit,
		"offset": offset,
	})

	followers, err := h.followUseCase.FindFollowers(ctx, userID, limit, offset)
	if err != nil {
		logger.Output("error finding followers 2", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get followers",
		})
	}

	logger.Output(followers, nil)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"followers": followers,
	})
}

// FindFollowing handles getting users that a user is following
func (h *FollowHandler) FindFollowing(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "FollowHandler.FindFollowing")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("error finding user ID 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	logger.Input(map[string]interface{}{
		"userID": userID,
		"limit":  limit,
		"offset": offset,
	})

	following, err := h.followUseCase.FindFollowing(ctx, userID, limit, offset)
	if err != nil {
		logger.Output("error finding following 2", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get following",
		})
	}

	logger.Output(following, nil)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"following": following,
	})
}
