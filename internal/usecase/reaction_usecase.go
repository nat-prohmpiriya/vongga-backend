package usecase

import (
	"context"
	"errors"
	"time"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
)

type reactionUseCase struct {
	reactionRepo        domain.ReactionRepository
	postRepo            domain.PostRepository
	commentRepo         domain.CommentRepository
	notificationUseCase domain.NotificationUseCase
	tracer              trace.Tracer
}

func NewReactionUseCase(
	reactionRepo domain.ReactionRepository,
	postRepo domain.PostRepository,
	commentRepo domain.CommentRepository,
	notificationUseCase domain.NotificationUseCase,
	tracer trace.Tracer,
) domain.ReactionUseCase {
	return &reactionUseCase{
		reactionRepo:        reactionRepo,
		postRepo:            postRepo,
		commentRepo:         commentRepo,
		notificationUseCase: notificationUseCase,
		tracer:              tracer,
	}
}

func (r *reactionUseCase) CreateReaction(ctx context.Context, userID, postID primitive.ObjectID, commentID *primitive.ObjectID, reactionType string) (*domain.Reaction, error) {
	ctx, span := r.tracer.Start(ctx, "ReactionUseCase.CreateReaction")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"userID":       userID,
		"postID":       postID,
		"commentID":    commentID,
		"reactionType": reactionType,
	}
	logger.Input(input)

	// Check if reaction already exists
	existing, err := r.reactionRepo.FindByUserAndTarget(ctx, userID, postID, commentID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		logger.Output("error finding existing reaction 1", err)
		return nil, err
	}

	// If reaction exists, update it
	if existing != nil {
		if existing.Type == reactionType {
			logger.Output(existing, nil)
			return existing, nil
		}
		existing.Type = reactionType
		err = r.reactionRepo.Update(ctx, existing)
		if err != nil {
			logger.Output("error updating existing reaction 2", err)
			return nil, err
		}
		logger.Output(existing, nil)
		return existing, nil
	}

	// Create new reaction
	now := time.Now()
	reaction := &domain.Reaction{
		BaseModel: domain.BaseModel{
			ID:        primitive.NewObjectID(),
			CreatedAt: now,
			UpdatedAt: now,
			IsActive:  true,
		},
		PostID:    postID,
		CommentID: commentID,
		UserID:    userID,
		Type:      reactionType,
	}

	err = r.reactionRepo.Create(ctx, reaction)
	if err != nil {
		logger.Output("error creating reaction 3", err)
		return nil, err
	}

	// Update reaction counts
	if commentID == nil {
		post, err := r.postRepo.FindByID(ctx, postID)
		if err != nil {
			logger.Output("error finding post 4", err)
			return nil, err
		}

		if post.ReactionCounts == nil {
			post.ReactionCounts = make(map[string]int)
		}
		post.ReactionCounts[reactionType]++

		err = r.postRepo.Update(ctx, post)
		if err != nil {
			logger.Output("error updating post reaction counts 5", err)
			return nil, err
		}
	} else {
		comment, err := r.commentRepo.FindByID(ctx, *commentID)
		if err != nil {
			logger.Output("error finding comment 6", err)
			return nil, err
		}

		if comment.ReactionCounts == nil {
			comment.ReactionCounts = make(map[string]int)
		}
		comment.ReactionCounts[reactionType]++

		err = r.commentRepo.Update(ctx, comment)
		if err != nil {
			logger.Output("error updating comment reaction counts 7", err)
			return nil, err
		}
	}

	// Create notification based on target type
	if commentID != nil {
		// Reaction on comment
		comment, err := r.commentRepo.FindByID(ctx, *commentID)
		if err != nil {
			logger.Output("error finding comment 2", err)
			// Don't return error, just skip notification
		} else if comment.UserID != userID { // Don't notify if user reacts to their own comment
			_, err = r.notificationUseCase.CreateNotification(
				ctx,
				comment.UserID, // recipientID (comment owner)
				userID,         // senderID (user who reacted)
				reaction.ID,    // refID (reference to the reaction)
				domain.NotificationTypeLike,
				"comment",                 // refType
				"reacted to your comment", // message
			)
			if err != nil {
				logger.Output("error creating notification 8", err)
				// Don't return error here as the reaction was created successfully
			}
		}
	} else {
		// Reaction on post
		post, err := r.postRepo.FindByID(ctx, postID)
		if err != nil {
			logger.Output("error finding post 9", err)
			// Don't return error, just skip notification
		} else if post.UserID != userID { // Don't notify if user reacts to their own post
			_, err = r.notificationUseCase.CreateNotification(
				ctx,
				post.UserID, // recipientID (post owner)
				userID,      // senderID (user who reacted)
				reaction.ID, // refID (reference to the reaction)
				domain.NotificationTypeLike,
				"post",                 // refType
				"reacted to your post", // message
			)
			if err != nil {
				logger.Output("error creating notification", err)
				// Don't return error here as the reaction was created successfully
			}
		}
	}

	logger.Output(reaction, nil)
	return reaction, nil
}

