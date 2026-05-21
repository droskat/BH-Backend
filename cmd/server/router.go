package main

import (
	"github.com/gin-gonic/gin"
	"github.com/droskat/BH-Backend/controllers"
	"github.com/droskat/BH-Backend/middlewares"
)

func setupRouter(
	authEngine *middlewares.AuthEngine,
	imageCtrl *controllers.ImageController,
	authCtrl *controllers.AuthController,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(middlewares.RecoveryMiddleware())
	router.Use(middlewares.RateLimitMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/login", authCtrl.Login)
		auth.POST("/refresh", authCtrl.Refresh)
	}

	api := router.Group("/api/v1")
	api.Use(authEngine.JWTAuthMiddleware())
	{
		api.POST("/images/bulk", imageCtrl.BulkUpload)
		api.GET("/images", imageCtrl.ListImages)
		api.GET("/images/:id", imageCtrl.GetImage)
		api.GET("/images/:id/download", imageCtrl.GetDownloadURL)
		api.PUT("/images/:id", imageCtrl.UpdateImage)
		api.DELETE("/images/:id", imageCtrl.DeleteImage)
	}

	return router
}
