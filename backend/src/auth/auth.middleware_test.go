package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret-key"
	authSvc := NewAuthService(secret)

	t.Run("Success_ValidToken", func(t *testing.T) {
		// Generar token válido
		userID := uint(123)
		expTime := time.Now().Add(time.Hour)
		token, _ := authSvc.GenerateToken(userID, expTime)

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		// Crear router con middleware
		router := gin.New()
		router.Use(authSvc.AuthMiddleware())
		router.GET("/protected", func(c *gin.Context) {
			userIDFromContext, _ := c.Get("userID")
			c.JSON(200, gin.H{"userID": userIDFromContext})
		})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "123")
	})

	t.Run("Fail_NoAuthHeader", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		// Sin header Authorization
		w := httptest.NewRecorder()

		router := gin.New()
		router.Use(authSvc.AuthMiddleware())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"success": true})
		})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization header is required")
	})

	t.Run("Fail_InvalidFormat", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat token")
		w := httptest.NewRecorder()

		router := gin.New()
		router.Use(authSvc.AuthMiddleware())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"success": true})
		})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization header format must be Bearer")
	})

	t.Run("Fail_InvalidToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid.jwt.token")
		w := httptest.NewRecorder()

		router := gin.New()
		router.Use(authSvc.AuthMiddleware())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"success": true})
		})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid token")
	})

	t.Run("Fail_ExpiredToken", func(t *testing.T) {
		// Generar token expirado
		userID := uint(123)
		expTime := time.Now().Add(-time.Hour) // Expirado hace 1 hora
		expiredToken, _ := authSvc.GenerateToken(userID, expTime)

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		w := httptest.NewRecorder()

		router := gin.New()
		router.Use(authSvc.AuthMiddleware())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"success": true})
		})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid token")
	})

	t.Run("Fail_OnlyBearerNoToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer")
		w := httptest.NewRecorder()

		router := gin.New()
		router.Use(authSvc.AuthMiddleware())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"success": true})
		})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization header format must be Bearer")
	})
}
