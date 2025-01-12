package usecase

import (
	"context"
	"time"

	"vongga-api/internal/domain"
	"vongga-api/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
)

type postUseCase struct {
	postRepo            domain.PostRepository
	subPostRepo         domain.SubPostRepository
	userRepo            domain.UserRepository
	notificationUseCase domain.NotificationUseCase
	tracer              trace.Tracer
}

func NewPostUseCase(
	postRepo domain.PostRepository,
	subPostRepo domain.SubPostRepository,
	userRepo domain.UserRepository,
	notificationUseCase domain.NotificationUseCase,
	tracer trace.Tracer,
) domain.PostUseCase {
	return &postUseCase{
		postRepo:            postRepo,
		subPostRepo:         subPostRepo,
		userRepo:            userRepo,
		notificationUseCase: notificationUseCase,
		tracer:              tracer,
	}
}

func (p *postUseCase) CreatePost(ctx context.Context, userID primitive.ObjectID, content string, media []domain.Media, tags []string, location *domain.Location, visibility string, subPosts []domain.SubPostInput) (*domain.Post, error) {
	ctx, span := p.tracer.Start(ctx, "PostUseCase.CreatePost")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"userID":     userID,
		"content":    content,
		"media":      media,
		"tags":       tags,
		"location":   location,
		"visibility": visibility,
		"subPosts":   subPosts,
	}
	logger.Input(input)

	now := time.Now()
	post := &domain.Post{
		BaseModel: domain.BaseModel{
			ID:        primitive.NewObjectID(),
			CreatedAt: now,
			UpdatedAt: now,
			IsActive:  true,
			Version:   1,
		},
		UserID:         userID,
		Content:        content,
		Media:          media,
		Tags:           tags,
		Location:       location,
		Visibility:     visibility,
		ReactionCounts: make(map[string]int),
		CommentCount:   0,
		SubPostCount:   len(subPosts),
		IsEdited:       false,
		EditHistory:    make([]domain.EditLog, 0),
	}

	err := p.postRepo.Create(ctx, post)
	if err != nil {
		logger.Output("error creating post 1", err)
		return nil, err
	}

	// Create subposts if any
	if len(subPosts) > 0 {
		for _, subPostInput := range subPosts {
			subPost := &domain.SubPost{
				BaseModel: domain.BaseModel{
					ID:        primitive.NewObjectID(),
					CreatedAt: now,
					UpdatedAt: now,
					IsActive:  true,
					Version:   1,
				},
				ParentID:       post.ID,
				UserID:         userID,
				Content:        subPostInput.Content,
				Media:          subPostInput.Media,
				ReactionCounts: make(map[string]int),
				CommentCount:   0,
				Order:          subPostInput.Order,
			}
			err := p.subPostRepo.Create(ctx, subPost)
			if err != nil {
				logger.Output("error creating subpost 2", err)
				return nil, err
			}
		}
	}

	// Check for mentions in content
	mentions := utils.ExtractMentions(content)
	for _, username := range mentions {
		// Find user by username
		mentionedUser, err := p.userRepo.FindByUsername(ctx, username)
		if err != nil {
			logger.Output("error finding user 3", err)
			continue // Skip if user not found
		}

		// Don't notify if user mentions themselves
		if mentionedUser.ID == userID {
			continue
		}

		// Create mention notification
		_, err = p.notificationUseCase.CreateNotification(
			ctx,
			mentionedUser.ID, // recipientID (mentioned user)
			userID,           // senderID (user who mentioned)
			post.ID,          // refID (reference to the post)
			domain.NotificationTypeMention,
			"post",                    // refType
			"mentioned you in a post", // message
		)
		if err != nil {
			logger.Output("error creating notification 4", err)
			// Don't return error here as the post was created successfully
		}
	}

	logger.Output(post, nil)
	return post, nil
}

func (p *postUseCase) UpdatePost(ctx context.Context, postID primitive.ObjectID, content string, media []domain.Media, tags []string, location *domain.Location, visibility string) (*domain.Post, error) {
	ctx, span := p.tracer.Start(ctx, "PostUseCase.UpdatePost")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"postID":     postID,
		"content":    content,
		"media":      media,
		"tags":       tags,
		"location":   location,
		"visibility": visibility,
	}
	logger.Input(input)

	post, err := p.postRepo.FindByID(ctx, postID)
	if err != nil {
		logger.Output("error finding post 1", err)
		return nil, err
	}

	// Create edit log
	editLog := domain.EditLog{
		Content:  post.Content,
		Media:    post.Media,
		Tags:     post.Tags,
		Location: post.Location,
		EditedAt: time.Now(),
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

	err = p.postRepo.Update(ctx, post)
	if err != nil {
		logger.Output("error updating post 2", err)
		return nil, err
	}

	// Check for mentions in content
	mentions := utils.ExtractMentions(content)
	for _, username := range mentions {
		// Find user by username
		mentionedUser, err := p.userRepo.FindByUsername(ctx, username)
		if err != nil {
			logger.Output("error finding user 3", err)
			continue // Skip if user not found
		}

		// Don't notify if user mentions themselves
		if mentionedUser.ID == post.UserID {
			continue
		}

		// Create mention notification
		_, err = p.notificationUseCase.CreateNotification(
			ctx,
			mentionedUser.ID, // recipientID (mentioned user)
			post.UserID,      // senderID (user who mentioned)
			post.ID,          // refID (reference to the post)
			domain.NotificationTypeMention,
			"post",                    // refType
			"mentioned you in a post", // message
		)
		if err != nil {
			logger.Output("error creating notification 4", err)
			// Don't return error here as the post was updated successfully
		}
	}

	logger.Output(post, nil)
	return post, nil
}

