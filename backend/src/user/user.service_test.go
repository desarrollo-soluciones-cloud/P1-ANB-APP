package user

import (
	"anb-app/src/auth"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *User) (*User, error) {
	args := m.Called(user)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(email string) (*User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func TestUserService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("SignUp_Success", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		authSvc := auth.NewAuthService("test-secret")
		userSvc := NewUserService(mockRepo, authSvc)

		req := &CreateUserRequest{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@test.com",
			Password:  "password123",
			Password2: "password123",
			City:      "Bogotá",
			Country:   "Colombia",
		}

		mockRepo.On("FindByEmail", req.Email).Return(nil, nil)
		mockRepo.On("Create", mock.AnythingOfType("*user.User")).Return(&User{
			ID:        1,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
			City:      req.City,
			Country:   req.Country,
		}, nil)

		ctx := &gin.Context{}
		result, err := userSvc.SignUp(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, req.Email, result.Email)
		assert.Equal(t, uint(1), result.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("SignUp_EmailExists", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		authSvc := auth.NewAuthService("test-secret")
		userSvc := NewUserService(mockRepo, authSvc)

		req := &CreateUserRequest{
			Email: "existing@test.com",
		}

		existingUser := &User{ID: 1, Email: req.Email}
		mockRepo.On("FindByEmail", req.Email).Return(existingUser, nil)

		ctx := &gin.Context{}
		result, err := userSvc.SignUp(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "email already exists")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Login_Success", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		authSvc := auth.NewAuthService("test-secret")
		userSvc := NewUserService(mockRepo, authSvc)

		req := &LoginRequest{
			Email:    "john@test.com",
			Password: "password123",
		}

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		user := &User{
			ID:       1,
			Email:    req.Email,
			Password: string(hashedPassword),
		}

		mockRepo.On("FindByEmail", req.Email).Return(user, nil)

		ctx := &gin.Context{}
		result, err := userSvc.Login(ctx, req)

		assert.NoError(t, err)
		assert.NotEmpty(t, result.AccessToken)
		assert.Equal(t, "Bearer", result.TokenType)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Login_InvalidCredentials", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		authSvc := auth.NewAuthService("test-secret")
		userSvc := NewUserService(mockRepo, authSvc)

		req := &LoginRequest{
			Email:    "wrong@test.com",
			Password: "wrongpassword",
		}

		mockRepo.On("FindByEmail", req.Email).Return(nil, nil)

		ctx := &gin.Context{}
		result, err := userSvc.Login(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid credentials")
		mockRepo.AssertExpectations(t)
	})
}
