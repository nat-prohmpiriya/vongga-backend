package usecase

import (
	"database/sql"
	"errors"

	"github.com/vongga/vongga-backend/domain"
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
	// Try to find existing user
	user, err := u.userRepo.FindByFirebaseUID(firebaseUID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create new user
			user = &domain.User{
				FirebaseUID: firebaseUID,
				Email:      email,
				FirstName:  firstName,
				LastName:   lastName,
				PhotoURL:   photoURL,
			}
			err = u.userRepo.Create(user)
			if err != nil {
				return nil, err
			}
			return user, nil
		}
		return nil, err
	}

	// Update existing user
	user.FirstName = firstName
	user.LastName = lastName
	user.PhotoURL = photoURL

	err = u.userRepo.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userUseCase) GetUserByID(id string) (*domain.User, error) {
	user, err := u.userRepo.FindByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

func (u *userUseCase) GetUserByFirebaseUID(firebaseUID string) (*domain.User, error) {
	user, err := u.userRepo.FindByFirebaseUID(firebaseUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}
