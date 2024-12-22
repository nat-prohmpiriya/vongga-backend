package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vongga/vongga-backend/domain"
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
