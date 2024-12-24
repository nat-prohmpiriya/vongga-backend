package usecase

import (
	"time"

	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type reactionUseCase struct {
	reactionRepo domain.ReactionRepository
	postRepo     domain.PostRepository
	commentRepo  domain.CommentRepository
}

func NewReactionUseCase(reactionRepo domain.ReactionRepository, postRepo domain.PostRepository, commentRepo domain.CommentRepository) domain.ReactionUseCase {
	return &reactionUseCase{
		reactionRepo: reactionRepo,
		postRepo:     postRepo,
		commentRepo:  commentRepo,
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

	// Get post
	post, err := r.postRepo.FindByID(postID)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Check if reaction already exists and update target
	var target interface{}
	if commentID == nil {
		target = post
	} else {
		comment, err := r.commentRepo.FindByID(*commentID)
		if err != nil {
			logger.LogOutput(nil, err)
			return nil, err
		}
		target = comment
	}

	// Create reaction
	reaction := &domain.Reaction{
		UserID:    userID,
		PostID:    postID,
		CommentID: commentID,
		Type:      reactionType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Update reaction counts
	if commentID == nil {
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
		comment, ok := target.(*domain.Comment)
		if !ok {
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

	// Get post
	post, err := r.postRepo.FindByID(reaction.PostID)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	// Update reaction counts
	if reaction.CommentID == nil {
		if count := post.ReactionCounts[reaction.Type]; count > 0 {
			post.ReactionCounts[reaction.Type]--
			err = r.postRepo.Update(post)
			if err != nil {
				logger.LogOutput(nil, err)
				return err
			}
		}
	} else {
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

	logger.LogOutput("Reaction deleted successfully", nil)
	return nil
}

func (r *reactionUseCase) GetReaction(reactionID primitive.ObjectID) (*domain.Reaction, error) {
	logger := utils.NewLogger("ReactionUseCase.GetReaction")
	logger.LogInput(reactionID)

	reaction, err := r.reactionRepo.FindByID(reactionID)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(reaction, nil)
	return reaction, nil
}

func (r *reactionUseCase) ListReactions(targetID primitive.ObjectID, isComment bool, limit, offset int) ([]domain.Reaction, error) {
	logger := utils.NewLogger("ReactionUseCase.ListReactions")
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
