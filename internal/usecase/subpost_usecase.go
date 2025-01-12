package usecase

import (
	"context"
	"time"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
)

type subPostUseCase struct {
	subPostRepo domain.SubPostRepository
	postRepo    domain.PostRepository
	tracer      trace.Tracer
}

func NewSubPostUseCase(subPostRepo domain.SubPostRepository, postRepo domain.PostRepository, tracer trace.Tracer) domain.SubPostUseCase {
	return &subPostUseCase{
		subPostRepo: subPostRepo,
		postRepo:    postRepo,
		tracer:      tracer,
	}
}

func (s *subPostUseCase) CreateSubPost(ctx context.Context, parentID, userID primitive.ObjectID, content string, media []domain.Media, order int) (*domain.SubPost, error) {
	ctx, span := s.tracer.Start(ctx, "SubPostUseCase.CreateSubPost")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"parentID": parentID,
		"userID":   userID,
		"content":  content,
		"media":    media,
		"order":    order,
	}
	logger.Input(input)

	// Find parent post to increment subpost count
	post, err := s.postRepo.FindByID(ctx, parentID)
	if err != nil {
		logger.Output("error finding parent post 1", err)
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

	err = s.subPostRepo.Create(ctx, subPost)
	if err != nil {
		logger.Output("error creating subpost 2", err)
		return nil, err
	}

	// Increment subpost count in parent post
	post.SubPostCount++
	err = s.postRepo.Update(ctx, post)
	if err != nil {
		logger.Output("error updating parent post 3", err)
		return nil, err
	}

	logger.Output(subPost, nil)
	return subPost, nil
}

func (s *subPostUseCase) UpdateSubPost(ctx context.Context, subPostID primitive.ObjectID, content string, media []domain.Media) (*domain.SubPost, error) {
	ctx, span := s.tracer.Start(ctx, "SubPostUseCase.UpdateSubPost")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"subPostID": subPostID,
		"content":   content,
		"media":     media,
	}
	logger.Input(input)

	subPost, err := s.subPostRepo.FindByID(ctx, subPostID)
	if err != nil {
		logger.Output("error finding subpost 1", err)
		return nil, err
	}

	subPost.Content = content
	subPost.Media = media
	subPost.UpdatedAt = time.Now()

	err = s.subPostRepo.Update(ctx, subPost)
	if err != nil {
		logger.Output("error updating subpost 2", err)
		return nil, err
	}

	logger.Output(subPost, nil)
	return subPost, nil
}

func (s *subPostUseCase) DeleteSubPost(ctx context.Context, subPostID primitive.ObjectID) error {
	ctx, span := s.tracer.Start(ctx, "SubPostUseCase.DeleteSubPost")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	logger.Input(subPostID)

	// Find subpost to get parentID
	subPost, err := s.subPostRepo.FindByID(ctx, subPostID)
	if err != nil {
		logger.Output("error finding subpost 1", err)
		return err
	}

	// Find parent post to decrement subpost count
	post, err := s.postRepo.FindByID(ctx, subPost.ParentID)
	if err != nil {
		logger.Output("error finding parent post 2", err)
		return err
	}

	err = s.subPostRepo.Delete(ctx, subPostID)
	if err != nil {
		logger.Output("error deleting subpost 3", err)
		return err
	}

	// Decrement subpost count in parent post
	if post.SubPostCount > 0 {
		post.SubPostCount--
		err = s.postRepo.Update(ctx, post)
		if err != nil {
			logger.Output("error updating parent post 4", err)
			return err
		}
	}

	logger.Output("SubPost deleted successfully", nil)
	return nil
}

func (s *subPostUseCase) FindSubPost(ctx context.Context, subPostID primitive.ObjectID) (*domain.SubPost, error) {
	ctx, span := s.tracer.Start(ctx, "SubPostUseCase.FindSubPost")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	logger.Input(subPostID)

	subPost, err := s.subPostRepo.FindByID(ctx, subPostID)
	if err != nil {
		logger.Output("error finding subpost 1", err)
		return nil, err
	}

	logger.Output(subPost, nil)
	return subPost, nil
}

func (s *subPostUseCase) FindManySubPosts(ctx context.Context, parentID primitive.ObjectID, limit, offset int) ([]domain.SubPost, error) {
	ctx, span := s.tracer.Start(ctx, "SubPostUseCase.FindManySubPosts")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"parentID": parentID,
		"limit":    limit,
		"offset":   offset,
	}
	logger.Input(input)

	subPosts, err := s.subPostRepo.FindByParentID(ctx, parentID, limit, offset)
	if err != nil {
		logger.Output("error finding subposts 1", err)
		return nil, err
	}

	logger.Output(subPosts, nil)
	return subPosts, nil
}

func (s *subPostUseCase) ReorderSubPosts(ctx context.Context, parentID primitive.ObjectID, orders map[primitive.ObjectID]int) error {
	ctx, span := s.tracer.Start(ctx, "SubPostUseCase.ReorderSubPosts")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"parentID": parentID,
		"orders":   orders,
	}
	logger.Input(input)

	err := s.subPostRepo.UpdateOrder(ctx, parentID, orders)
	if err != nil {
		logger.Output("error reordering subposts 1", err)
		return err
	}

	logger.Output("SubPosts reordered successfully", nil)
	return nil
}
