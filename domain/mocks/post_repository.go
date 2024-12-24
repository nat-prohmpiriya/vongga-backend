package mocks

import (
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostRepository struct {
	mock.Mock
}

func (m *PostRepository) Create(post *domain.Post) error {
	args := m.Called(post)
	return args.Error(0)
}

func (m *PostRepository) Update(post *domain.Post) error {
	args := m.Called(post)
	return args.Error(0)
}

func (m *PostRepository) Delete(id primitive.ObjectID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *PostRepository) FindByID(id primitive.ObjectID) (*domain.Post, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Post), args.Error(1)
}

func (m *PostRepository) FindByUserID(userID primitive.ObjectID, limit, offset int) ([]domain.Post, error) {
	args := m.Called(userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Post), args.Error(1)
}

func (m *PostRepository) FindSubPosts(postID primitive.ObjectID) ([]domain.SubPost, error) {
	args := m.Called(postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.SubPost), args.Error(1)
}
