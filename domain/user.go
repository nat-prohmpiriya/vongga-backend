package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthProvider string

const (
	Google AuthProvider = "google"
	Apple  AuthProvider = "apple"
	Email  AuthProvider = "email"
)

type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updatedAt"`
	FirebaseUID string             `bson:"firebase_uid" json:"firebaseUid"`
	DisplayName string             `bson:"display_name" json:"displayName"`
	FirstName   string             `bson:"first_name" json:"firstName"`
	LastName    string             `bson:"last_name" json:"lastName"`
	Email       string             `bson:"email" json:"email"`
	PhotoURL    string             `bson:"photo_url" json:"photoUrl"`
	Provider    AuthProvider       `bson:"provider" json:"provider"`
}

type UserRepository interface {
	Create(user *User) error
	FindByFirebaseUID(firebaseUID string) (*User, error)
	FindByEmail(email string) (*User, error)
	FindByID(id string) (*User, error)
	Update(user *User) error
}

type UserUseCase interface {
	CreateOrUpdateUser(firebaseUID, email, firstName, lastName, photoURL string) (*User, error)
	GetUserByID(id string) (*User, error)
	GetUserByFirebaseUID(firebaseUID string) (*User, error)
}
