package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthService interface {
	GenerateToken(userID uint, expirationTime time.Time) (string, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
	AuthMiddleware() gin.HandlerFunc
}

type authService struct {
	jwtSecret []byte
}

func NewAuthService(secret string) AuthService {
	return &authService{
		jwtSecret: []byte(secret),
	}
}

func (s *authService) GenerateToken(userID uint, expirationTime time.Time) (string, error) {
	claims := &jwt.RegisteredClaims{
		Subject:   fmt.Sprint(userID),
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	return tokenString, err
}

func (s *authService) ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}
