package handler

import (
	"regexp"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
)

type UserHandler struct {
	userUseCase domain.UserUseCase
}

func NewUserHandler(router fiber.Router, userUseCase domain.UserUseCase) *UserHandler {
	handler := &UserHandler{
		userUseCase: userUseCase,
	}

	router.Post("/", handler.CreateOrUpdateUser)
	router.Get("/profile", handler.GetProfile)
	router.Patch("/", handler.UpdateUser)
	router.Get("/:username", handler.GetUserByUsername)
	router.Get("/check-username", handler.CheckUsername)
	router.Delete("/", handler.DeleteAccount)

	return handler
}

func (h *UserHandler) CreateOrUpdateUser(c *fiber.Ctx) error {
	logger := utils.NewLogger("UserHandler.CreateOrUpdateUser")
	
	// Get Firebase user data from context (set by middleware)
	userID := c.Locals("user_id").(string)
	email := c.Locals("email").(string)

	var req struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		PhotoURL  string `json:"photoUrl"`
	}

	if err := c.BodyParser(&req); err != nil {
		logger.LogInput(req)
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	logger.LogInput(userID, email, req)
	user, err := h.userUseCase.CreateOrUpdateUser(
		userID,
		email,
		req.FirstName,
		req.LastName,
		req.PhotoURL,
	)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(user, nil)
	return c.JSON(fiber.Map{
		"user": user,
	})
}

func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	logger := utils.NewLogger("UserHandler.GetProfile")
	
	userID := c.Locals("user_id").(string)
	logger.LogInput(userID)

	user, err := h.userUseCase.GetUserByID(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(user, nil)
	return c.JSON(fiber.Map{
		"user": user,
	})
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	logger := utils.NewLogger("UserHandler.UpdateUser")
	
	userID := c.Locals("user_id").(string)

	var req struct {
		FirstName      *string    `json:"firstName"`
		LastName       *string    `json:"lastName"`
		Username       *string    `json:"username"`
		DisplayName    *string    `json:"displayName"`
		Bio            *string    `json:"bio"`
		Avatar         *string    `json:"avatar"`
		PhotoProfile   *string    `json:"photoProfile"`
		PhotoCover     *string    `json:"photoCover"`
		DateOfBirth    *time.Time `json:"dateOfBirth"`
		Gender         *string    `json:"gender"`
		InterestedIn   []string   `json:"interestedIn"`
		Location       *domain.GeoLocation `json:"location"`
		RelationStatus *string    `json:"relationStatus"`
		Height         *float64   `json:"height"`
		Interests      []string   `json:"interests"`
		Occupation     *string    `json:"occupation"`
		Education      *string    `json:"education"`
		PhoneNumber    *string    `json:"phoneNumber"`
		DatingPhotos   []domain.DatingPhoto `json:"datingPhotos"`
		IsVerified     *bool      `json:"isVerified"`
		IsActive       *bool      `json:"isActive"`
		Live           *domain.Live `json:"live"`
	}

	if err := c.BodyParser(&req); err != nil {
		logger.LogInput(req)
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	logger.LogInput(userID, req)
	user, err := h.userUseCase.GetUserByID(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Username validation if provided
	if req.Username != nil {
		if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(*req.Username) {
			err := fiber.NewError(fiber.StatusBadRequest, "username can only contain letters and numbers")
			logger.LogOutput(nil, err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if len(*req.Username) < 3 || len(*req.Username) > 15 {
			err := fiber.NewError(fiber.StatusBadRequest, "username must be between 3 and 15 characters")
			logger.LogOutput(nil, err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		// Check if username is already taken by another user
		existingUser, err := h.userUseCase.GetUserByUsername(*req.Username)
		if err != nil {
			logger.LogOutput(nil, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if existingUser != nil && existingUser.ID != user.ID {
			err := fiber.NewError(fiber.StatusBadRequest, "username is already taken")
			logger.LogOutput(nil, err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		user.Username = *req.Username
	}

	// Update only provided fields
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
	}
	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}
	if req.PhotoProfile != nil {
		user.PhotoProfile = *req.PhotoProfile
	}
	if req.PhotoCover != nil {
		user.PhotoCover = *req.PhotoCover
	}
	if req.DateOfBirth != nil {
		user.DateOfBirth = *req.DateOfBirth
	}
	if req.Gender != nil {
		user.Gender = *req.Gender
	}
	if req.InterestedIn != nil {
		user.InterestedIn = req.InterestedIn
	}
	if req.Location != nil {
		user.Location = *req.Location
	}
	if req.RelationStatus != nil {
		user.RelationStatus = *req.RelationStatus
	}
	if req.Height != nil {
		user.Height = *req.Height
	}
	if req.Interests != nil {
		user.Interests = req.Interests
	}
	if req.Occupation != nil {
		user.Occupation = *req.Occupation
	}
	if req.Education != nil {
		user.Education = *req.Education
	}
	if req.PhoneNumber != nil {
		user.PhoneNumber = *req.PhoneNumber
	}
	if req.DatingPhotos != nil {
		user.DatingPhotos = req.DatingPhotos
	}
	if req.IsVerified != nil {
		user.IsVerified = *req.IsVerified
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.Live != nil {
		user.Live = *req.Live
	}

	// Update timestamp and version
	user.UpdatedAt = time.Now()
	user.Version++

	err = h.userUseCase.UpdateUser(user)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(user, nil)
	return c.JSON(fiber.Map{
		"user": user,
	})
}

func (h *UserHandler) GetUserByUsername(c *fiber.Ctx) error {
	logger := utils.NewLogger("UserHandler.GetUserByUsername")
	
	username := c.Params("username")
	if username == "" {
		err := fiber.NewError(fiber.StatusBadRequest, "username is required")
		logger.LogInput(username)
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogInput(username)
	user, err := h.userUseCase.GetUserByUsername(username)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if user == nil {
		err := fiber.NewError(fiber.StatusNotFound, "user not found")
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(user, nil)
	return c.JSON(fiber.Map{
		"user": user,
	})
}

func (h *UserHandler) CheckUsername(c *fiber.Ctx) error {
	logger := utils.NewLogger("UserHandler.CheckUsername")
	
	username := c.Query("username")
	if username == "" {
		err := fiber.NewError(fiber.StatusBadRequest, "username is required")
		logger.LogInput(username)
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Check if username is valid (only alphanumeric characters)
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(username) {
		err := fiber.NewError(fiber.StatusBadRequest, "username can only contain letters and numbers")
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
			"available": false,
		})
	}

	// Check length
	if len(username) < 3 || len(username) > 15 {
		err := fiber.NewError(fiber.StatusBadRequest, "username must be between 3 and 15 characters")
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
			"available": false,
		})
	}

	logger.LogInput(username)
	user, err := h.userUseCase.GetUserByUsername(username)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput(map[string]bool{"available": user == nil}, nil)
	return c.JSON(fiber.Map{
		"available": user == nil,
	})
}

func (h *UserHandler) DeleteAccount(c *fiber.Ctx) error {
	logger := utils.NewLogger("UserHandler.DeleteAccount")
	
	userID := c.Locals("user_id").(string)
	authClient := c.Locals("firebase_auth")
	logger.LogInput(userID)

	err := h.userUseCase.DeleteAccount(userID, authClient)
	if err != nil {
		logger.LogOutput(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.LogOutput("success", nil)
	return c.JSON(fiber.Map{
		"message": "Account deleted successfully",
	})
}
