package user

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) SignUp(ctx *gin.Context, req *CreateUserRequest) (*UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserResponse), args.Error(1)
}

func (m *MockUserService) Login(ctx *gin.Context, req *LoginRequest) (*TokenResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TokenResponse), args.Error(1)
}

func TestUserController(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("SignUp_Success", func(t *testing.T) {
		mockSvc := new(MockUserService)
		controller := NewUserController(mockSvc)

		userReq := CreateUserRequest{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@test.com",
			Password:  "password123",
			Password2: "password123",
			City:      "Bogotá",
			Country:   "Colombia",
		}

		userResp := &UserResponse{
			ID:        1,
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@test.com",
			City:      "Bogotá",
			Country:   "Colombia",
		}

		mockSvc.On("SignUp", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*user.CreateUserRequest")).Return(userResp, nil)

		body, _ := json.Marshal(userReq)
		req := httptest.NewRequest("POST", "/auth/signup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := gin.New()
		router.POST("/auth/signup", controller.SignUp)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response UserResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "john@test.com", response.Email)
		assert.Equal(t, uint(1), response.ID)

		mockSvc.AssertExpectations(t)
	})

	t.Run("SignUp_InvalidJSON", func(t *testing.T) {
		mockSvc := new(MockUserService)
		controller := NewUserController(mockSvc)

		req := httptest.NewRequest("POST", "/auth/signup", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := gin.New()
		router.POST("/auth/signup", controller.SignUp)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response["error"], "Invalid or malformed input data")
	})

	t.Run("SignUp_PasswordMismatch", func(t *testing.T) {
		mockSvc := new(MockUserService)
		controller := NewUserController(mockSvc)

		userReq := CreateUserRequest{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@test.com",
			Password:  "password123",
			Password2: "different123",
			City:      "Bogotá",
			Country:   "Colombia",
		}

		body, _ := json.Marshal(userReq)
		req := httptest.NewRequest("POST", "/auth/signup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := gin.New()
		router.POST("/auth/signup", controller.SignUp)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response["error"], "Passwords do not match")
	})

	t.Run("Login_Success", func(t *testing.T) {
		mockSvc := new(MockUserService)
		controller := NewUserController(mockSvc)

		loginReq := LoginRequest{
			Email:    "john@test.com",
			Password: "password123",
		}

		tokenResp := &TokenResponse{
			AccessToken: "test.jwt.token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		}

		mockSvc.On("Login", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*user.LoginRequest")).Return(tokenResp, nil)

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := gin.New()
		router.POST("/auth/login", controller.Login)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response TokenResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "test.jwt.token", response.AccessToken)
		assert.Equal(t, "Bearer", response.TokenType)

		mockSvc.AssertExpectations(t)
	})

	t.Run("Login_InvalidCredentials", func(t *testing.T) {
		mockSvc := new(MockUserService)
		controller := NewUserController(mockSvc)

		loginReq := LoginRequest{
			Email:    "wrong@test.com",
			Password: "wrongpass",
		}

		mockSvc.On("Login", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*user.LoginRequest")).Return(nil, assert.AnError)

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := gin.New()
		router.POST("/auth/login", controller.Login)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid credentials.", response["error"])

		mockSvc.AssertExpectations(t)
	})
}
