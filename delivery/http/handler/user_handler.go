package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"regexp"
	"time"
)

type UserHandler struct {
	userUseCase domain.UserUseCase
}

func NewUserHandler(userUseCase domain.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

func (h *UserHandler) CreateOrUpdateUser(c *fiber.Ctx) error {
	// Get Firebase user data from context (set by middleware)
	firebaseUID := c.Locals("firebase_uid").(string)
	email := c.Locals("email").(string)

	var req struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		PhotoURL  string `json:"photoUrl"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	user, err := h.userUseCase.CreateOrUpdateUser(
		firebaseUID,
		email,
		req.FirstName,
		req.LastName,
		req.PhotoURL,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"user": user,
	})
}

func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	firebaseUID := c.Locals("firebase_uid").(string)

	user, err := h.userUseCase.GetUserByFirebaseUID(firebaseUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"user": user,
	})
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	firebaseUID := c.Locals("firebase_uid").(string)

	var req struct {
		FirstName   *string `json:"firstName"`
		LastName    *string `json:"lastName"`
		PhotoURL    *string `json:"photoUrl"`
		DisplayName *string `json:"displayName"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	user, err := h.userUseCase.GetUserByFirebaseUID(firebaseUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Update only provided fields
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.PhotoURL != nil {
		user.PhotoURL = *req.PhotoURL
	}
	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}

	// Update timestamp
	user.UpdatedAt = time.Now()

	err = h.userUseCase.UpdateUser(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"user": user,
	})
}

func (h *UserHandler) CheckUsername(c *fiber.Ctx) error {
	username := c.Query("username")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username is required",
		})
	}

	// Check if username is valid (only alphanumeric characters)
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(username) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username can only contain letters and numbers",
			"available": false,
		})
	}

	// Check length
	if len(username) < 3 || len(username) > 15 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username must be between 3 and 15 characters",
			"available": false,
		})
	}

	user, err := h.userUseCase.GetUserByUsername(username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"available": user == nil,
	})
}
