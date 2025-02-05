package usecase

import (
	"context"
	"fmt"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
)

type userUseCase struct {
	userRepo domain.UserRepository
}

func NewUserUseCase(userRepo domain.UserRepository) domain.UserUseCase {
	return &userUseCase{
		userRepo: userRepo,
	}
}

func (u *userUseCase) CreateOrUpdateUser(firebaseUID, email, firstName, lastName, photoURL string) (*domain.User, error) {
	logger := utils.NewLogger("UserUseCase.CreateOrUpdateUser")
	input := map[string]interface{}{
		"firebaseUID": firebaseUID,
		"email":       email,
		"firstName":   firstName,
		"lastName":    lastName,
		"photoURL":    photoURL,
	}
	logger.LogInput(input)

	// Check if user exists
	user, err := u.userRepo.FindByFirebaseUID(firebaseUID)
	if err != nil {
		logger.LogOutput(nil, err)
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

		err = u.userRepo.Create(user)
		if err != nil {
			logger.LogOutput(nil, err)
			return nil, err
		}

		logger.LogOutput(user, nil)
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

		err = u.userRepo.Update(user)
		if err != nil {
			logger.LogOutput(nil, err)
			return nil, err
		}
	}

	logger.LogOutput(user, nil)
	return user, nil
}

func (u *userUseCase) GetUserByFirebaseUID(firebaseUID string) (*domain.User, error) {
	logger := utils.NewLogger("UserUseCase.GetUserByFirebaseUID")
	logger.LogInput(firebaseUID)

	user, err := u.userRepo.FindByFirebaseUID(firebaseUID)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	if user == nil {
		err = fmt.Errorf("user not found")
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(user, nil)
	return user, nil
}

func (u *userUseCase) GetUserByID(id string) (*domain.User, error) {
	logger := utils.NewLogger("UserUseCase.GetUserByID")
	logger.LogInput(id)

	user, err := u.userRepo.FindByID(id)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	if user == nil {
		err = fmt.Errorf("user not found")
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(user, nil)
	return user, nil
}

func (u *userUseCase) GetUserByUsername(username string) (*domain.User, error) {
	logger := utils.NewLogger("UserUseCase.GetUserByUsername")
	logger.LogInput(username)

	user, err := u.userRepo.FindByUsername(username)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(user, nil)
	return user, nil
}

func (u *userUseCase) UpdateUser(user *domain.User) error {
	logger := utils.NewLogger("UserUseCase.UpdateUser")
	logger.LogInput(user)

	err := u.userRepo.Update(user)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(user, nil)
	return nil
}

func (u *userUseCase) DeleteAccount(userID string, authClient interface{}) error {
	logger := utils.NewLogger("UserUseCase.DeleteAccount")
	logger.LogInput(userID)

	// Get user to get Firebase UID
	user, err := u.userRepo.FindByID(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}
	if user == nil {
		err = fmt.Errorf("user not found")
		logger.LogOutput(nil, err)
		return err
	}

	// Delete user from Firebase
	firebaseAuth := authClient.(*auth.Client)
	err = firebaseAuth.DeleteUser(context.Background(), user.FirebaseUID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Soft delete user in our database
	err = u.userRepo.SoftDelete(userID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput("success", nil)
	return nil
}

func (u *userUseCase) GetUserList(req *domain.UserListRequest) (*domain.UserListResponse, error) {
	logger := utils.NewLogger("UserUseCase.GetUserList")
	logger.LogInput(req)

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

	// Get users from repository
	users, totalCount, err := u.userRepo.GetUserList(req)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Convert User to UserListItem
	userItems := make([]domain.UserListItem, len(users))
	for i, user := range users {
		userItems[i] = domain.UserListItem{
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

	response := &domain.UserListResponse{
		Users:      userItems,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}

	logger.LogOutput(response, nil)
	return response, nil
}
