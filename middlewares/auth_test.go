package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/droskat/BH-Backend/config"
	"github.com/stretchr/testify/assert"
)

func newTestAuthEngine() *AuthEngine {
	cfg := config.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessExpiry:  15 * 60 * 1e9,
		RefreshExpiry: 168 * 60 * 60 * 1e9,
		Issuer:        "test-platform",
	}
	return NewAuthEngine(cfg)
}

func TestGenerateTokenPair_Success(t *testing.T) {
	engine := newTestAuthEngine()
	userID := uuid.New()

	pair, err := engine.GenerateTokenPair(userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.Greater(t, pair.ExpiresIn, int64(0))
}

func TestValidateAccessToken_Success(t *testing.T) {
	engine := newTestAuthEngine()
	userID := uuid.New()

	pair, err := engine.GenerateTokenPair(userID)
	assert.NoError(t, err)

	claims, err := engine.ValidateAccessToken(pair.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, userID.String(), claims.UserID)
	assert.Equal(t, AccessToken, claims.TokenType)
}

func TestValidateRefreshToken_Success(t *testing.T) {
	engine := newTestAuthEngine()
	userID := uuid.New()

	pair, err := engine.GenerateTokenPair(userID)
	assert.NoError(t, err)

	claims, err := engine.ValidateRefreshToken(pair.RefreshToken)
	assert.NoError(t, err)
	assert.Equal(t, userID.String(), claims.UserID)
	assert.Equal(t, RefreshToken, claims.TokenType)
}

func TestValidateAccessToken_WrongType(t *testing.T) {
	engine := newTestAuthEngine()
	userID := uuid.New()

	pair, err := engine.GenerateTokenPair(userID)
	assert.NoError(t, err)

	_, err = engine.ValidateAccessToken(pair.RefreshToken)
	assert.Error(t, err)
}

func TestValidateAccessToken_InvalidToken(t *testing.T) {
	engine := newTestAuthEngine()

	_, err := engine.ValidateAccessToken("invalid.token.here")
	assert.Error(t, err)
}

func TestJWTAuthMiddleware_MissingHeader(t *testing.T) {
	engine := newTestAuthEngine()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(engine.JWTAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
	engine := newTestAuthEngine()
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	pair, _ := engine.GenerateTokenPair(userID)

	router := gin.New()
	router.Use(engine.JWTAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		uid := c.MustGet("user_id").(uuid.UUID)
		c.JSON(200, gin.H{"user_id": uid.String()})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTAuthMiddleware_ExpiredToken(t *testing.T) {
	engine := newTestAuthEngine()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(engine.JWTAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzIiwiZXhwIjoxfQ.invalid")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
