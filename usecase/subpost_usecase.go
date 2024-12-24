package usecase

import (
	"time"

	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type subPostUseCase struct {
	subPostRepo domain.SubPostRepository
	postRepo    domain.PostRepository
}

func NewSubPostUseCase(subPostRepo domain.SubPostRepository, postRepo domain.PostRepository) domain.SubPostUseCase {
	return &subPostUseCase{
		subPostRepo: subPostRepo,
		postRepo:    postRepo,
	}
}

func (s *subPostUseCase) CreateSubPost(parentID, userID primitive.ObjectID, content string, media []domain.Media, order int) (*domain.SubPost, error) {
	logger := utils.NewLogger("SubPostUseCase.CreateSubPost")
	input := map[string]interface{}{
		"parentID": parentID,
		"userID":   userID,
		"content":  content,
		"media":    media,
		"order":    order,
	}
	logger.LogInput(input)

	// Get parent post to increment subpost count
	post, err := s.postRepo.FindByID(parentID)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	subPost := &domain.SubPost{
		ParentID:       parentID,
		UserID:         userID,
		Content:        content,
		Media:          media,
		ReactionCounts: make(map[string]int),
		CommentCount:   0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Order:          order,
	}

	err = s.subPostRepo.Create(subPost)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Increment subpost count in parent post
	post.SubPostCount++
	err = s.postRepo.Update(post)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(subPost, nil)
	return subPost, nil
}

func (s *subPostUseCase) UpdateSubPost(subPostID primitive.ObjectID, content string, media []domain.Media) (*domain.SubPost, error) {
	logger := utils.NewLogger("SubPostUseCase.UpdateSubPost")
	input := map[string]interface{}{
		"subPostID": subPostID,
		"content":   content,
		"media":     media,
	}
	logger.LogInput(input)

	subPost, err := s.subPostRepo.FindByID(subPostID)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	subPost.Content = content
	subPost.Media = media
	subPost.UpdatedAt = time.Now()

	err = s.subPostRepo.Update(subPost)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(subPost, nil)
	return subPost, nil
}

func (s *subPostUseCase) DeleteSubPost(subPostID primitive.ObjectID) error {
	logger := utils.NewLogger("SubPostUseCase.DeleteSubPost")
	logger.LogInput(subPostID)

	// Get subpost to get parentID
	subPost, err := s.subPostRepo.FindByID(subPostID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Get parent post to decrement subpost count
	post, err := s.postRepo.FindByID(subPost.ParentID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	err = s.subPostRepo.Delete(subPostID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Decrement subpost count in parent post
	if post.SubPostCount > 0 {
		post.SubPostCount--
		err = s.postRepo.Update(post)
		if err != nil {
			logger.LogOutput(nil, err)
			return err
		}
	}

	logger.LogOutput("SubPost deleted successfully", nil)
	return nil
}

func (s *subPostUseCase) GetSubPost(subPostID primitive.ObjectID) (*domain.SubPost, error) {
	logger := utils.NewLogger("SubPostUseCase.GetSubPost")
	logger.LogInput(subPostID)

	subPost, err := s.subPostRepo.FindByID(subPostID)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(subPost, nil)
	return subPost, nil
}

func (s *subPostUseCase) ListSubPosts(parentID primitive.ObjectID, limit, offset int) ([]domain.SubPost, error) {
	logger := utils.NewLogger("SubPostUseCase.ListSubPosts")
	input := map[string]interface{}{
		"parentID": parentID,
		"limit":    limit,
		"offset":   offset,
	}
	logger.LogInput(input)

	subPosts, err := s.subPostRepo.FindByParentID(parentID, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(subPosts, nil)
	return subPosts, nil
}

func (s *subPostUseCase) ReorderSubPosts(parentID primitive.ObjectID, orders map[primitive.ObjectID]int) error {
	logger := utils.NewLogger("SubPostUseCase.ReorderSubPosts")
	input := map[string]interface{}{
		"parentID": parentID,
		"orders":   orders,
	}
	logger.LogInput(input)

	err := s.subPostRepo.UpdateOrder(parentID, orders)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput("SubPosts reordered successfully", nil)
	return nil
}
