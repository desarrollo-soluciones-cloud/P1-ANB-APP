package video

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockVideoService struct {
	mock.Mock
}

func (m *MockVideoService) Upload(ctx *gin.Context, req *UploadVideoRequest, file *multipart.FileHeader, userID uint) (*VideoResponse, error) {
	args := m.Called(ctx, req, file, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*VideoResponse), args.Error(1)
}

func (m *MockVideoService) ListByUserID(userID uint) ([]VideoResponse, error) {
	args := m.Called(userID)
	return args.Get(0).([]VideoResponse), args.Error(1)
}

func (m *MockVideoService) GetByID(videoID uint, userID uint) (*VideoResponse, error) {
	args := m.Called(videoID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*VideoResponse), args.Error(1)
}

func (m *MockVideoService) Delete(videoID uint, userID uint) error {
	args := m.Called(videoID, userID)
	return args.Error(0)
}

func (m *MockVideoService) ListPublic() ([]VideoResponse, error) {
	args := m.Called()
	return args.Get(0).([]VideoResponse), args.Error(1)
}

func (m *MockVideoService) MarkAsProcessed(videoID uint, userID uint) (*VideoResponse, error) {
	args := m.Called(videoID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*VideoResponse), args.Error(1)
}

func (m *MockVideoService) GetRankings() ([]RankingResponse, error) {
	args := m.Called()
	return args.Get(0).([]RankingResponse), args.Error(1)
}

// Helper para crear contexto con userID
func createContextWithUser(userID uint) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", userID)
	return c
}

func TestVideoController(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("ListMyVideos_Success", func(t *testing.T) {
		mockSvc := new(MockVideoService)
		controller := NewVideoController(mockSvc)

		userID := uint(1)
		videos := []VideoResponse{
			{ID: 1, UserID: userID, Title: "Test Video", Status: "processed"},
			{ID: 2, UserID: userID, Title: "Test Video 2", Status: "uploaded"},
		}

		mockSvc.On("ListByUserID", userID).Return(videos, nil)

		req := httptest.NewRequest("GET", "/videos", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", userID)

		controller.ListMyVideos(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []VideoResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 2)
		assert.Equal(t, "Test Video", response[0].Title)

		mockSvc.AssertExpectations(t)
	})

	t.Run("ListMyVideos_Unauthorized", func(t *testing.T) {
		mockSvc := new(MockVideoService)
		controller := NewVideoController(mockSvc)

		req := httptest.NewRequest("GET", "/videos", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		// No se establece userID

		controller.ListMyVideos(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "User not authenticated", response["error"])
	})

	t.Run("GetVideoByID_Success", func(t *testing.T) {
		mockSvc := new(MockVideoService)
		controller := NewVideoController(mockSvc)

		videoID := uint(1)
		userID := uint(1)
		videoResp := &VideoResponse{
			ID:     videoID,
			UserID: userID,
			Title:  "Test Video",
			Status: "processed",
		}

		mockSvc.On("GetByID", videoID, userID).Return(videoResp, nil)

		req := httptest.NewRequest("GET", "/videos/1", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", userID)
		c.Params = []gin.Param{{Key: "video_id", Value: "1"}}

		controller.GetVideoByID(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response VideoResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, videoID, response.ID)
		assert.Equal(t, "Test Video", response.Title)

		mockSvc.AssertExpectations(t)
	})

	t.Run("GetVideoByID_NotFound", func(t *testing.T) {
		mockSvc := new(MockVideoService)
		controller := NewVideoController(mockSvc)

		videoID := uint(999)
		userID := uint(1)

		mockSvc.On("GetByID", videoID, userID).Return(nil, assert.AnError)

		req := httptest.NewRequest("GET", "/videos/999", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", userID)
		c.Params = []gin.Param{{Key: "video_id", Value: "999"}}

		controller.GetVideoByID(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockSvc.AssertExpectations(t)
	})

	t.Run("GetVideoByID_InvalidID", func(t *testing.T) {
		mockSvc := new(MockVideoService)
		controller := NewVideoController(mockSvc)

		req := httptest.NewRequest("GET", "/videos/invalid", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", uint(1))
		c.Params = []gin.Param{{Key: "video_id", Value: "invalid"}}

		controller.GetVideoByID(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid video ID format", response["error"])
	})

	t.Run("DeleteVideo_Success", func(t *testing.T) {
		mockSvc := new(MockVideoService)
		controller := NewVideoController(mockSvc)

		videoID := uint(1)
		userID := uint(1)

		mockSvc.On("Delete", videoID, userID).Return(nil)

		req := httptest.NewRequest("DELETE", "/videos/1", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", userID)
		c.Params = []gin.Param{{Key: "video_id", Value: "1"}}

		controller.DeleteVideo(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response["message"], "eliminado exitosamente")
		assert.Equal(t, float64(videoID), response["video_id"])

		mockSvc.AssertExpectations(t)
	})

	t.Run("ListPublicVideos_Success", func(t *testing.T) {
		mockSvc := new(MockVideoService)
		controller := NewVideoController(mockSvc)

		videos := []VideoResponse{
			{ID: 1, Title: "Public Video 1", Status: "processed", VoteCount: 10},
			{ID: 2, Title: "Public Video 2", Status: "processed", VoteCount: 5},
		}

		mockSvc.On("ListPublic").Return(videos, nil)

		req := httptest.NewRequest("GET", "/public/videos", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		controller.ListPublicVideos(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []VideoResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 2)
		assert.Equal(t, "Public Video 1", response[0].Title)

		mockSvc.AssertExpectations(t)
	})

	t.Run("MarkVideoAsProcessed_Success", func(t *testing.T) {
		mockSvc := new(MockVideoService)
		controller := NewVideoController(mockSvc)

		videoID := uint(1)
		userID := uint(1)
		now := time.Now()
		videoResp := &VideoResponse{
			ID:          videoID,
			UserID:      userID,
			Title:       "Test Video",
			Status:      "processed",
			ProcessedAt: &now,
		}

		mockSvc.On("MarkAsProcessed", videoID, userID).Return(videoResp, nil)

		req := httptest.NewRequest("POST", "/videos/1/mark-processed", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", userID)
		c.Params = []gin.Param{{Key: "video_id", Value: "1"}}

		controller.MarkVideoAsProcessed(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response VideoResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "processed", response.Status)
		assert.NotNil(t, response.ProcessedAt)

		mockSvc.AssertExpectations(t)
	})

	t.Run("GetRankings_Success", func(t *testing.T) {
		mockSvc := new(MockVideoService)
		controller := NewVideoController(mockSvc)

		rankings := []RankingResponse{
			{Position: 1, VideoID: 1, Title: "Top Video", VoteCount: 100},
			{Position: 2, VideoID: 2, Title: "Second Video", VoteCount: 50},
		}

		mockSvc.On("GetRankings").Return(rankings, nil)

		req := httptest.NewRequest("GET", "/public/rankings", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		controller.GetRankings(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []RankingResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 2)
		assert.Equal(t, 1, response[0].Position)
		assert.Equal(t, "Top Video", response[0].Title)

		mockSvc.AssertExpectations(t)
	})
}
