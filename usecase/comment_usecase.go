package usecase

import (
	"time"

	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type commentUseCase struct {
	commentRepo domain.CommentRepository
	postRepo    domain.PostRepository
}

func NewCommentUseCase(commentRepo domain.CommentRepository, postRepo domain.PostRepository) domain.CommentUseCase {
	return &commentUseCase{
		commentRepo: commentRepo,
		postRepo:    postRepo,
	}
}

func (c *commentUseCase) CreateComment(userID, postID primitive.ObjectID, content string, media []domain.Media, replyTo *primitive.ObjectID) (*domain.Comment, error) {
	logger := utils.NewLogger("CommentUseCase.CreateComment")
	input := map[string]interface{}{
		"userID":  userID,
		"postID":  postID,
		"content": content,
		"media":   media,
		"replyTo": replyTo,
	}
	logger.LogInput(input)

	// Get post to increment comment count
	post, err := c.postRepo.FindByID(postID)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	now := time.Now()
	comment := &domain.Comment{
		BaseModel: domain.BaseModel{
			ID:        primitive.NewObjectID(),
			CreatedAt: now,
			UpdatedAt: now,
			IsActive:  true,
			Version:   1,
		},
		PostID:         postID,
		UserID:         userID,
		Content:        content,
		Media:          media,
		ReplyTo:        replyTo,
		ReactionCounts: make(map[string]int),
	}

	err = c.commentRepo.Create(comment)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Increment comment count in post
	post.CommentCount++
	err = c.postRepo.Update(post)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(comment, nil)
	return comment, nil
}

func (c *commentUseCase) UpdateComment(commentID primitive.ObjectID, content string, media []domain.Media) (*domain.Comment, error) {
	logger := utils.NewLogger("CommentUseCase.UpdateComment")
	input := map[string]interface{}{
		"commentID": commentID,
		"content":   content,
		"media":     media,
	}
	logger.LogInput(input)

	comment, err := c.commentRepo.FindByID(commentID)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	comment.Content = content
	comment.Media = media
	comment.UpdatedAt = time.Now()

	err = c.commentRepo.Update(comment)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(comment, nil)
	return comment, nil
}

func (c *commentUseCase) DeleteComment(commentID primitive.ObjectID) error {
	logger := utils.NewLogger("CommentUseCase.DeleteComment")
	logger.LogInput(commentID)

	// Get comment to get postID
	comment, err := c.commentRepo.FindByID(commentID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Get post to decrement comment count
	post, err := c.postRepo.FindByID(comment.PostID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	err = c.commentRepo.Delete(commentID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Decrement comment count in post
	if post.CommentCount > 0 {
		post.CommentCount--
		err = c.postRepo.Update(post)
		if err != nil {
			logger.LogOutput(nil, err)
			return err
		}
	}

	logger.LogOutput("Comment deleted successfully", nil)
	return nil
}

func (c *commentUseCase) GetComment(commentID primitive.ObjectID) (*domain.Comment, error) {
	logger := utils.NewLogger("CommentUseCase.GetComment")
	logger.LogInput(commentID)

	comment, err := c.commentRepo.FindByID(commentID)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(comment, nil)
	return comment, nil
}

func (c *commentUseCase) ListComments(postID primitive.ObjectID, limit, offset int) ([]domain.Comment, error) {
	logger := utils.NewLogger("CommentUseCase.ListComments")
	input := map[string]interface{}{
		"postID": postID,
		"limit":  limit,
		"offset": offset,
	}
	logger.LogInput(input)

	comments, err := c.commentRepo.FindByPostID(postID, limit, offset)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(comments, nil)
	return comments, nil
}
