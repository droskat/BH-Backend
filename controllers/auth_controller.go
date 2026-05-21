package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/droskat/BH-Backend/middlewares"
	"github.com/droskat/BH-Backend/models"
)

type AuthController struct {
	authEngine *middlewares.AuthEngine
}

func NewAuthController(authEngine *middlewares.AuthEngine) *AuthController {
	return &AuthController{authEngine: authEngine}
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation failed",
			Details: err.Error(),
		})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid user_id format",
		})
		return
	}

	tokenPair, err := ctrl.authEngine.GenerateTokenPair(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to generate tokens",
		})
		return
	}

	c.JSON(http.StatusOK, tokenPair)
}

func (ctrl *AuthController) Refresh(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation failed",
			Details: err.Error(),
		})
		return
	}

	claims, err := ctrl.authEngine.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "invalid refresh token",
			Details: err.Error(),
		})
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "invalid user identity in token",
		})
		return
	}

	tokenPair, err := ctrl.authEngine.GenerateTokenPair(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to generate tokens",
		})
		return
	}

	c.JSON(http.StatusOK, tokenPair)
}
