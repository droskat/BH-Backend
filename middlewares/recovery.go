package middlewares

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/droskat/BH-Backend/models"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC RECOVERY] %v", r)
				c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
					Error: "internal server error",
				})
			}
		}()
		c.Next()
	}
}
