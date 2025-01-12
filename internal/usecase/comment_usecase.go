package usecase

import (
	"context"
	"time"

	"vongga_api/internal/domain"
	"vongga_api/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
)

type commentUseCase struct {
	commentRepo         domain.CommentRepository
	postRepo            domain.PostRepository
	notificationUseCase domain.NotificationUseCase
	userRepo            domain.UserRepository
	tracer              trace.Tracer
}

func NewCommentUseCase(
	commentRepo domain.CommentRepository,
	postRepo domain.PostRepository,
	notificationUseCase domain.NotificationUseCase,
	userRepo domain.UserRepository,
	tracer trace.Tracer,
) domain.CommentUseCase {
	return &commentUseCase{
		commentRepo:         commentRepo,
		postRepo:            postRepo,
		notificationUseCase: notificationUseCase,
		userRepo:            userRepo,
		tracer:              tracer,
	}
}

func (c *commentUseCase) CreateComment(ctx context.Context, userID, postID primitive.ObjectID, content string, media []domain.Media, replyTo *primitive.ObjectID) (*domain.Comment, error) {
	ctx, span := c.tracer.Start(ctx, "CommentUseCase.CreateComment")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	input := map[string]interface{}{
		"userID":  userID,
		"postID":  postID,
		"content": content,
		"media":   media,
		"replyTo": replyTo,
	}
	logger.Input(input)

	// Find post to increment comment count and get post owner
	post, err := c.postRepo.FindByID(ctx, postID)
	if err != nil {
		logger.Output("error finding post 1", err)
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

	err = c.commentRepo.Create(ctx, comment)
	if err != nil {
		logger.Output("error creating comment 3", err)
		return nil, err
	}

	// Check for mentions in content
	mentions := utils.ExtractMentions(content)
	for _, username := range mentions {
		// Find user by username
		mentionedUser, err := c.userRepo.FindByUsername(ctx, username)
		if err != nil {
			logger.Output("error finding user 2", err)
			continue // Skip if user not found
		}

		// Don't notify if user mentions themselves
		if mentionedUser.ID == userID {
			continue
		}

		// Create mention notification
		_, err = c.notificationUseCase.CreateNotification(
			ctx,
			mentionedUser.ID, // recipientID (mentioned user)
			userID,           // senderID (user who mentioned)
			comment.ID,       // refID (reference to the comment)
			domain.NotificationTypeMention,
			"comment",                    // refType
			"mentioned you in a comment", // message
		)
		if err != nil {
			logger.Output("error creating mention notification", err)
			// Don't return error here as the comment was created successfully
		}
	}

	// If this is a reply to another comment, notify the original comment owner
	if replyTo != nil {
		originalComment, err := c.commentRepo.FindByID(ctx, *replyTo)
		if err != nil {
			logger.Output("error finding original comment", err)
			// Don't return error, just skip notification
		} else {
			// Create notification for reply
			_, err = c.notificationUseCase.CreateNotification(
				ctx,
				originalComment.UserID, // recipientID (original comment owner)
				userID,                 // senderID (user who replied)
				comment.ID,             // refID (reference to the reply)
				domain.NotificationTypeComment,
				"comment",                 // refType
				"replied to your comment", // message
			)
			if err != nil {
				logger.Output("error creating reply notification", err)
				// Don't return error here as the comment was created successfully
			}
		}
	} else {
		// This is a comment on a post, notify the post owner
		// Only notify if the commenter is not the post owner
		if post.UserID != userID {
			_, err = c.notificationUseCase.CreateNotification(
				ctx,
				post.UserID, // recipientID (post owner)
				userID,      // senderID (commenter)
				comment.ID,  // refID (reference to the comment)
				domain.NotificationTypeComment,
				"post",                   // refType
				"commented on your post", // message
			)
			if err != nil {
				logger.Output("error creating post notification", err)
				// Don't return error here as the comment was created successfully
			}
		}
	}

	// Increment comment count in post
	post.CommentCount++
	err = c.postRepo.Update(ctx, post)
	if err != nil {
		logger.Output("error updating post 4", err)
		return nil, err
	}

	logger.Output(comment, nil)
	return comment, nil
}

func (c *commentUseCase) UpdateComment(ctx context.Context, commentID primitive.ObjectID, content string, media []domain.Media) (*domain.Comment, error) {
	ctx, span := c.tracer.Start(ctx, "CommentUseCase.UpdateComment")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	input := map[string]interface{}{
		"commentID": commentID,
		"content":   content,
		"media":     media,
	}
	logger.Input(input)

	comment, err := c.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		logger.Output("error finding comment 1", err)
		return nil, err
	}

	comment.Content = content
	comment.Media = media
	comment.UpdatedAt = time.Now()

	err = c.commentRepo.Update(ctx, comment)
	if err != nil {
		logger.Output("error updating comment 2", err)
		return nil, err
	}

	logger.Output(comment, nil)
	return comment, nil
}

func (c *commentUseCase) DeleteComment(ctx context.Context, commentID primitive.ObjectID) error {
	ctx, span := c.tracer.Start(ctx, "CommentUseCase.DeleteComment")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(commentID)

	// Find comment to get postID
	comment, err := c.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		logger.Output("error finding comment 1", err)
		return err
	}

	// Find post to decrement comment count
	post, err := c.postRepo.FindByID(ctx, comment.PostID)
	if err != nil {
		logger.Output("error finding post 2", err)
		return err
	}

	err = c.commentRepo.Delete(ctx, commentID)
	if err != nil {
		logger.Output("error deleting comment 3", err)
		return err
	}

	// Decrement comment count in post
	if post.CommentCount > 0 {
		post.CommentCount--
		err = c.postRepo.Update(ctx, post)
		if err != nil {
			logger.Output("error updating post 4", err)
			return err
		}
	}

	logger.Output("Comment deleted successfully", nil)
	return nil
}

func (c *commentUseCase) FindComment(ctx context.Context, commentID primitive.ObjectID) (*domain.Comment, error) {
	ctx, span := c.tracer.Start(ctx, "CommentUseCase.FindComment")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(commentID)

	comment, err := c.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		logger.Output("error finding comment 1", err)
		return nil, err
	}

	logger.Output(comment, nil)
	return comment, nil
}

func (c *commentUseCase) FindManyComments(ctx context.Context, postID primitive.ObjectID, limit, offset int) ([]domain.Comment, error) {
	ctx, span := c.tracer.Start(ctx, "CommentUseCase.FindManyComments")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	input := map[string]interface{}{
		"postID": postID,
		"limit":  limit,
		"offset": offset,
	}
	logger.Input(input)

	comments, err := c.commentRepo.FindByPostID(ctx, postID, limit, offset)
	if err != nil {
		logger.Output("error finding comments 1", err)
		return nil, err
	}

	logger.Output(comments, nil)
	return comments, nil
}
