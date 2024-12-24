package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrPostNotFound = errors.New("post not found")
var ErrHasSubPosts = errors.New("post has subposts")

func TestPostUseCase_CreatePost(t *testing.T) {
	mockPostRepo := new(mocks.PostRepository)
	mockSubPostRepo := new(mocks.SubPostRepository)
	postUseCase := NewPostUseCase(mockPostRepo, mockSubPostRepo)

	userID := primitive.NewObjectID()
	content := "Test post content"
	media := []domain.Media{
		{
			Type: "image",
			URL:  "https://example.com/image.jpg",
		},
	}
	tags := []string{"test", "example"}
	location := &domain.Location{
		Coordinates: []float64{100.5018, 13.7563},
		Type:        "Point",
		PlaceName:   "Bangkok",
	}
	visibility := "public"

	t.Run("Success", func(t *testing.T) {
		mockPostRepo.On("Create", mock.AnythingOfType("*domain.Post")).Return(nil).Once()

		post, err := postUseCase.CreatePost(userID, content, media, tags, location, visibility)

		assert.NoError(t, err)
		assert.NotNil(t, post)
		assert.Equal(t, userID, post.UserID)
		assert.Equal(t, content, post.Content)
		assert.Equal(t, media, post.Media)
		assert.Equal(t, tags, post.Tags)
		assert.Equal(t, location, post.Location)
		assert.Equal(t, visibility, post.Visibility)
		assert.False(t, post.IsEdited)
		assert.Empty(t, post.EditHistory)
		assert.Zero(t, post.CommentCount)
		assert.Zero(t, post.SubPostCount)
		assert.Zero(t, post.ShareCount)
		assert.Zero(t, post.ViewCount)

		mockPostRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockPostRepo.On("Create", mock.AnythingOfType("*domain.Post")).Return(assert.AnError).Once()

		post, err := postUseCase.CreatePost(userID, content, media, tags, location, visibility)

		assert.Error(t, err)
		assert.Nil(t, post)
		mockPostRepo.AssertExpectations(t)
	})
}

