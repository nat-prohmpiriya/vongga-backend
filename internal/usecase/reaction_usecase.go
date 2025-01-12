package usecase

import (
	"errors"
	"time"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type reactionUseCase struct {
	reactionRepo        domain.ReactionRepository
	postRepo            domain.PostRepository
	commentRepo         domain.CommentRepository
	notificationUseCase domain.NotificationUseCase
}

func NewReactionUseCase(
	reactionRepo domain.ReactionRepository,
	postRepo domain.PostRepository,
	commentRepo domain.CommentRepository,
	notificationUseCase domain.NotificationUseCase,
) domain.ReactionUseCase {
	return &reactionUseCase{
		reactionRepo:        reactionRepo,
		postRepo:            postRepo,
		commentRepo:         commentRepo,
		notificationUseCase: notificationUseCase,
	}
}

func (r *reactionUseCase) CreateReaction(userID, postID primitive.ObjectID, commentID *primitive.ObjectID, reactionType string) (*domain.Reaction, error) {
	logger := utils.NewLogger("ReactionUseCase.CreateReaction")
	input := map[string]interface{}{
		"userID":       userID,
		"postID":       postID,
		"commentID":    commentID,
		"reactionType": reactionType,
	}
	logger.LogInput(input)

	// Check if reaction already exists
	existing, err := r.reactionRepo.FindByUserAndTarget(userID, postID, commentID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		logger.LogOutput(nil, err)
		return nil, err
	}

	// If reaction exists, update it
	if existing != nil {
		if existing.Type == reactionType {
			logger.LogOutput(existing, nil)
			return existing, nil
		}
		existing.Type = reactionType
		err = r.reactionRepo.Update(existing)
		if err != nil {
			logger.LogOutput(nil, err)
			return nil, err
		}
		logger.LogOutput(existing, nil)
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

	err = r.reactionRepo.Create(reaction)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Update reaction counts
	if commentID == nil {
		post, err := r.postRepo.FindByID(postID)
		if err != nil {
			logger.LogOutput(nil, err)
			return nil, err
		}

		if post.ReactionCounts == nil {
			post.ReactionCounts = make(map[string]int)
		}
		post.ReactionCounts[reactionType]++

		err = r.postRepo.Update(post)
		if err != nil {
			logger.LogOutput(nil, err)
			return nil, err
		}
	} else {
		comment, err := r.commentRepo.FindByID(*commentID)
		if err != nil {
			logger.LogOutput(nil, err)
			return nil, err
		}

		if comment.ReactionCounts == nil {
			comment.ReactionCounts = make(map[string]int)
		}
		comment.ReactionCounts[reactionType]++

		err = r.commentRepo.Update(comment)
		if err != nil {
			logger.LogOutput(nil, err)
			return nil, err
		}
	}

	// Create notification based on target type
	if commentID != nil {
		// Reaction on comment
		comment, err := r.commentRepo.FindByID(*commentID)
		if err != nil {
			logger.LogOutput(nil, err)
			// Don't return error, just skip notification
		} else if comment.UserID != userID { // Don't notify if user reacts to their own comment
			_, err = r.notificationUseCase.CreateNotification(
				comment.UserID, // recipientID (comment owner)
				userID,         // senderID (user who reacted)
				reaction.ID,    // refID (reference to the reaction)
				domain.NotificationTypeLike,
				"comment",                 // refType
				"reacted to your comment", // message
			)
			if err != nil {
				logger.LogOutput(nil, err)
				// Don't return error here as the reaction was created successfully
			}
		}
	} else {
		// Reaction on post
		post, err := r.postRepo.FindByID(postID)
		if err != nil {
			logger.LogOutput(nil, err)
			// Don't return error, just skip notification
		} else if post.UserID != userID { // Don't notify if user reacts to their own post
			_, err = r.notificationUseCase.CreateNotification(
				post.UserID, // recipientID (post owner)
				userID,      // senderID (user who reacted)
				reaction.ID, // refID (reference to the reaction)
				domain.NotificationTypeLike,
				"post",                 // refType
				"reacted to your post", // message
			)
			if err != nil {
				logger.LogOutput(nil, err)
				// Don't return error here as the reaction was created successfully
			}
		}
	}

	logger.LogOutput(reaction, nil)
	return reaction, nil
}

func (r *reactionUseCase) DeleteReaction(reactionID primitive.ObjectID) error {
	logger := utils.NewLogger("ReactionUseCase.DeleteReaction")
	logger.LogInput(reactionID)

	reaction, err := r.reactionRepo.FindByID(reactionID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Update reaction counts
	if reaction.CommentID == nil {
		// Find post and update reaction count
		post, err := r.postRepo.FindByID(reaction.PostID)
		if err != nil {
			logger.LogOutput(nil, err)
			return err
		}
		if count := post.ReactionCounts[reaction.Type]; count > 0 {
			post.ReactionCounts[reaction.Type]--
			err = r.postRepo.Update(post)
			if err != nil {
				logger.LogOutput(nil, err)
				return err
			}
		}
	} else {
		// Find comment and update reaction count
		comment, err := r.commentRepo.FindByID(*reaction.CommentID)
		if err != nil {
			logger.LogOutput(nil, err)
			return err
		}
		if count := comment.ReactionCounts[reaction.Type]; count > 0 {
			comment.ReactionCounts[reaction.Type]--
			err = r.commentRepo.Update(comment)
			if err != nil {
				logger.LogOutput(nil, err)
				return err
			}
		}
	}

	// Delete reaction
	err = r.reactionRepo.Delete(reactionID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput("Reaction deleted successfully", nil)
	return nil
}

func (r *reactionUseCase) FindReaction(reactionID primitive.ObjectID) (*domain.Reaction, error) {
	logger := utils.NewLogger("ReactionUseCase.FindReaction")
	logger.LogInput(reactionID)

	reaction, err := r.reactionRepo.FindByID(reactionID)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(reaction, nil)
	return reaction, nil
}

func (r *reactionUseCase) FindManyReactions(targetID primitive.ObjectID, isComment bool, limit, offset int) ([]domain.Reaction, error) {
	logger := utils.NewLogger("ReactionUseCase.FindManyReactions")
	input := map[string]interface{}{
		"targetID":  targetID,
		"isComment": isComment,
		"limit":     limit,
		"offset":    offset,
	}
	logger.LogInput(input)

	var reactions []domain.Reaction
	var err error
	if isComment {
		reactions, err = r.reactionRepo.FindByCommentID(targetID, limit, offset)
	} else {
		reactions, err = r.reactionRepo.FindByPostID(targetID, limit, offset)
	}
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(reactions, nil)
	return reactions, nil
}
