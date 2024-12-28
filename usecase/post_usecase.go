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
	userRepo    domain.UserRepository
	notificationUseCase domain.NotificationUseCase
}

func NewPostUseCase(
	postRepo domain.PostRepository, 
	subPostRepo domain.SubPostRepository, 
	userRepo domain.UserRepository,
	notificationUseCase domain.NotificationUseCase,
) domain.PostUseCase {
	return &postUseCase{
		postRepo:    postRepo,
		subPostRepo: subPostRepo,
		userRepo:    userRepo,
		notificationUseCase: notificationUseCase,
	}
}

func (p *postUseCase) CreatePost(userID primitive.ObjectID, content string, media []domain.Media, tags []string, location *domain.Location, visibility string, subPosts []domain.SubPostInput) (*domain.Post, error) {
	logger := utils.NewLogger("PostUseCase.CreatePost")
	input := map[string]interface{}{
		"userID":     userID,
		"content":    content,
		"media":      media,
		"tags":       tags,
		"location":   location,
		"visibility": visibility,
		"subPosts":   subPosts,
	}
	logger.LogInput(input)

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

	err := p.postRepo.Create(post)
	if err != nil {
		logger.LogOutput(nil, err)
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
			err := p.subPostRepo.Create(subPost)
			if err != nil {
				logger.LogOutput(nil, err)
				return nil, err
			}
		}
	}

	// Check for mentions in content
	mentions := utils.ExtractMentions(content)
	for _, username := range mentions {
		// Find user by username
		mentionedUser, err := p.userRepo.FindByUsername(username)
		if err != nil {
			logger.LogOutput(nil, err)
			continue // Skip if user not found
		}

		// Don't notify if user mentions themselves
		if mentionedUser.ID == userID {
			continue
		}

		// Create mention notification
		_, err = p.notificationUseCase.CreateNotification(
			mentionedUser.ID,     // recipientID (mentioned user)
			userID,               // senderID (user who mentioned)
			post.ID,              // refID (reference to the post)
			domain.NotificationTypeMention,
			"post",               // refType
			"mentioned you in a post", // message
		)
		if err != nil {
			logger.LogOutput(nil, err)
			// Don't return error here as the post was created successfully
		}
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

	err = p.postRepo.Update(post)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	// Check for mentions in content
	mentions := utils.ExtractMentions(content)
	for _, username := range mentions {
		// Find user by username
		mentionedUser, err := p.userRepo.FindByUsername(username)
		if err != nil {
			logger.LogOutput(nil, err)
			continue // Skip if user not found
		}

		// Don't notify if user mentions themselves
		if mentionedUser.ID == post.UserID {
			continue
		}

		// Create mention notification
		_, err = p.notificationUseCase.CreateNotification(
			mentionedUser.ID,     // recipientID (mentioned user)
			post.UserID,          // senderID (user who mentioned)
			post.ID,              // refID (reference to the post)
			domain.NotificationTypeMention,
			"post",               // refType
			"mentioned you in a post", // message
		)
		if err != nil {
			logger.LogOutput(nil, err)
			// Don't return error here as the post was updated successfully
		}
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

	// Get user data
	user, err := p.userRepo.FindByID(post.UserID.Hex())
	if err != nil {
		logger.LogOutput(nil, err)
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

func (p *postUseCase) ListPosts(userID primitive.ObjectID, limit, offset int, includeSubPosts bool, hasMedia bool, mediaType string) ([]domain.PostWithDetails, error) {
	logger := utils.NewLogger("PostUseCase.ListPosts")

	input := map[string]interface{}{
		"userID":         userID,
		"limit":         limit,
		"offset":        offset,
		"includeSubPosts": includeSubPosts,
		"hasMedia":      hasMedia,
		"mediaType":     mediaType,
	}
	logger.LogInput(input)

	posts, err := p.postRepo.FindByUserID(userID, limit, offset, hasMedia, mediaType)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	var result []domain.PostWithDetails
	for _, post := range posts {
		postCopy := post
		postWithDetails := domain.PostWithDetails{
			Post: &postCopy,
		}

		// Get user details
		user, err := p.userRepo.FindByID(post.UserID.Hex())
		if err != nil {
			logger.LogOutput(nil, fmt.Errorf("error fetching user data for post %s: %w", post.ID.Hex(), err))
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

		// Get sub-posts if requested
		if includeSubPosts {
			subPosts, err := p.subPostRepo.FindByParentID(post.ID, 0, 0)
			if err != nil {
				logger.LogOutput(nil, fmt.Errorf("error fetching subposts for post %s: %w", post.ID.Hex(), err))
				continue
			}
			postWithDetails.SubPosts = subPosts
		}

		result = append(result, postWithDetails)
	}

	logger.LogOutput(result, nil)
	return result, nil
}
