package user

import (
	"anb-app/src/auth"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var ctx = context.Background()

type UserRepository interface {
	Create(user *User) (*User, error)
	FindByEmail(email string) (*User, error)
	GetRankings() ([]RankingResponse, error) // <-- NUEVO
}

type userService struct {
	userRepo    UserRepository
	authService auth.AuthService
	redisClient *redis.Client
}

func NewUserService(userRepo UserRepository, authService auth.AuthService, redisClient *redis.Client) UserService { // <-- CAMBIO AQUÍ
	return &userService{
		userRepo:    userRepo,
		authService: authService,
		redisClient: redisClient,
	}
}

func (s *userService) SignUp(ctx *gin.Context, req *CreateUserRequest) (*UserResponse, error) {
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

func (s *userService) GetRankings() ([]RankingResponse, error) {
	cacheKey := "rankings"

	cachedRankings, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var rankings []RankingResponse
		json.Unmarshal([]byte(cachedRankings), &rankings)
		return rankings, nil
	}

	if err != redis.Nil {
		return nil, err
	}

	rankings, err := s.userRepo.GetRankings()
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(rankings)
	if err != nil {
		return nil, err
	}
	s.redisClient.Set(ctx, cacheKey, jsonData, 2*time.Minute)

	return rankings, nil
}
