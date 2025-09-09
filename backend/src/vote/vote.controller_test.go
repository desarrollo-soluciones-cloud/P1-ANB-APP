package vote

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockVoteService struct {
	mock.Mock
}

func (m *MockVoteService) CreateVote(userID, videoID uint) error {
	args := m.Called(userID, videoID)
	return args.Error(0)
}

func (m *MockVoteService) DeleteVote(userID, videoID uint) error {
	args := m.Called(userID, videoID)
	return args.Error(0)
}

func TestVoteController(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Create_Success", func(t *testing.T) {
		mockSvc := new(MockVoteService)
		controller := NewVoteController(mockSvc)

		userID := uint(1)
		videoID := uint(1)

		mockSvc.On("CreateVote", userID, videoID).Return(nil)

		req := httptest.NewRequest("POST", "/public/videos/1/vote", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", userID)
		c.Params = []gin.Param{{Key: "video_id", Value: "1"}}

		controller.Create(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Vote successfully registered.", response["message"])

		mockSvc.AssertExpectations(t)
	})

	t.Run("Create_Unauthorized", func(t *testing.T) {
		mockSvc := new(MockVoteService)
		controller := NewVoteController(mockSvc)

		req := httptest.NewRequest("POST", "/public/videos/1/vote", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		// No se establece userID
		c.Params = []gin.Param{{Key: "video_id", Value: "1"}}

		controller.Create(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Missing authentication", response["error"])
	})

	t.Run("Create_InvalidVideoID", func(t *testing.T) {
		mockSvc := new(MockVoteService)
		controller := NewVoteController(mockSvc)

		userID := uint(1)

		req := httptest.NewRequest("POST", "/public/videos/invalid/vote", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", userID)
		c.Params = []gin.Param{{Key: "video_id", Value: "invalid"}}

		controller.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid video ID format", response["error"])
	})

	t.Run("Create_AlreadyVoted", func(t *testing.T) {
		mockSvc := new(MockVoteService)
		controller := NewVoteController(mockSvc)

		userID := uint(1)
		videoID := uint(1)

		mockSvc.On("CreateVote", userID, videoID).Return(assert.AnError)

		req := httptest.NewRequest("POST", "/public/videos/1/vote", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", userID)
		c.Params = []gin.Param{{Key: "video_id", Value: "1"}}

		controller.Create(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockSvc.AssertExpectations(t)
	})

	t.Run("Delete_Success", func(t *testing.T) {
		mockSvc := new(MockVoteService)
		controller := NewVoteController(mockSvc)

		userID := uint(1)
		videoID := uint(1)

		mockSvc.On("DeleteVote", userID, videoID).Return(nil)

		req := httptest.NewRequest("DELETE", "/public/videos/1/vote", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", userID)
		c.Params = []gin.Param{{Key: "video_id", Value: "1"}}

		controller.Delete(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Voto successfully deleted.", response["message"])

		mockSvc.AssertExpectations(t)
	})

	t.Run("Delete_Unauthorized", func(t *testing.T) {
		mockSvc := new(MockVoteService)
		controller := NewVoteController(mockSvc)

		req := httptest.NewRequest("DELETE", "/public/videos/1/vote", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		// No se establece userID
		c.Params = []gin.Param{{Key: "video_id", Value: "1"}}

		controller.Delete(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Missing authentication.", response["error"])
	})

	t.Run("Delete_InvalidVideoID", func(t *testing.T) {
		mockSvc := new(MockVoteService)
		controller := NewVoteController(mockSvc)

		userID := uint(1)

		req := httptest.NewRequest("DELETE", "/public/videos/invalid/vote", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", userID)
		c.Params = []gin.Param{{Key: "video_id", Value: "invalid"}}

		controller.Delete(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid video ID format", response["error"])
	})

	t.Run("Delete_VoteNotExists", func(t *testing.T) {
		mockSvc := new(MockVoteService)
		controller := NewVoteController(mockSvc)

		userID := uint(1)
		videoID := uint(1)

		mockSvc.On("DeleteVote", userID, videoID).Return(assert.AnError)

		req := httptest.NewRequest("DELETE", "/public/videos/1/vote", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", userID)
		c.Params = []gin.Param{{Key: "video_id", Value: "1"}}

		controller.Delete(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockSvc.AssertExpectations(t)
	})
}