func (p *postUseCase) DeletePost(ctx context.Context, postID primitive.ObjectID) error {
	ctx, span := p.tracer.Start(ctx, "PostUseCase.DeletePost")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	logger.Input(postID)

	// Delete all subposts first
	subPosts, err := p.subPostRepo.FindByParentID(ctx, postID, 0, 0) // Find all subposts
	if err != nil {
		logger.Output("error finding subposts 1", err)
		return err
	}
	for _, subPost := range subPosts {
		err = p.subPostRepo.Delete(ctx, subPost.ID)
		if err != nil {
			logger.Output("error deleting subpost 2", err)
			return err
		}
	}

	// Delete the post
	err = p.postRepo.Delete(ctx, postID)
	if err != nil {
		logger.Output("error deleting post 3", err)
		return err
	}

	logger.Output("Post and all related subposts deleted successfully", nil)
	return nil
}

func (p *postUseCase) FindPost(ctx context.Context, postID primitive.ObjectID, includeSubPosts bool) (*domain.PostWithDetails, error) {
	ctx, span := p.tracer.Start(ctx, "PostUseCase.FindPost")
	defer span.End()
	logger := utils.NewTraceLogger(span)

	input := map[string]interface{}{
		"postID":          postID,
		"includeSubPosts": includeSubPosts,
	}
	logger.Input(input)

	post, err := p.postRepo.FindByID(ctx, postID)
	if err != nil {
		logger.Output("error finding post 1", err)
		return nil, err
	}

	// Find user data
	user, err := p.userRepo.FindByID(ctx, post.UserID.Hex())
	if err != nil {
		logger.Output("error finding user 2", err)
		return nil, err
	}

	// Map to PostUser with limited fields
	postUser := &domain.PostUser{
		ID:           user.ID,
		Username:     user.Username,
		DisplayName:  user.DisplayName,
		PhotoProfile: user.PhotoProfile,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
	}

	result := &domain.PostWithDetails{
		Post: post,
		User: postUser,
	}

	if includeSubPosts {
		subPosts, err := p.subPostRepo.FindByParentID(ctx, postID, 0, 0) // Find all subposts
		if err != nil {
			logger.Output("error finding subposts 3", err)
			return nil, err
		}
		result.SubPosts = subPosts
	}

	logger.Output(result, nil)
	return result, nil
}

func (p *postUseCase) FindManyPosts(ctx context.Context, userID primitive.ObjectID, limit, offset int, includeSubPosts bool, hasMedia bool, mediaType string) ([]domain.PostWithDetails, error) {
	ctx, span := p.tracer.Start(ctx, "PostUseCase.FindManyPosts")

	input := map[string]interface{}{
		"userID":          userID,
		"limit":           limit,
		"offset":          offset,
		"includeSubPosts": includeSubPosts,
		"hasMedia":        hasMedia,
		"mediaType":       mediaType,
	}
	logger := utils.NewTraceLogger(span)
	logger.Input(input)

	posts, err := p.postRepo.FindByUserID(ctx, userID, limit, offset, hasMedia, mediaType)
	if err != nil {
		logger.Output("error finding posts 1", err)
		return nil, err
	}

	var result []domain.PostWithDetails
	for _, post := range posts {
		postCopy := post
		postWithDetails := domain.PostWithDetails{
			Post: &postCopy,
		}

		// Find user details
		user, err := p.userRepo.FindByID(ctx, post.UserID.Hex())
		if err != nil {
			logger.Output("error finding user 2", err)
			continue
		}

		postWithDetails.User = &domain.PostUser{
			ID:           user.ID,
			Username:     user.Username,
			DisplayName:  user.DisplayName,
			PhotoProfile: user.PhotoProfile,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
		}

		// Find sub-posts if requested
		if includeSubPosts {
			subPosts, err := p.subPostRepo.FindByParentID(ctx, post.ID, 0, 0)
			if err != nil {
				logger.Output("error finding subposts 3", err)
				continue
			}
			postWithDetails.SubPosts = subPosts
		}

		result = append(result, postWithDetails)
	}

	logger.Output(result, nil)
	return result, nil
}
