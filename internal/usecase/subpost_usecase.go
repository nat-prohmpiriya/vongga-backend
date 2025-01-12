package usecase

import (
	"time"

	"vongga-api/internal/domain"
	"vongga-api/utils"

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
	logger := utils.NewTraceLogger("SubPostUseCase.CreateSubPost")
	input := map[string]interface{}{
		"parentID": parentID,
		"userID":   userID,
		"content":  content,
		"media":    media,
		"order":    order,
	}
	logger.Input(input)

	// Find parent post to increment subpost count
	post, err := s.postRepo.FindByID(parentID)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	subPost := &domain.SubPost{
		ParentID:       parentID,
		UserID:         userID,
		Content:        content,
		Media:          media,
		ReactionCounts: make(map[string]int),
		CommentCount:   0,
		Order:          order,
	}

	err = s.subPostRepo.Create(subPost)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	// Increment subpost count in parent post
	post.SubPostCount++
	err = s.postRepo.Update(post)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(subPost, nil)
	return subPost, nil
}

func (s *subPostUseCase) UpdateSubPost(subPostID primitive.ObjectID, content string, media []domain.Media) (*domain.SubPost, error) {
	logger := utils.NewTraceLogger("SubPostUseCase.UpdateSubPost")
	input := map[string]interface{}{
		"subPostID": subPostID,
		"content":   content,
		"media":     media,
	}
	logger.Input(input)

	subPost, err := s.subPostRepo.FindByID(subPostID)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	subPost.Content = content
	subPost.Media = media
	subPost.UpdatedAt = time.Now()

	err = s.subPostRepo.Update(subPost)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(subPost, nil)
	return subPost, nil
}

func (s *subPostUseCase) DeleteSubPost(subPostID primitive.ObjectID) error {
	logger := utils.NewTraceLogger("SubPostUseCase.DeleteSubPost")
	logger.Input(subPostID)

	// Find subpost to get parentID
	subPost, err := s.subPostRepo.FindByID(subPostID)
	if err != nil {
		logger.Output(nil, err)
		return err
	}

	// Find parent post to decrement subpost count
	post, err := s.postRepo.FindByID(subPost.ParentID)
	if err != nil {
		logger.Output(nil, err)
		return err
	}

	err = s.subPostRepo.Delete(subPostID)
	if err != nil {
		logger.Output(nil, err)
		return err
	}

	// Decrement subpost count in parent post
	if post.SubPostCount > 0 {
		post.SubPostCount--
		err = s.postRepo.Update(post)
		if err != nil {
			logger.Output(nil, err)
			return err
		}
	}

	logger.Output("SubPost deleted successfully", nil)
	return nil
}

func (s *subPostUseCase) FindSubPost(subPostID primitive.ObjectID) (*domain.SubPost, error) {
	logger := utils.NewTraceLogger("SubPostUseCase.FindSubPost")
	logger.Input(subPostID)

	subPost, err := s.subPostRepo.FindByID(subPostID)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(subPost, nil)
	return subPost, nil
}

func (s *subPostUseCase) FindManySubPosts(parentID primitive.ObjectID, limit, offset int) ([]domain.SubPost, error) {
	logger := utils.NewTraceLogger("SubPostUseCase.FindManySubPosts")
	input := map[string]interface{}{
		"parentID": parentID,
		"limit":    limit,
		"offset":   offset,
	}
	logger.Input(input)

	subPosts, err := s.subPostRepo.FindByParentID(parentID, limit, offset)
	if err != nil {
		logger.Output(nil, err)
		return nil, err
	}

	logger.Output(subPosts, nil)
	return subPosts, nil
}

func (s *subPostUseCase) ReorderSubPosts(parentID primitive.ObjectID, orders map[primitive.ObjectID]int) error {
	logger := utils.NewTraceLogger("SubPostUseCase.ReorderSubPosts")
	input := map[string]interface{}{
		"parentID": parentID,
		"orders":   orders,
	}
	logger.Input(input)

	err := s.subPostRepo.UpdateOrder(parentID, orders)
	if err != nil {
		logger.Output(nil, err)
		return err
	}

	logger.Output("SubPosts reordered successfully", nil)
	return nil
}
