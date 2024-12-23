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
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updatedAt"`
	DeletedAt *time.Time         `bson:"deleted_at,omitempty" json:"deletedAt,omitempty"`
	IsActive  bool               `bson:"is_active" json:"isActive"`
	Version   int                `bson:"version" json:"version"`
}

type GeoLocation struct {
	Type        string    `bson:"type" json:"type"`
	Coordinates []float64 `bson:"coordinates" json:"coordinates"`
}

type DatingPhoto struct {
	URL        string `bson:"url" json:"url"`
	IsMain     bool   `bson:"is_main" json:"isMain"`
	IsApproved bool   `bson:"is_approved" json:"isApproved"`
}

type User struct {
	BaseModel      `bson:",inline"`
	FirebaseUID    string        `bson:"firebase_uid" json:"-"`
	Username       string        `bson:"username" json:"username"`
	Email          string        `bson:"email" json:"email"`
	Password       string        `bson:"password,omitempty" json:"-"`
	FirstName      string        `bson:"first_name" json:"firstName"`
	LastName       string        `bson:"last_name" json:"lastName"`
	Avatar         string        `bson:"avatar" json:"avatar"`
	Bio            string        `bson:"bio" json:"bio"`
	PhotoProfile   string        `bson:"photo_profile" json:"photoProfile"`
	PhotoCover     string        `bson:"photo_cover" json:"photoCover"`
	FollowersCount int           `bson:"followers_count" json:"followersCount"`
	FollowingCount int           `bson:"following_count" json:"followingCount"`
	FriendsCount   int           `bson:"friends_count" json:"friendsCount"`
	Provider       AuthProvider  `bson:"provider" json:"provider"`
	EmailVerified  bool          `bson:"email_verified" json:"emailVerified"`
	DateOfBirth    time.Time     `bson:"date_of_birth" json:"dateOfBirth"`
	Gender         string        `bson:"gender" json:"gender"`
	InterestedIn   []string      `bson:"interested_in" json:"interestedIn"`
	Location       GeoLocation   `bson:"location" json:"location"`
	RelationStatus string        `bson:"relation_status" json:"relationStatus"`
	Height         float64       `bson:"height" json:"height"`
	Interests      []string      `bson:"interests" json:"interests"`
	Occupation     string        `bson:"occupation" json:"occupation"`
	Education      string        `bson:"education" json:"education"`
	DatingPhotos   []DatingPhoto `bson:"dating_photos" json:"datingPhotos"`
	IsVerified     bool          `bson:"is_verified" json:"isVerified"`
	IsActive       bool          `bson:"is_active" json:"isActive"`
	PhoneNumber    string        `bson:"phone_number,omitempty" json:"phoneNumber,omitempty"`
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
