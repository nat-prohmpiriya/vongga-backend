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
	FirebaseUID string             `bson:"firebase_uid" json:"firebase_uid"`
	FirstName   string             `bson:"first_name" json:"first_name"`
	LastName    string             `bson:"last_name" json:"last_name"`
	Email       string             `bson:"email" json:"email"`
	PhotoURL    string             `bson:"photo_url" json:"photo_url"`
	Provider    AuthProvider       `bson:"provider" json:"provider"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
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
