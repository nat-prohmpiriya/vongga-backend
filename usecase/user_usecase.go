package usecase

import (
	// "context"
	"time"

	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"go.mongodb.org/mongo-driver/mongo"
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
		if err == mongo.ErrNoDocuments {
			// Create new user
			user = &domain.User{
				FirebaseUID: firebaseUID,
				Email:       email,
				FirstName:   firstName,
				LastName:    lastName,
				PhotoURL:    photoURL,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
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
	user.Email = email
	user.FirstName = firstName
	user.LastName = lastName
	user.PhotoURL = photoURL
	user.UpdatedAt = time.Now()

	err = u.userRepo.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userUseCase) GetUserByID(id string) (*domain.User, error) {
	return u.userRepo.FindByID(id)
}

func (u *userUseCase) GetUserByFirebaseUID(firebaseUID string) (*domain.User, error) {
	return u.userRepo.FindByFirebaseUID(firebaseUID)
}
