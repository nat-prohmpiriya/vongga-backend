package domain

import (
	"time"
)

type AuthProvider string

const (
	Google AuthProvider = "google"
	Apple  AuthProvider = "apple"
	Email  AuthProvider = "email"
)

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
	DisplayName    string        `bson:"displayName" json:"displayName"`
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

type UserListItem struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	DisplayName    string `json:"displayName"`
	Email          string `json:"email"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	Avatar         string `json:"avatar"`
	PhotoProfile   string `json:"photoProfile"`
	PhotoCover     string `json:"photoCover"`
	FollowersCount int    `json:"followersCount"`
	FollowingCount int    `json:"followingCount"`
	FriendsCount   int    `json:"friendsCount"`
}

type UserListRequest struct {
	Page     int    `json:"page" query:"page"`
	PageSize int    `json:"pageSize" query:"pageSize"`
	Search   string `json:"search" query:"search"`
	SortBy   string `json:"sortBy" query:"sortBy"`
	SortDir  string `json:"sortDir" query:"sortDir"`
	Status   string `json:"status" query:"status"`
}

type UserListResponse struct {
	Users      []UserListItem `json:"users"`
	TotalCount int64         `json:"totalCount"`
	Page       int           `json:"page"`
	PageSize   int          `json:"pageSize"`
}

type UserRepository interface {
	Create(user *User) error
	FindByFirebaseUID(firebaseUID string) (*User, error)
	FindByEmail(email string) (*User, error)
	FindByID(id string) (*User, error)
	FindByUsername(username string) (*User, error)
	Update(user *User) error
	SoftDelete(id string) error
	GetUserList(req *UserListRequest) ([]User, int64, error)
}

type UserUseCase interface {
	CreateOrUpdateUser(firebaseUID, email, firstName, lastName, photoURL string) (*User, error)
	GetUserByID(id string) (*User, error)
	GetUserByFirebaseUID(firebaseUID string) (*User, error)
	GetUserByUsername(username string) (*User, error)
	UpdateUser(user *User) error
	DeleteAccount(userID string, authClient interface{}) error
	GetUserList(req *UserListRequest) (*UserListResponse, error)
}
