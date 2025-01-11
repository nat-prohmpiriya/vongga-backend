package mocks

import (
	"vongga-api/internal/domain"

	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubPostRepository struct {
	mock.Mock
}

func (m *SubPostRepository) Create(subPost *domain.SubPost) error {
	args := m.Called(subPost)
	return args.Error(0)
}

func (m *SubPostRepository) Update(subPost *domain.SubPost) error {
	args := m.Called(subPost)
	return args.Error(0)
}

func (m *SubPostRepository) Delete(id primitive.ObjectID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *SubPostRepository) FindByID(id primitive.ObjectID) (*domain.SubPost, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SubPost), args.Error(1)
}

func (m *SubPostRepository) FindByPostID(postID primitive.ObjectID) ([]domain.SubPost, error) {
	args := m.Called(postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.SubPost), args.Error(1)
}

func (m *SubPostRepository) FindByParentID(parentID primitive.ObjectID, limit, offset int) ([]domain.SubPost, error) {
	args := m.Called(parentID, limit, offset)
	if args.Get(0) == nil {
		return []domain.SubPost{}, args.Error(1)
	}
	return args.Get(0).([]domain.SubPost), args.Error(1)
}

func (m *SubPostRepository) UpdateOrder(postID primitive.ObjectID, orderMap map[primitive.ObjectID]int) error {
	args := m.Called(postID, orderMap)
	return args.Error(0)
}
