package usecase

import (
	"context"
	"fmt"
	"time"

	"vongga_api/internal/domain"
	"vongga_api/utils"

	"firebase.google.com/go/v4/auth"
	"go.opentelemetry.io/otel/trace"
)

type userUseCase struct {
	userRepo domain.UserRepository
	tracer   trace.Tracer
}

func NewUserUseCase(userRepo domain.UserRepository, tracer trace.Tracer) domain.UserUseCase {
	return &userUseCase{
		userRepo: userRepo,
		tracer:   tracer,
	}
}

func (u *userUseCase) CreateOrUpdateUser(ctx context.Context, firebaseUID, email, firstName, lastName, photoURL string) (*domain.User, error) {
	ctx, span := u.tracer.Start(ctx, "UserUseCase.CreateOrUpdateUser")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"firebaseUID": firebaseUID,
		"email":       email,
		"firstName":   firstName,
		"lastName":    lastName,
		"photoURL":    photoURL,
	}
	logger.Input(input)

	// Check if user exists
	user, err := u.userRepo.FindByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		logger.Output("error finding user 1", err)
		return nil, err
	}

	if user == nil {
		// Create new user
		user = &domain.User{
			FirebaseUID: firebaseUID,
			Email:       email,
			FirstName:   firstName,
			LastName:    lastName,
			Avatar:      photoURL,
		}

		err = u.userRepo.Create(ctx, user)
		if err != nil {
			logger.Output("error creating user 2", err)
			return nil, err
		}

		logger.Output(user, nil)
		return user, nil
	}

	// Update existing user if needed
	if user.Email != email || user.Avatar != photoURL || user.FirstName != firstName || user.LastName != lastName {
		user.Email = email
		user.FirstName = firstName
		user.LastName = lastName
		user.Avatar = photoURL
		user.UpdatedAt = time.Now()
		user.Version++

		err = u.userRepo.Update(ctx, user)
		if err != nil {
			logger.Output("error updating user 3", err)
			return nil, err
		}
	}

	logger.Output(user, nil)
	return user, nil
}

func (u *userUseCase) FindUserByFirebaseUID(ctx context.Context, firebaseUID string) (*domain.User, error) {
	ctx, span := u.tracer.Start(ctx, "UserUseCase.FindUserByFirebaseUID")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	logger.Input(firebaseUID)

	user, err := u.userRepo.FindByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		logger.Output("error finding user 1", err)
		return nil, err
	}

	if user == nil {
		err = fmt.Errorf("user not found")
		logger.Output("user not found 2", err)
		return nil, err
	}

	logger.Output(user, nil)
	return user, nil
}

func (u *userUseCase) FindUserByID(ctx context.Context, id string) (*domain.User, error) {
	ctx, span := u.tracer.Start(ctx, "UserUseCase.FindUserByID")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	logger.Input(id)

	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		logger.Output("error finding user 1", err)
		return nil, err
	}

	if user == nil {
		err = fmt.Errorf("user not found")
		logger.Output("user not found 2", err)
		return nil, err
	}

	logger.Output(user, nil)
	return user, nil
}

func (u *userUseCase) FindUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	ctx, span := u.tracer.Start(ctx, "UserUseCase.FindUserByUsername")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	logger.Input(username)

	user, err := u.userRepo.FindByUsername(ctx, username)
	if err != nil {
		logger.Output("error finding user 1", err)
		return nil, err
	}

	logger.Output(user, nil)
	return user, nil
}

func (u *userUseCase) UpdateUser(ctx context.Context, user *domain.User) error {
	ctx, span := u.tracer.Start(ctx, "UserUseCase.UpdateUser")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	logger.Input(user)

	err := u.userRepo.Update(ctx, user)
	if err != nil {
		logger.Output("error updating user 1", err)
		return err
	}

	logger.Output(user, nil)
	return nil
}

func (u *userUseCase) DeleteAccount(ctx context.Context, userID string, authClient interface{}) error {
	ctx, span := u.tracer.Start(ctx, "UserUseCase.DeleteAccount")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	logger.Input(userID)

	// Find user to get Firebase UID
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		logger.Output("error finding user 1", err)
		return err
	}
	if user == nil {
		err = fmt.Errorf("user not found")
		logger.Output("user not found 2", err)
		return err
	}

	// Delete user from Firebase
	firebaseAuth := authClient.(*auth.Client)
	err = firebaseAuth.DeleteUser(context.Background(), user.FirebaseUID)
	if err != nil {
		logger.Output("error deleting user from Firebase 3", err)
		return err
	}

	// Soft delete user in our database
	err = u.userRepo.SoftDelete(ctx, userID)
	if err != nil {
		logger.Output("error soft deleting user 4", err)
		return err
	}

	logger.Output("success", nil)
	return nil
}

func (u *userUseCase) FindUserFindMany(ctx context.Context, req *domain.UserFindManyRequest) (*domain.UserFindManyResponse, error) {
	ctx, span := u.tracer.Start(ctx, "UserUseCase.FindUserFindMany")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	logger.Input(req)

	// Validate request
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// Validate sort parameters
	validSortFields := map[string]bool{
		"createdAt": true,
		"firstName": true,
		"lastName":  true,
		"username":  true,
	}
	if req.SortBy != "" && !validSortFields[req.SortBy] {
		req.SortBy = "createdAt"
	}
	if req.SortDir != "asc" && req.SortDir != "desc" {
		req.SortDir = "desc"
	}

	// Find users from repository
	users, totalCount, err := u.userRepo.FindUserFindMany(ctx, req)
	if err != nil {
		logger.Output("error finding users 1", err)
		return nil, err
	}

	// Convert User to UserFindManyItem
	userItems := make([]domain.UserFindManyItem, len(users))
	for i, user := range users {
		userItems[i] = domain.UserFindManyItem{
			ID:             user.ID.Hex(),
			Username:       user.Username,
			DisplayName:    user.DisplayName,
			Email:          user.Email,
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			Avatar:         user.Avatar,
			PhotoProfile:   user.PhotoProfile,
			PhotoCover:     user.PhotoCover,
			FollowersCount: user.FollowersCount,
			FollowingCount: user.FollowingCount,
			FriendsCount:   user.FriendsCount,
		}
	}

	response := &domain.UserFindManyResponse{
		Users:      userItems,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}

	logger.Output(response, nil)
	return response, nil
}
