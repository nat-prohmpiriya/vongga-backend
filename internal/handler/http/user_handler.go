package handler

import (
	"regexp"
	"time"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

type UserHandler struct {
	userUseCase domain.UserUseCase
	tracer      trace.Tracer
}

func NewUserHandler(router fiber.Router, userUseCase domain.UserUseCase, tracer trace.Tracer) *UserHandler {
	handler := &UserHandler{
		tracer:      tracer,
		userUseCase: userUseCase,
	}

	router.Patch("/", handler.UpdateUser)
	router.Delete("/", handler.DeleteAccount)
	router.Post("/", handler.CreateOrUpdateUser)
	router.Get("/me", handler.FindProfile)
	router.Get("/check-username", handler.CheckUsername)
	router.Get("/list", handler.FindUserFindMany)
	router.Get("/:username", handler.FindUserByUsername)

	return handler
}

func (h *UserHandler) CreateOrUpdateUser(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "UserHandler.CreateOrUpdateUser")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("unauthorized access attempt 1", err)
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
		logger.Output("invalid request body 2", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Email validation
	if email == "" {
		err := fiber.NewError(fiber.StatusBadRequest, "email is required")
		logger.Output("missing email 3", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(email) {
		err := fiber.NewError(fiber.StatusBadRequest, "invalid email format")
		logger.Output("invalid email format 4", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Input(map[string]interface{}{
		"userID":    userID,
		"email":     email,
		"firstName": req.FirstName,
		"lastName":  req.LastName,
		"photoURL":  req.PhotoURL,
	})
	user, err := h.userUseCase.CreateOrUpdateUser(
		ctx,
		userID.Hex(),
		email,
		req.FirstName,
		req.LastName,
		req.PhotoURL,
	)
	if err != nil {
		logger.Output("failed to create/update user 5", err)
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
	ctx, span := h.tracer.Start(c.Context(), "UserHandler.FindProfile")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("unauthorized access attempt 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	logger.Input(userID)

	user, err := h.userUseCase.FindUserByID(ctx, userID.Hex())
	if err != nil {
		logger.Output("failed to find user profile 2", err)
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
	ctx, span := h.tracer.Start(c.Context(), "UserHandler.UpdateUser")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("unauthorized access attempt 1", err)
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
		logger.Output("invalid request body 2", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	logger.Input(map[string]interface{}{
		"userID":         userID,
		"firstName":      req.FirstName,
		"lastName":       req.LastName,
		"username":       req.Username,
		"displayName":    req.DisplayName,
		"bio":            req.Bio,
		"avatar":         req.Avatar,
		"photoProfile":   req.PhotoProfile,
		"photoCover":     req.PhotoCover,
		"dateOfBirth":    req.DateOfBirth,
		"gender":         req.Gender,
		"interestedIn":   req.InterestedIn,
		"location":       req.Location,
		"relationStatus": req.RelationStatus,
		"height":         req.Height,
		"interests":      req.Interests,
		"occupation":     req.Occupation,
		"education":      req.Education,
		"phoneNumber":    req.PhoneNumber,
		"datingPhotos":   req.DatingPhotos,
		"isVerified":     req.IsVerified,
		"isActive":       req.IsActive,
		"live":           req.Live,
	})
	user, err := h.userUseCase.FindUserByID(ctx, userID.Hex())
	if err != nil {
		logger.Output("failed to find user for update 3", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Username validation if provided
	if req.Username != nil {
		if *req.Username == "" {
			err := fiber.NewError(fiber.StatusBadRequest, "username is required")
			logger.Output("missing username 4", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(*req.Username) {
			err := fiber.NewError(fiber.StatusBadRequest, "username can only contain letters and numbers")
			logger.Output("invalid username format 5", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if len(*req.Username) < 3 || len(*req.Username) > 20 {
			err := fiber.NewError(fiber.StatusBadRequest, "username must be between 3 and 20 characters")
			logger.Output("invalid username length 6", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		// Check if username is already taken by another user
		existingUser, err := h.userUseCase.FindUserByUsername(ctx, *req.Username)
		if err != nil {
			logger.Output("failed to check username availability 7", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if existingUser != nil && existingUser.ID != user.ID {
			err := fiber.NewError(fiber.StatusBadRequest, "username is already taken")
			logger.Output("username already taken 8", err)
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

	err = h.userUseCase.UpdateUser(ctx, user)
	if err != nil {
		logger.Output("failed to update user 9", err)
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
	ctx, span := h.tracer.Start(c.Context(), "UserHandler.FindUserByUsername")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	username := c.Params("username")
	if username == "" {
		err := fiber.NewError(fiber.StatusBadRequest, "username is required")
		logger.Input(username)
		logger.Output("missing username 1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Input(username)
	user, err := h.userUseCase.FindUserByUsername(ctx, username)
	if err != nil {
		logger.Output("failed to find user by username 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if user == nil {
		err := fiber.NewError(fiber.StatusNotFound, "user not found")
		logger.Output("user not found 3", err)
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
	ctx, span := h.tracer.Start(c.Context(), "UserHandler.FindUserFindMany")
	defer span.End()
	logger := utils.NewTraceLogger(span)

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
	response, err := h.userUseCase.FindUserFindMany(ctx, req)
	if err != nil {
		logger.Output("failed to find users 1", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output(response, nil)
	return c.JSON(response)
}

func (h *UserHandler) CheckUsername(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "UserHandler.CheckUsername")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	username := c.Query("username")
	if username == "" {
		err := fiber.NewError(fiber.StatusBadRequest, "username is required")
		logger.Input(username)
		logger.Output("missing username 1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Check if username is valid (only alphanumeric characters)
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(username) {
		err := fiber.NewError(fiber.StatusBadRequest, "username can only contain letters and numbers")
		logger.Output("invalid username format 2", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":     err.Error(),
			"available": false,
		})
	}

	// Check length
	if len(username) < 3 || len(username) > 15 {
		err := fiber.NewError(fiber.StatusBadRequest, "username must be between 3 and 15 characters")
		logger.Output("invalid username length 3", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":     err.Error(),
			"available": false,
		})
	}

	logger.Input(username)
	user, err := h.userUseCase.FindUserByUsername(ctx, username)
	if err != nil {
		logger.Output("failed to check username 4", err)
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
	ctx, span := h.tracer.Start(c.Context(), "UserHandler.DeleteAccount")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	userID, err := utils.FindUserIDFromContext(c)
	if err != nil {
		logger.Output("unauthorized access attempt 1", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	authClient := c.Locals("firebase_auth")
	logger.Input(userID)

	err = h.userUseCase.DeleteAccount(ctx, userID.Hex(), authClient)
	if err != nil {
		logger.Output("failed to delete account 2", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	logger.Output("success", nil)
	return c.JSON(fiber.Map{
		"message": "Account deleted successfully",
	})
}
