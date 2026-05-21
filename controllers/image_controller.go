package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/droskat/BH-Backend/models"
	"github.com/droskat/BH-Backend/services"
)

type ImageController struct {
	service *services.ImageService
}

func NewImageController(service *services.ImageService) *ImageController {
	return &ImageController{service: service}
}

func (ctrl *ImageController) BulkUpload(c *gin.Context) {
	var req models.BulkUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation failed",
			Details: err.Error(),
		})
		return
	}

	if len(req.Images) > 50 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "batch size exceeds maximum of 50 items",
		})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	resp, err := ctrl.service.BulkUpload(c.Request.Context(), userID, req.Images)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to process bulk upload",
		})
		return
	}

	c.JSON(http.StatusAccepted, resp)
}

func (ctrl *ImageController) GetImage(c *gin.Context) {
	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid image_id format",
		})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	resp, err := ctrl.service.GetImageByID(c.Request.Context(), imageID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (ctrl *ImageController) ListImages(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	images, err := ctrl.service.GetUserImages(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to retrieve images",
		})
		return
	}

	c.JSON(http.StatusOK, images)
}

func (ctrl *ImageController) GetDownloadURL(c *gin.Context) {
	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid image_id format",
		})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	resp, err := ctrl.service.GetDownloadURL(c.Request.Context(), imageID, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (ctrl *ImageController) UpdateImage(c *gin.Context) {
	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid image_id format",
		})
		return
	}

	var req models.UpdateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation failed",
			Details: err.Error(),
		})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	if err := ctrl.service.UpdateImage(c.Request.Context(), imageID, userID, req); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (ctrl *ImageController) DeleteImage(c *gin.Context) {
	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid image_id format",
		})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	if err := ctrl.service.DeleteImage(c.Request.Context(), imageID, userID); err != nil {
		handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrNotFound):
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "resource not found"})
	case errors.Is(err, services.ErrForbidden):
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "access denied"})
	default:
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
	}
}
