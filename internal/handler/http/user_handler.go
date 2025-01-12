package handler

import (
	"regexp"
	"time"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userUseCase domain.UserUseCase
}

func NewUserHandler(router fiber.Router, userUseCase domain.UserUseCase) *UserHandler {
	handler := &UserHandler{
		userUseCase: userUseCase,
	}

	router.Patch("/", handler.UpdateUser)
	router.Delete("/", handler.DeleteAccount)
	router.Post("/", handler.CreateOrUpdateUser)
	router.Find("/me", handler.FindProfile)
	router.Find("/check-username", handler.CheckUsername)
	router.Find("/list", handler.FindUserFindMany)
	router.Find("/:username", handler.FindUserByUsername)

	return handler
}

func (h *UserHandler) CreateOrUpdateUser(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("UserHandler.CreateOrUpdateUser")

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	email := c.Locals("email").(string)

	var req struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		PhotoURL  string `json:"photoUrl"`
	}

	if err := c.BodyParser(&req); err != nil {
		logger.Input(req)
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Email validation
	if email == "" {
		err := fiber.NewError(fiber.StatusBadRequest, "email is required")
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(email) {
		err := fiber.NewError(fiber.StatusBadRequest, "invalid email format")
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Input(userID, email, req)
	user, err := h.userUseCase.CreateOrUpdateUser(
		userID.Hex(),
		email,
		req.FirstName,
		req.LastName,
		req.PhotoURL,
	)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(user, nil)
	return c.JSON(fiber.Map{
		"user": user,
	})
}

func (h *UserHandler) FindProfile(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("UserHandler.FindProfile")

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	logger.Input(userID)

	user, err := h.userUseCase.FindUserByID(userID.Hex())
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(user, nil)
	return c.JSON(fiber.Map{
		"user": user,
	})
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("UserHandler.UpdateUser")

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req struct {
		FirstName      *string              `json:"firstName"`
		LastName       *string              `json:"lastName"`
		Username       *string              `json:"username"`
		DisplayName    *string              `json:"displayName"`
		Bio            *string              `json:"bio"`
		Avatar         *string              `json:"avatar"`
		PhotoProfile   *string              `json:"photoProfile"`
		PhotoCover     *string              `json:"photoCover"`
		DateOfBirth    *time.Time           `json:"dateOfBirth"`
		Gender         *string              `json:"gender"`
		InterestedIn   []string             `json:"interestedIn"`
		Location       *domain.GeoLocation  `json:"location"`
		RelationStatus *string              `json:"relationStatus"`
		Height         *float64             `json:"height"`
		Interests      []string             `json:"interests"`
		Occupation     *string              `json:"occupation"`
		Education      *string              `json:"education"`
		PhoneNumber    *string              `json:"phoneNumber"`
		DatingPhotos   []domain.DatingPhoto `json:"datingPhotos"`
		IsVerified     *bool                `json:"isVerified"`
		IsActive       *bool                `json:"isActive"`
		Live           *domain.Live         `json:"live"`
	}

	if err := c.BodyParser(&req); err != nil {
		logger.Input(req)
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	logger.Input(userID, req)
	user, err := h.userUseCase.FindUserByID(userID.Hex())
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Username validation if provided
	if req.Username != nil {
		if *req.Username == "" {
			err := fiber.NewError(fiber.StatusBadRequest, "username is required")
			logger.Output(nil, err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(*req.Username) {
			err := fiber.NewError(fiber.StatusBadRequest, "username can only contain letters and numbers")
			logger.Output(nil, err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if len(*req.Username) < 3 || len(*req.Username) > 20 {
			err := fiber.NewError(fiber.StatusBadRequest, "username must be between 3 and 20 characters")
			logger.Output(nil, err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		// Check if username is already taken by another user
		existingUser, err := h.userUseCase.FindUserByUsername(*req.Username)
		if err != nil {
			logger.Output(nil, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if existingUser != nil && existingUser.ID != user.ID {
			err := fiber.NewError(fiber.StatusBadRequest, "username is already taken")
			logger.Output(nil, err)
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
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(user, nil)
	return c.JSON(fiber.Map{
		"user": user,
	})
}

func (h *UserHandler) FindUserByUsername(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("UserHandler.FindUserByUsername")

	username := c.Params("username")
	if username == "" {
		err := fiber.NewError(fiber.StatusBadRequest, "username is required")
		logger.Input(username)
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Input(username)
	user, err := h.userUseCase.FindUserByUsername(username)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if user == nil {
		err := fiber.NewError(fiber.StatusNotFound, "user not found")
		logger.Output(nil, err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(user, nil)
	return c.JSON(fiber.Map{
		"user": user,
	})
}

func (h *UserHandler) FindUserFindMany(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("UserHandler.FindUserFindMany")

	// Parse query parameters
	req := &domain.UserFindManyRequest{
		Page:     c.QueryInt("page", 1),
		PageSize: c.QueryInt("pageSize", 10),
		Search:   c.Query("search"),
		SortBy:   c.Query("sortBy"),
		SortDir:  c.Query("sortDir"),
		Status:   c.Query("status"),
	}

	logger.Input(req)

	// Find user list from use case
	response, err := h.userUseCase.FindUserFindMany(req)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(response, nil)
	return c.JSON(response)
}

func (h *UserHandler) CheckUsername(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("UserHandler.CheckUsername")

	username := c.Query("username")
	if username == "" {
		err := fiber.NewError(fiber.StatusBadRequest, "username is required")
		logger.Input(username)
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Check if username is valid (only alphanumeric characters)
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(username) {
		err := fiber.NewError(fiber.StatusBadRequest, "username can only contain letters and numbers")
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":     err.Error(),
			"available": false,
		})
	}

	// Check length
	if len(username) < 3 || len(username) > 15 {
		err := fiber.NewError(fiber.StatusBadRequest, "username must be between 3 and 15 characters")
		logger.Output(nil, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":     err.Error(),
			"available": false,
		})
	}

	logger.Input(username)
	user, err := h.userUseCase.FindUserByUsername(username)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(map[string]bool{"available": user == nil}, nil)
	return c.JSON(fiber.Map{
		"available": user == nil,
	})
}

func (h *UserHandler) DeleteAccount(c *fiber.Ctx) error {
	logger := utils.NewTraceLogger("UserHandler.DeleteAccount")

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	authClient := c.Locals("firebase_auth")
	logger.Input(userID)

	err = h.userUseCase.DeleteAccount(userID.Hex(), authClient)
	if err != nil {
		logger.Output(nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output("success", nil)
	return c.JSON(fiber.Map{
		"message": "Account deleted successfully",
	})
}
