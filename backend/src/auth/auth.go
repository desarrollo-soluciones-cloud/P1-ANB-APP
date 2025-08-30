package auth

import (
	// "project-one/src/user" // <-- ELIMINAMOS ESTA LÍNEA
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthService ahora espera un uint (el ID del usuario)
type AuthService interface {
	GenerateToken(userID uint, expirationTime time.Time) (string, error)
}

type authService struct {
	jwtSecret []byte
}

func NewAuthService(secret string) AuthService {
	return &authService{
		jwtSecret: []byte(secret),
	}
}

// GenerateToken ahora recibe el userID directamente.
func (s *authService) GenerateToken(userID uint, expirationTime time.Time) (string, error) {
	claims := &jwt.RegisteredClaims{
		Subject:   string(rune(userID)), // Usamos el ID directamente
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)

	return tokenString, err
}
