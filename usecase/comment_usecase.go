package usecase

import (
	"time"

	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type commentUseCase struct {
	commentRepo        domain.CommentRepository
	postRepo          domain.PostRepository
	notificationUseCase domain.NotificationUseCase
}

func NewCommentUseCase(
	commentRepo domain.CommentRepository,
	postRepo domain.PostRepository,
	notificationUseCase domain.NotificationUseCase,
) domain.CommentUseCase {
	return &commentUseCase{
		commentRepo:        commentRepo,
		postRepo:          postRepo,
		notificationUseCase: notificationUseCase,
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

	// Get post to increment comment count and get post owner
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
		ReactionCounts: make(map[string]int),
		ReplyTo:        replyTo,
	}

	err = c.commentRepo.Create(comment)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	// If this is a reply to another comment, notify the original comment owner
	if replyTo != nil {
		originalComment, err := c.commentRepo.FindByID(*replyTo)
		if err != nil {
			logger.LogOutput(nil, err)
			// Don't return error, just skip notification
		} else {
			// Create notification for reply
			_, err = c.notificationUseCase.CreateNotification(
				originalComment.UserID, // recipientID (original comment owner)
				userID,                 // senderID (user who replied)
				comment.ID,             // refID (reference to the reply)
				domain.NotificationTypeComment,
				"comment",              // refType
				"replied to your comment", // message
			)
			if err != nil {
				logger.LogOutput(nil, err)
				// Don't return error here as the comment was created successfully
			}
		}
	} else {
		// This is a comment on a post, notify the post owner
		// Only notify if the commenter is not the post owner
		if post.UserID != userID {
			_, err = c.notificationUseCase.CreateNotification(
				post.UserID,            // recipientID (post owner)
				userID,                 // senderID (commenter)
				comment.ID,             // refID (reference to the comment)
				domain.NotificationTypeComment,
				"post",                 // refType
				"commented on your post", // message
			)
			if err != nil {
				logger.LogOutput(nil, err)
				// Don't return error here as the comment was created successfully
			}
		}
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
