package usecase

import (
	"time"

	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type postUseCase struct {
	postRepo    domain.PostRepository
	subPostRepo domain.SubPostRepository
}

func NewPostUseCase(postRepo domain.PostRepository, subPostRepo domain.SubPostRepository) domain.PostUseCase {
	return &postUseCase{
		postRepo:    postRepo,
		subPostRepo: subPostRepo,
	}
}

func (p *postUseCase) CreatePost(userID primitive.ObjectID, content string, media []domain.Media, tags []string, location *domain.Location, visibility string) (*domain.Post, error) {
	logger := utils.NewLogger("PostUseCase.CreatePost")
	input := map[string]interface{}{
		"userID":     userID,
		"content":    content,
		"media":      media,
		"tags":       tags,
		"location":   location,
		"visibility": visibility,
	}
	logger.LogInput(input)

	post := &domain.Post{
		UserID:         userID,
		Content:        content,
		Media:          media,
		Tags:           tags,
		Location:       location,
		Visibility:     visibility,
		ReactionCounts: make(map[string]int),
		CommentCount:   0,
		SubPostCount:   0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		IsEdited:       false,
		EditHistory:    make([]domain.EditLog, 0),
	}

	err := p.postRepo.Create(post)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(post, nil)
	return post, nil
}

func (p *postUseCase) UpdatePost(postID primitive.ObjectID, content string, media []domain.Media, tags []string, location *domain.Location, visibility string) (*domain.Post, error) {
	logger := utils.NewLogger("PostUseCase.UpdatePost")
	input := map[string]interface{}{
		"postID":     postID,
		"content":    content,
		"media":      media,
		"tags":       tags,
		"location":   location,
		"visibility": visibility,
	}
	logger.LogInput(input)

	post, err := p.postRepo.FindByID(postID)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Create edit log
	editLog := domain.EditLog{
		Content:   post.Content,
		Media:     post.Media,
		Tags:      post.Tags,
		Location:  post.Location,
		EditedAt:  time.Now(),
	}
	post.EditHistory = append(post.EditHistory, editLog)

	// Update post
	post.Content = content
	post.Media = media
	post.Tags = tags
	post.Location = location
	post.Visibility = visibility
	post.UpdatedAt = time.Now()
	post.IsEdited = true

	err = p.postRepo.Update(post)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(post, nil)
	return post, nil
}

func (p *postUseCase) DeletePost(postID primitive.ObjectID) error {
	logger := utils.NewLogger("PostUseCase.DeletePost")
	logger.LogInput(postID)

	// Delete all subposts first
	subPosts, err := p.subPostRepo.FindByParentID(postID, 0, 0) // Get all subposts
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}
	for _, subPost := range subPosts {
		err = p.subPostRepo.Delete(subPost.ID)
		if err != nil {
			logger.LogOutput(nil, err)
			return err
		}
	}

	// Delete the post
	err = p.postRepo.Delete(postID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput("Post and all related subposts deleted successfully", nil)
	return nil
}

func (p *postUseCase) GetPost(postID primitive.ObjectID, includeSubPosts bool) (*domain.PostWithDetails, error) {
	logger := utils.NewLogger("PostUseCase.GetPost")
	input := map[string]interface{}{
		"postID":          postID,
		"includeSubPosts": includeSubPosts,
	}
	logger.LogInput(input)

	post, err := p.postRepo.FindByID(postID)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	result := &domain.PostWithDetails{
		Post: post,
	}

	if includeSubPosts {
		subPosts, err := p.subPostRepo.FindByParentID(postID, 0, 0) // Get all subposts
		if err != nil {
			logger.LogOutput(nil, err)
			return nil, err
		}
		result.SubPosts = subPosts
	}

	logger.LogOutput(result, nil)
	return result, nil
}

func (p *postUseCase) ListPosts(userID primitive.ObjectID, limit, offset int, includeSubPosts bool) ([]domain.PostWithDetails, error) {
	logger := utils.NewLogger("PostUseCase.ListPosts")
	input := map[string]interface{}{
		"userID":          userID,
		"limit":           limit,
		"offset":          offset,
		"includeSubPosts": includeSubPosts,
	}
	logger.LogInput(input)

	posts, err := p.postRepo.FindByUserID(userID, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	result := make([]domain.PostWithDetails, len(posts))
	for i, post := range posts {
		result[i].Post = &post
		if includeSubPosts {
			subPosts, err := p.subPostRepo.FindByParentID(post.ID, 0, 0) // Get all subposts
			if err != nil {
				logger.LogOutput(nil, err)
				return nil, err
			}
			result[i].SubPosts = subPosts
		}
	}

	logger.LogOutput(result, nil)
	return result, nil
}
