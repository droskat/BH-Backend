package middlewares

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/droskat/BH-Backend/config"
	"github.com/droskat/BH-Backend/models"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	UserID    string    `json:"user_id"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

type AuthEngine struct {
	cfg config.JWTConfig
}

func NewAuthEngine(cfg config.JWTConfig) *AuthEngine {
	return &AuthEngine{cfg: cfg}
}

func (a *AuthEngine) GenerateTokenPair(userID uuid.UUID) (*models.TokenPairResponse, error) {
	now := time.Now()

	accessClaims := &Claims{
		UserID:    userID.String(),
		TokenType: AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.cfg.Issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(a.cfg.AccessExpiry)),
			ID:        uuid.New().String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString([]byte(a.cfg.AccessSecret))
	if err != nil {
		return nil, err
	}

	refreshClaims := &Claims{
		UserID:    userID.String(),
		TokenType: RefreshToken,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.cfg.Issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(a.cfg.RefreshExpiry)),
			ID:        uuid.New().String(),
		},
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshTokenObj.SignedString([]byte(a.cfg.RefreshSecret))
	if err != nil {
		return nil, err
	}

	return &models.TokenPairResponse{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		ExpiresIn:    int64(a.cfg.AccessExpiry.Seconds()),
	}, nil
}

func (a *AuthEngine) ValidateAccessToken(tokenStr string) (*Claims, error) {
	return a.parseToken(tokenStr, a.cfg.AccessSecret, AccessToken)
}

func (a *AuthEngine) ValidateRefreshToken(tokenStr string) (*Claims, error) {
	return a.parseToken(tokenStr, a.cfg.RefreshSecret, RefreshToken)
}

func (a *AuthEngine) parseToken(tokenStr, secret string, expectedType TokenType) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.TokenType != expectedType {
		return nil, errors.New("invalid token type")
	}
	return claims, nil
}

func (a *AuthEngine) JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "missing authorization header",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid authorization header format",
			})
			return
		}

		claims, err := a.ValidateAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "invalid or expired token",
				Details: err.Error(),
			})
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid user identity in token",
			})
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
