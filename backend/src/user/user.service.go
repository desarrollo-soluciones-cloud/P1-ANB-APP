package user

import (
	"errors"
	"project-one/src/auth"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	FindByEmail(email string) (*User, error)
	Create(user *User) (*User, error)
}

type userService struct {
	userRepo    UserRepository
	authService auth.AuthService
}

func NewUserService(userRepo UserRepository, authService auth.AuthService) UserService {
	return &userService{
		userRepo:    userRepo,
		authService: authService,
	}
}

func (s *userService) Register(ctx *gin.Context, req *CreateUserRequest) (*UserResponse, error) {
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("database error while checking email")
	}
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("could not hash password")
	}

	newUser := &User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  string(hashedPassword),
		City:      req.City,
		Country:   req.Country,
	}

	createdUser, err := s.userRepo.Create(newUser)
	if err != nil {
		return nil, errors.New("could not create user in database")
	}

	response := &UserResponse{
		ID:        createdUser.ID,
		FirstName: createdUser.FirstName,
		LastName:  createdUser.LastName,
		Email:     createdUser.Email,
		City:      createdUser.City,
		Country:   createdUser.Country,
		CreatedAt: createdUser.CreatedAt,
	}
	return response, nil
}

func (s *userService) Login(ctx *gin.Context, req *LoginRequest) (*TokenResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("database error")
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	tokenString, err := s.authService.GenerateToken(user.ID, expirationTime) // <-- LÍNEA CORREGIDA
	if err != nil {
		return nil, errors.New("could not generate token")
	}

	response := &TokenResponse{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   expirationTime.Unix(),
	}

	return response, nil
}