func TestPostUseCase_UpdatePost(t *testing.T) {
	mockPostRepo := new(mocks.PostRepository)
	mockSubPostRepo := new(mocks.SubPostRepository)
	postUseCase := NewPostUseCase(mockPostRepo, mockSubPostRepo)

	postID := primitive.NewObjectID()
	content := "Updated post content"
	media := []domain.Media{
		{
			Type: "image",
			URL:  "https://example.com/updated-image.jpg",
		},
	}
	tags := []string{"updated", "test"}
	location := &domain.Location{
		Coordinates: []float64{100.5018, 13.7563},
		Type:        "Point",
		PlaceName:   "Bangkok",
	}
	visibility := "public"

	existingPost := &domain.Post{
		ID:          postID,
		Content:     "Original content",
		Media:       []domain.Media{},
		Tags:        []string{"original"},
		Location:    nil,
		Visibility:  "private",
		IsEdited:    false,
		EditHistory: []domain.EditLog{},
		CreatedAt:   time.Now().Add(-time.Hour),
		UpdatedAt:   time.Now().Add(-time.Hour),
	}

	t.Run("Success", func(t *testing.T) {
		mockPostRepo.On("FindByID", postID).Return(existingPost, nil).Once()
		mockPostRepo.On("Update", mock.MatchedBy(func(post *domain.Post) bool {
			return post.ID == postID &&
				post.Content == content &&
				len(post.Media) == len(media) &&
				post.Media[0].Type == media[0].Type &&
				post.Media[0].URL == media[0].URL &&
				len(post.Tags) == len(tags) &&
				post.Tags[0] == tags[0] &&
				post.Tags[1] == tags[1] &&
				post.Location.Type == location.Type &&
				post.Location.Coordinates[0] == location.Coordinates[0] &&
				post.Location.Coordinates[1] == location.Coordinates[1] &&
				post.Location.PlaceName == location.PlaceName &&
				post.Visibility == visibility &&
				post.IsEdited &&
				len(post.EditHistory) == 1 &&
				post.EditHistory[0].Content == existingPost.Content &&
				len(post.EditHistory[0].Media) == len(existingPost.Media) &&
				len(post.EditHistory[0].Tags) == len(existingPost.Tags) &&
				post.EditHistory[0].Location == existingPost.Location
		})).Return(nil).Once()

		updatedPost, err := postUseCase.UpdatePost(postID, content, media, tags, location, visibility)

		assert.NoError(t, err)
		assert.NotNil(t, updatedPost)
		assert.Equal(t, content, updatedPost.Content)
		assert.Equal(t, media, updatedPost.Media)
		assert.Equal(t, tags, updatedPost.Tags)
		assert.Equal(t, location, updatedPost.Location)
		assert.Equal(t, visibility, updatedPost.Visibility)
		assert.True(t, updatedPost.IsEdited)
		assert.Len(t, updatedPost.EditHistory, 1)
		assert.Equal(t, existingPost.Content, updatedPost.EditHistory[0].Content)
		assert.Equal(t, existingPost.Media, updatedPost.EditHistory[0].Media)
		assert.Equal(t, existingPost.Tags, updatedPost.EditHistory[0].Tags)
		assert.Equal(t, existingPost.Location, updatedPost.EditHistory[0].Location)

		mockPostRepo.AssertExpectations(t)
	})

	t.Run("Post Not Found", func(t *testing.T) {
		mockPostRepo.On("FindByID", postID).Return(nil, ErrPostNotFound).Once()

		updatedPost, err := postUseCase.UpdatePost(postID, content, media, tags, location, visibility)

		assert.Error(t, err)
		assert.Equal(t, ErrPostNotFound, err)
		assert.Nil(t, updatedPost)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("Update Error", func(t *testing.T) {
		mockPostRepo.On("FindByID", postID).Return(existingPost, nil).Once()
		mockPostRepo.On("Update", mock.AnythingOfType("*domain.Post")).Return(assert.AnError).Once()

		updatedPost, err := postUseCase.UpdatePost(postID, content, media, tags, location, visibility)

		assert.Error(t, err)
		assert.Nil(t, updatedPost)
		mockPostRepo.AssertExpectations(t)
	})
}

func TestPostUseCase_DeletePost(t *testing.T) {
	mockPostRepo := new(mocks.PostRepository)
	mockSubPostRepo := new(mocks.SubPostRepository)
	postUseCase := NewPostUseCase(mockPostRepo, mockSubPostRepo)

	postID := primitive.NewObjectID()

	t.Run("Success", func(t *testing.T) {
		mockSubPostRepo.On("FindByParentID", postID, 0, 0).Return([]domain.SubPost{}, nil).Once()
		mockPostRepo.On("Delete", postID).Return(nil).Once()

		err := postUseCase.DeletePost(postID)

		assert.NoError(t, err)
		mockPostRepo.AssertExpectations(t)
		mockSubPostRepo.AssertExpectations(t)
	})

	t.Run("Post Not Found", func(t *testing.T) {
		mockSubPostRepo.On("FindByParentID", postID, 0, 0).Return([]domain.SubPost{}, nil).Once()
		mockPostRepo.On("Delete", postID).Return(ErrPostNotFound).Once()

		err := postUseCase.DeletePost(postID)

		assert.Error(t, err)
		assert.Equal(t, ErrPostNotFound, err)
		mockPostRepo.AssertExpectations(t)
		mockSubPostRepo.AssertExpectations(t)
	})

	t.Run("Has SubPosts", func(t *testing.T) {
		subPosts := []domain.SubPost{
			{
				ID:      primitive.NewObjectID(),
				Content: "Sub post 1",
				Order:   1,
			},
		}
		mockSubPostRepo.On("FindByParentID", postID, 0, 0).Return(subPosts, nil).Once()

		err := postUseCase.DeletePost(postID)

		assert.Error(t, err)
		assert.Equal(t, ErrHasSubPosts, err)
		mockPostRepo.AssertExpectations(t)
		mockSubPostRepo.AssertExpectations(t)
	})
}

func TestPostUseCase_GetPost(t *testing.T) {
	mockPostRepo := new(mocks.PostRepository)
	mockSubPostRepo := new(mocks.SubPostRepository)
	postUseCase := NewPostUseCase(mockPostRepo, mockSubPostRepo)

	postID := primitive.NewObjectID()
	post := &domain.Post{
		ID:      postID,
		Content: "Test post",
		Media: []domain.Media{
			{
				Type: "image",
				URL:  "https://example.com/image.jpg",
			},
		},
		Tags: []string{"test"},
		Location: &domain.Location{
			Coordinates: []float64{100.5018, 13.7563},
			Type:        "Point",
			PlaceName:   "Bangkok",
		},
		Visibility: "public",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	t.Run("Success - Without SubPosts", func(t *testing.T) {
		mockPostRepo.On("FindByID", postID).Return(post, nil).Once()

		result, err := postUseCase.GetPost(postID, false)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, post.ID, result.Post.ID)
		assert.Equal(t, post.Content, result.Post.Content)
		assert.Equal(t, post.Media, result.Post.Media)
		assert.Equal(t, post.Tags, result.Post.Tags)
		assert.Equal(t, post.Location, result.Post.Location)
		assert.Equal(t, post.Visibility, result.Post.Visibility)
		assert.Empty(t, result.SubPosts)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("Success - With SubPosts", func(t *testing.T) {
		subPosts := []domain.SubPost{
			{
				ID:      primitive.NewObjectID(),
				Content: "Sub post 1",
				Order:   1,
			},
			{
				ID:      primitive.NewObjectID(),
				Content: "Sub post 2",
				Order:   2,
			},
		}
		mockPostRepo.On("FindByID", postID).Return(post, nil).Once()
		mockSubPostRepo.On("FindByPostID", postID).Return(subPosts, nil).Once()

		result, err := postUseCase.GetPost(postID, true)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, post.ID, result.Post.ID)
		assert.Equal(t, len(subPosts), len(result.SubPosts))
		for i, subPost := range result.SubPosts {
			assert.Equal(t, subPosts[i].ID, subPost.ID)
			assert.Equal(t, subPosts[i].Content, subPost.Content)
			assert.Equal(t, subPosts[i].Order, subPost.Order)
		}
		mockPostRepo.AssertExpectations(t)
		mockSubPostRepo.AssertExpectations(t)
	})

	t.Run("Post Not Found", func(t *testing.T) {
		mockPostRepo.On("FindByID", postID).Return(nil, ErrPostNotFound).Once()

		result, err := postUseCase.GetPost(postID, false)

		assert.Error(t, err)
		assert.Equal(t, ErrPostNotFound, err)
		assert.Nil(t, result)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("SubPosts Error", func(t *testing.T) {
		mockPostRepo.On("FindByID", postID).Return(post, nil).Once()
		mockSubPostRepo.On("FindByPostID", postID).Return(nil, assert.AnError).Once()

		result, err := postUseCase.GetPost(postID, true)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockPostRepo.AssertExpectations(t)
		mockSubPostRepo.AssertExpectations(t)
	})
}

func TestPostUseCase_ListPosts(t *testing.T) {
	mockPostRepo := new(mocks.PostRepository)
	mockSubPostRepo := new(mocks.SubPostRepository)
	postUseCase := NewPostUseCase(mockPostRepo, mockSubPostRepo)

	userID := primitive.NewObjectID()
	limit := 10
	offset := 0
	posts := []domain.Post{
		{
			ID:      primitive.NewObjectID(),
			Content: "Post 1",
		},
		{
			ID:      primitive.NewObjectID(),
			Content: "Post 2",
		},
	}

	t.Run("Success - Without SubPosts", func(t *testing.T) {
		mockPostRepo.On("FindByUserID", userID, limit, offset).Return(posts, nil).Once()

		results, err := postUseCase.ListPosts(userID, limit, offset, false)

		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.Equal(t, len(posts), len(results))
		for i, result := range results {
			assert.Equal(t, posts[i].ID, result.Post.ID)
			assert.Equal(t, posts[i].Content, result.Post.Content)
			assert.Empty(t, result.SubPosts)
		}
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("Success - With SubPosts", func(t *testing.T) {
		subPosts := []domain.SubPost{
			{
				ID:      primitive.NewObjectID(),
				Content: "Sub post 1",
				Order:   1,
			},
		}
		mockPostRepo.On("FindByUserID", userID, limit, offset).Return(posts, nil).Once()
		mockSubPostRepo.On("FindByPostID", posts[0].ID).Return(subPosts, nil).Once()
		mockSubPostRepo.On("FindByPostID", posts[1].ID).Return([]domain.SubPost{}, nil).Once()

		results, err := postUseCase.ListPosts(userID, limit, offset, true)

		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.Equal(t, len(posts), len(results))
		assert.Equal(t, len(subPosts), len(results[0].SubPosts))
		assert.Empty(t, results[1].SubPosts)
		mockPostRepo.AssertExpectations(t)
		mockSubPostRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockPostRepo.On("FindByUserID", userID, limit, offset).Return(nil, assert.AnError).Once()

		results, err := postUseCase.ListPosts(userID, limit, offset, false)

		assert.Error(t, err)
		assert.Nil(t, results)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("SubPosts Error", func(t *testing.T) {
		mockPostRepo.On("FindByUserID", userID, limit, offset).Return(posts, nil).Once()
		mockSubPostRepo.On("FindByPostID", posts[0].ID).Return(nil, assert.AnError).Once()

		results, err := postUseCase.ListPosts(userID, limit, offset, true)

		assert.Error(t, err)
		assert.Nil(t, results)
		mockPostRepo.AssertExpectations(t)
		mockSubPostRepo.AssertExpectations(t)
	})
}
