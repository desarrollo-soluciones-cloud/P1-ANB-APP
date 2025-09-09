package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuthService(t *testing.T) {
	secret := "test-secret-key"
	authSvc := NewAuthService(secret)

	t.Run("GenerateToken", func(t *testing.T) {
		userID := uint(1)
		expTime := time.Now().Add(time.Hour)

		token, err := authSvc.GenerateToken(userID, expTime)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("ValidateToken_Success", func(t *testing.T) {
		userID := uint(1)
		expTime := time.Now().Add(time.Hour)

		token, _ := authSvc.GenerateToken(userID, expTime)
		validToken, err := authSvc.ValidateToken(token)

		assert.NoError(t, err)
		assert.True(t, validToken.Valid)
	})

	t.Run("ValidateToken_Invalid", func(t *testing.T) {
		invalidToken := "invalid.token.here"

		_, err := authSvc.ValidateToken(invalidToken)

		assert.Error(t, err)
	})

	t.Run("ValidateToken_Expired", func(t *testing.T) {
		userID := uint(1)
		expTime := time.Now().Add(-time.Hour) // Token expirado

		token, _ := authSvc.GenerateToken(userID, expTime)
		_, err := authSvc.ValidateToken(token)

		assert.Error(t, err)
	})
}