func (r *reactionUseCase) DeleteReaction(ctx context.Context, reactionID primitive.ObjectID) error {
	ctx, span := r.tracer.Start(ctx, "ReactionUseCase.DeleteReaction")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	logger.Input(reactionID)

	reaction, err := r.reactionRepo.FindByID(ctx, reactionID)
	if err != nil {
		logger.Output("error finding reaction 1", err)
		return err
	}

	// Update reaction counts
	if reaction.CommentID == nil {
		// Find post and update reaction count
		post, err := r.postRepo.FindByID(ctx, reaction.PostID)
		if err != nil {
			logger.Output("error finding post 2", err)
			return err
		}
		if count := post.ReactionCounts[reaction.Type]; count > 0 {
			post.ReactionCounts[reaction.Type]--
			err = r.postRepo.Update(ctx, post)
			if err != nil {
				logger.Output("error updating post reaction counts 3", err)
				return err
			}
		}
	} else {
		// Find comment and update reaction count
		comment, err := r.commentRepo.FindByID(ctx, *reaction.CommentID)
		if err != nil {
			logger.Output("error finding comment 4", err)
			return err
		}
		if count := comment.ReactionCounts[reaction.Type]; count > 0 {
			comment.ReactionCounts[reaction.Type]--
			err = r.commentRepo.Update(ctx, comment)
			if err != nil {
				logger.Output("error updating comment reaction counts 5", err)
				return err
			}
		}
	}

	// Delete reaction
	err = r.reactionRepo.Delete(ctx, reactionID)
	if err != nil {
		logger.Output("error deleting reaction 6", err)
		return err
	}

	logger.Output("Reaction deleted successfully", nil)
	return nil
}

func (r *reactionUseCase) FindReaction(ctx context.Context, reactionID primitive.ObjectID) (*domain.Reaction, error) {
	ctx, span := r.tracer.Start(ctx, "ReactionUseCase.FindReaction")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	logger.Input(reactionID)

	reaction, err := r.reactionRepo.FindByID(ctx, reactionID)
	if err != nil {
		logger.Output("error finding reaction 1", err)
		return nil, err
	}

	logger.Output(reaction, nil)
	return reaction, nil
}

func (r *reactionUseCase) FindManyReactions(ctx context.Context, targetID primitive.ObjectID, isComment bool, limit, offset int) ([]domain.Reaction, error) {
	ctx, span := r.tracer.Start(ctx, "ReactionUseCase.FindManyReactions")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"targetID":  targetID,
		"isComment": isComment,
		"limit":     limit,
		"offset":    offset,
	}
	logger.Input(input)

	var reactions []domain.Reaction
	var err error
	if isComment {
		reactions, err = r.reactionRepo.FindByCommentID(ctx, targetID, limit, offset)
	} else {
		reactions, err = r.reactionRepo.FindByPostID(ctx, targetID, limit, offset)
	}
	if err != nil {
		logger.Output("error finding reactions 1", err)
		return nil, err
	}

	logger.Output(reactions, nil)
	return reactions, nil
}
