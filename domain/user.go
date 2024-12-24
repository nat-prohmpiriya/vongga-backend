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

type BaseModel struct {
	ID        primitive.ObjectID `bson:"id,omitempty" json:"id"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
	DeletedAt *time.Time         `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`
	IsActive  bool               `bson:"isActive" json:"isActive"`
	Version   int                `bson:"version" json:"version"`
}

type GeoLocation struct {
	Type        string    `bson:"type" json:"type"`
	Coordinates []float64 `bson:"coordinates" json:"coordinates"`
}

type DatingPhoto struct {
	URL        string `bson:"url" json:"url"`
	IsMain     bool   `bson:"isMain" json:"isMain"`
	IsApproved bool   `bson:"isApproved" json:"isApproved"`
}

type User struct {
	BaseModel      `bson:",inline"`
	FirebaseUID    string        `bson:"firebaseUid" json:"-"`
	Username       string        `bson:"username" json:"username"`
	Email          string        `bson:"email" json:"email"`
	Password       string        `bson:"password,omitempty" json:"-"`
	FirstName      string        `bson:"firstName" json:"firstName"`
	LastName       string        `bson:"lastName" json:"lastName"`
	Avatar         string        `bson:"avatar" json:"avatar"`
	Bio            string        `bson:"bio" json:"bio"`
	PhotoProfile   string        `bson:"photoProfile" json:"photoProfile"`
	PhotoCover     string        `bson:"photoCover" json:"photoCover"`
	FollowersCount int           `bson:"followersCount" json:"followersCount"`
	FollowingCount int           `bson:"followingCount" json:"followingCount"`
	FriendsCount   int           `bson:"friendsCount" json:"friendsCount"`
	Provider       AuthProvider  `bson:"provider" json:"provider"`
	EmailVerified  bool          `bson:"emailVerified" json:"emailVerified"`
	DateOfBirth    time.Time     `bson:"dateOfBirth" json:"dateOfBirth"`
	Gender         string        `bson:"gender" json:"gender"`
	InterestedIn   []string      `bson:"interestedIn" json:"interestedIn"`
	Location       GeoLocation   `bson:"location" json:"location"`
	RelationStatus string        `bson:"relationStatus" json:"relationStatus"`
	Height         float64       `bson:"height" json:"height"`
	Interests      []string      `bson:"interests" json:"interests"`
	Occupation     string        `bson:"occupation" json:"occupation"`
	Education      string        `bson:"education" json:"education"`
	DatingPhotos   []DatingPhoto `bson:"datingPhotos" json:"datingPhotos"`
	IsVerified     bool          `bson:"isVerified" json:"isVerified"`
	IsActive       bool          `bson:"isActive" json:"isActive"`
	PhoneNumber    string        `bson:"phoneNumber,omitempty" json:"phoneNumber,omitempty"`
	Live           Live          `bson:"live" json:"live"`
}

type Live struct {
	City    string `bson:"city" json:"city"`
	Country string `bson:"country" json:"country"`
}

type UserRepository interface {
	Create(user *User) error
	FindByFirebaseUID(firebaseUID string) (*User, error)
	FindByEmail(email string) (*User, error)
	FindByID(id string) (*User, error)
	FindByUsername(username string) (*User, error)
	Update(user *User) error
}

type UserUseCase interface {
	CreateOrUpdateUser(firebaseUID, email, firstName, lastName, photoURL string) (*User, error)
	GetUserByID(id string) (*User, error)
	GetUserByFirebaseUID(firebaseUID string) (*User, error)
	GetUserByUsername(username string) (*User, error)
	UpdateUser(user *User) error
}
