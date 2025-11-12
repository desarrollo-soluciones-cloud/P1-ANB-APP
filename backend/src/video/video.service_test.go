package video

import (
	"anb-app/src/queue"
	"context"
	"mime/multipart"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockVideoRepository struct {
	mock.Mock
}

func (m *MockVideoRepository) Create(video *Video) (*Video, error) {
	args := m.Called(video)
	return args.Get(0).(*Video), args.Error(1)
}

func (m *MockVideoRepository) FindByUserID(userID uint) ([]Video, error) {
	args := m.Called(userID)
	return args.Get(0).([]Video), args.Error(1)
}

func (m *MockVideoRepository) FindByID(videoID uint) (*Video, error) {
	args := m.Called(videoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Video), args.Error(1)
}

func (m *MockVideoRepository) Delete(videoID uint) error {
	args := m.Called(videoID)
	return args.Error(0)
}

func (m *MockVideoRepository) FindPublic() ([]Video, error) {
	args := m.Called()
	return args.Get(0).([]Video), args.Error(1)
}

func (m *MockVideoRepository) Update(video *Video) error {
	args := m.Called(video)
	return args.Error(0)
}

func (m *MockVideoRepository) GetRankings() ([]RankingResponse, error) {
	args := m.Called()
	return args.Get(0).([]RankingResponse), args.Error(1)
}

// Mock para StorageService
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) Upload(file multipart.File, s3Key string) error {
	args := m.Called(file, s3Key)
	return args.Error(0)
}

func (m *MockStorageService) Delete(s3Key string) error {
	args := m.Called(s3Key)
	return args.Error(0)
}

func (m *MockStorageService) GetPresignedURL(s3Key string, expiration time.Duration) (string, error) {
	args := m.Called(s3Key, expiration)
	return args.String(0), args.Error(1)
}

// Mock para QueueClient
type MockQueueClient struct {
	mock.Mock
}

func (m *MockQueueClient) EnqueueTask(ctx context.Context, taskType string, payload queue.TaskPayload, maxRetry int, timeout time.Duration) (string, error) {
	args := m.Called(ctx, taskType, payload, maxRetry, timeout)
	return args.String(0), args.Error(1)
}

func (m *MockQueueClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestVideoService(t *testing.T) {
	t.Run("ListByUserID_Success", func(t *testing.T) {
		mockRepo := new(MockVideoRepository)
		mockStorage := new(MockStorageService)
		mockQueue := new(MockQueueClient)
		videoSvc := NewVideoService(mockRepo, mockQueue, mockStorage)

		userID := uint(1)
		videos := []Video{
			{ID: 1, UserID: userID, Title: "Test Video", Status: "processed"},
		}

		mockRepo.On("FindByUserID", userID).Return(videos, nil)

		result, err := videoSvc.ListByUserID(userID)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "Test Video", result[0].Title)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetByID_Success", func(t *testing.T) {
		mockRepo := new(MockVideoRepository)
		mockStorage := new(MockStorageService)
		mockQueue := new(MockQueueClient)
		videoSvc := NewVideoService(mockRepo, mockQueue, mockStorage)

		videoID := uint(1)
		userID := uint(1)
		video := &Video{
			ID:     videoID,
			UserID: userID,
			Title:  "Test Video",
			Status: "processed",
		}

		mockRepo.On("FindByID", videoID).Return(video, nil)

		result, err := videoSvc.GetByID(videoID, userID)

		assert.NoError(t, err)
		assert.Equal(t, videoID, result.ID)
		assert.Equal(t, "Test Video", result.Title)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		mockRepo := new(MockVideoRepository)
		mockStorage := new(MockStorageService)
		mockQueue := new(MockQueueClient)
		videoSvc := NewVideoService(mockRepo, mockQueue, mockStorage)

		videoID := uint(999)
		userID := uint(1)

		mockRepo.On("FindByID", videoID).Return(nil, nil)

		result, err := videoSvc.GetByID(videoID, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetByID_PermissionDenied", func(t *testing.T) {
		mockRepo := new(MockVideoRepository)
		mockStorage := new(MockStorageService)
		mockQueue := new(MockQueueClient)
		videoSvc := NewVideoService(mockRepo, mockQueue, mockStorage)

		videoID := uint(1)
		userID := uint(1)
		ownerID := uint(2) // Usuario diferente

		video := &Video{
			ID:     videoID,
			UserID: ownerID, // El video pertenece a otro usuario
			Title:  "Other User Video",
		}

		mockRepo.On("FindByID", videoID).Return(video, nil)

		result, err := videoSvc.GetByID(videoID, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "permission")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Delete_Success", func(t *testing.T) {
		mockRepo := new(MockVideoRepository)
		mockStorage := new(MockStorageService)
		mockQueue := new(MockQueueClient)
		videoSvc := NewVideoService(mockRepo, mockQueue, mockStorage)

		videoID := uint(1)
		userID := uint(1)
		video := &Video{
			ID:          videoID,
			UserID:      userID,
			Status:      "uploaded",
			OriginalURL: "originals/test.mp4",
		}

		mockRepo.On("FindByID", videoID).Return(video, nil)
		mockStorage.On("Delete", "originals/test.mp4").Return(nil)
		mockRepo.On("Delete", videoID).Return(nil)

		err := videoSvc.Delete(videoID, userID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
	})

	t.Run("ListPublic_Success", func(t *testing.T) {
		mockRepo := new(MockVideoRepository)
		mockStorage := new(MockStorageService)
		mockQueue := new(MockQueueClient)
		videoSvc := NewVideoService(mockRepo, mockQueue, mockStorage)

		videos := []Video{
			{ID: 1, Title: "Public Video 1", Status: "processed", VoteCount: 10},
			{ID: 2, Title: "Public Video 2", Status: "processed", VoteCount: 5},
		}

		mockRepo.On("FindPublic").Return(videos, nil)

		result, err := videoSvc.ListPublic()

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("MarkAsProcessed_Success", func(t *testing.T) {
		mockRepo := new(MockVideoRepository)
		mockStorage := new(MockStorageService)
		mockQueue := new(MockQueueClient)
		videoSvc := NewVideoService(mockRepo, mockQueue, mockStorage)

		videoID := uint(1)
		userID := uint(1)
		video := &Video{
			ID:          videoID,
			UserID:      userID,
			Title:       "Test Video",
			Status:      "uploaded",
			OriginalURL: "originals/test-video.mov",
		}

		mockRepo.On("FindByID", videoID).Return(video, nil)
		mockStorage.On("GetPresignedURL", "originals/test-video.mov", time.Hour).Return("https://s3.amazonaws.com/presigned-url-original", nil)
		mockStorage.On("GetPresignedURL", "processed/test-video.mp4", time.Hour).Return("https://s3.amazonaws.com/presigned-url-processed", nil)
		mockRepo.On("Update", mock.AnythingOfType("*video.Video")).Return(nil)

		result, err := videoSvc.MarkAsProcessed(videoID, userID)

		assert.NoError(t, err)
		assert.Equal(t, "processed", result.Status)
		// La respuesta contiene URLs presignadas de S3, no las claves
		assert.Equal(t, "https://s3.amazonaws.com/presigned-url-processed", result.ProcessedURL)
		assert.Equal(t, "https://s3.amazonaws.com/presigned-url-original", result.OriginalURL)
		assert.NotNil(t, result.ProcessedAt)
		mockRepo.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
	})
}
